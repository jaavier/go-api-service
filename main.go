package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/config"
	"github.com/jaavier/go-api-service/internal/handler"
	"github.com/jaavier/go-api-service/internal/store"
	"github.com/jaavier/go-api-service/internal/worker"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Use a startup timeout so a misconfigured DB fails fast.
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	db, err := store.NewDB(dbCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	userStore := store.NewUserStore(db)
	userHandler := handler.NewUserHandler(userStore)

	// Background worker pool sized by cfg.MaxWorkers. Each worker consumes
	// jobs from a shared channel and exits cleanly when the channel is closed
	// or workerCtx is cancelled (see graceful shutdown below).
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	jobs := make(chan string, cfg.MaxWorkers)

	// Track every worker goroutine so shutdown can wait for them to drain
	// before the deferred db.Close() runs — otherwise a worker could touch a
	// DB handle that has already been closed.
	var workers sync.WaitGroup
	for i := 0; i < cfg.MaxWorkers; i++ {
		wg := worker.Run(workerCtx, jobs, func(msg string) {
			log.Printf("worker: processed job %q", msg)
		})
		workers.Add(1)
		go func() {
			defer workers.Done()
			wg.Wait()
		}()
	}

	// Guard against double-close of jobs: main is currently the only producer,
	// but closing exactly once keeps the shutdown path safe if that changes.
	var closeJobs sync.Once
	stopWorkers := func() {
		closeJobs.Do(func() { close(jobs) })
		workerCancel()
		workers.Wait() // block until every worker goroutine has returned
	}

	r := mux.NewRouter()
	r.HandleFunc("/users", userHandler.List).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", userHandler.Get).Methods(http.MethodGet)
	r.HandleFunc("/users", userHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", userHandler.Delete).Methods(http.MethodDelete)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s (workers=%d)", srv.Addr, cfg.MaxWorkers)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	case sig := <-quit:
		log.Printf("received signal %s, shutting down", sig)

		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()

		// Stop accepting HTTP work first, then drain the worker pool, and only
		// after both have finished let the deferred db.Close() run.
		if err := srv.Shutdown(shutCtx); err != nil {
			// Still drain workers before returning so db.Close() is safe.
			stopWorkers()
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		stopWorkers()
		log.Println("server stopped cleanly")
	}

	return nil
}
