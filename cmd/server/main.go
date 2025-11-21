package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathannewyen/fhir-health-interop/internal/database"
	"github.com/nathannewyen/fhir-health-interop/internal/handlers"
	custommiddleware "github.com/nathannewyen/fhir-health-interop/internal/middleware"
	"github.com/nathannewyen/fhir-health-interop/internal/repository"
	"github.com/nathannewyen/fhir-health-interop/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure zerolog for console output with human-readable format
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Initialize database connection
	dbConfig := database.PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, dbError := database.NewPostgresConnection(dbConfig)
	if dbError != nil {
		log.Fatal().Err(dbError).Msg("Failed to connect to database")
	}
	defer databaseConnection.Close()

	log.Info().Msg("Database connection established")

	// Initialize repository and service layers
	patientRepository := repository.NewPostgresPatientRepository(databaseConnection)
	patientService := service.NewPatientService(patientRepository)

	// Create a new Chi router instance
	router := chi.NewRouter()

	// Add custom logging middleware and panic recovery
	router.Use(custommiddleware.Logger(log.Logger))
	router.Use(middleware.Recoverer)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	patientHandler := handlers.NewPatientHandlerWithService(patientService)
	samplePatientHandler := handlers.NewPatientHandler()

	// Register health check endpoint
	router.Get("/health", healthHandler.Check)

	// Register FHIR Patient endpoints
	router.Get("/fhir/Patient/sample", samplePatientHandler.GetSamplePatient)
	router.Post("/fhir/Patient", patientHandler.Create)
	router.Get("/fhir/Patient/{id}", patientHandler.GetByID)
	router.Get("/fhir/Patient", patientHandler.GetAll)

	// Define server port
	serverPort := ":8080"

	// Log server startup
	log.Info().Str("port", serverPort).Msg("FHIR Health Interop server starting")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health              - Health check")
	fmt.Println("  GET  /fhir/Patient/sample - Sample patient (hardcoded)")
	fmt.Println("  POST /fhir/Patient        - Create patient")
	fmt.Println("  GET  /fhir/Patient/{id}   - Get patient by ID")
	fmt.Println("  GET  /fhir/Patient        - Get all patients")
	fmt.Println()

	serverError := http.ListenAndServe(serverPort, router)
	if serverError != nil {
		log.Fatal().Err(serverError).Msg("Failed to start server")
	}
}
