package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	for i := 0; i < cfg.MaxWorkers; i++ {
		worker.Run(workerCtx, jobs, func(msg string) {
			log.Printf("worker: processed job %q", msg)
		})
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

		// Stop feeding the pool and signal the workers to exit.
		close(jobs)
		workerCancel()

		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		log.Println("server stopped cleanly")
	}

	return nil
}
