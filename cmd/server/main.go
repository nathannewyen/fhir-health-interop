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

	log.Info().Msg("PostgreSQL connection established")

	// Initialize MongoDB connection
	mongoConfig := database.MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, mongoError := database.NewMongoConnection(mongoConfig)
	if mongoError != nil {
		log.Fatal().Err(mongoError).Msg("Failed to connect to MongoDB")
	}

	log.Info().Msg("MongoDB connection established")

	// Initialize repository and service layers
	patientRepository := repository.NewPostgresPatientRepository(databaseConnection)
	patientService := service.NewPatientService(patientRepository)

	observationRepository := repository.NewMongoObservationRepository(mongoDatabase)
	observationService := service.NewObservationService(observationRepository)

	// Create a new Chi router instance
	router := chi.NewRouter()

	// Add middleware in order: RequestID -> Logger -> ErrorHandler -> Recoverer -> Validator
	router.Use(custommiddleware.RequestID)
	router.Use(custommiddleware.Logger(log.Logger))
	router.Use(custommiddleware.ErrorHandler)
	router.Use(middleware.Recoverer)
	router.Use(custommiddleware.FHIRValidator)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	patientHandler := handlers.NewPatientHandlerWithService(patientService)
	samplePatientHandler := handlers.NewPatientHandler()
	observationHandler := handlers.NewObservationHandler(observationService)

	// Register health check endpoint
	router.Get("/health", healthHandler.Check)

	// Register FHIR Patient endpoints
	router.Get("/fhir/Patient/sample", samplePatientHandler.GetSamplePatient)
	router.Post("/fhir/Patient", patientHandler.Create)
	router.Get("/fhir/Patient/{id}", patientHandler.GetByID)
	router.Get("/fhir/Patient", patientHandler.GetAll)

	// Register FHIR Observation endpoints
	router.Post("/fhir/Observation", observationHandler.Create)
	router.Get("/fhir/Observation/{id}", observationHandler.GetByID)
	router.Get("/fhir/Observation", observationHandler.GetAll)

	// Define server port
	serverPort := ":8080"

	// Log server startup
	log.Info().Str("port", serverPort).Msg("FHIR Health Interop server starting")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health                     - Health check")
	fmt.Println("  GET  /fhir/Patient/sample        - Sample patient (hardcoded)")
	fmt.Println("  POST /fhir/Patient               - Create patient")
	fmt.Println("  GET  /fhir/Patient/{id}          - Get patient by ID")
	fmt.Println("  GET  /fhir/Patient               - Get all patients")
	fmt.Println("  POST /fhir/Observation           - Create observation")
	fmt.Println("  GET  /fhir/Observation/{id}      - Get observation by ID")
	fmt.Println("  GET  /fhir/Observation?patient=X - Get observations by patient ID")
	fmt.Println()

	serverError := http.ListenAndServe(serverPort, router)
	if serverError != nil {
		log.Fatal().Err(serverError).Msg("Failed to start server")
	}
}
