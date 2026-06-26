package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"
	"ticket-system/internal/store"

	"github.com/gorilla/mux"
)

func main() {
	s := store.New()

	authH := handlers.NewAuthHandler(s)
	ticketH := handlers.NewTicketHandler(s)

	r := mux.NewRouter()

	// Health
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)

	// Auth routes (public)
	r.HandleFunc("/auth/register", authH.Register).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", authH.Login).Methods(http.MethodPost)

	// Ticket routes (protected)
	api := r.PathPrefix("").Subrouter()
	api.Use(middleware.Authenticate)
	api.HandleFunc("/tickets", ticketH.Create).Methods(http.MethodPost)
	api.HandleFunc("/tickets", ticketH.List).Methods(http.MethodGet)
	api.HandleFunc("/tickets/{id}", ticketH.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tickets/{id}/status", ticketH.UpdateStatus).Methods(http.MethodPatch)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
