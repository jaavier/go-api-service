package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/handler"
	"github.com/jaavier/go-api-service/internal/store"
)

func main() {
	db, err := store.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	userStore := store.NewUserStore(db)
	userHandler := handler.NewUserHandler(userStore)

	r := mux.NewRouter()
	r.HandleFunc("/users", userHandler.List).Methods("GET")
	r.HandleFunc("/users/{id}", userHandler.Get).Methods("GET")
	r.HandleFunc("/users", userHandler.Create).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
