package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathannewyen/fhir-health-interop/internal/handlers"
)

func main() {
	// Create a new Chi router instance
	router := chi.NewRouter()

	// Add built-in middleware for logging and panic recovery
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Initialize health handler
	healthHandler := handlers.NewHealthHandler()

	// Register health check endpoint
	router.Get("/health", healthHandler.Check)

	// Define server port
	serverPort := ":8080"

	// Start the HTTP server
	fmt.Printf("FHIR Health Interop server starting on port %s\n", serverPort)
	serverError := http.ListenAndServe(serverPort, router)
	if serverError != nil {
		log.Fatalf("Failed to start server: %v", serverError)
	}
}
