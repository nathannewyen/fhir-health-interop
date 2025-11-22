package service

import (
	"context"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"github.com/nathannewyen/fhir-health-interop/internal/repository"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// ObservationService handles business logic for Observation resources
type ObservationService struct {
	observationRepository repository.ObservationRepository
	observationMapper     *models.ObservationMapper
}

// NewObservationService creates a new observation service instance
func NewObservationService(observationRepository repository.ObservationRepository) *ObservationService {
	return &ObservationService{
		observationRepository: observationRepository,
		observationMapper:     models.NewObservationMapper(),
	}
}

// CreateObservation creates a new observation from FHIR resource
func (service *ObservationService) CreateObservation(ctx context.Context, fhirObservation *fhir.Observation) (*fhir.Observation, error) {
	// Convert FHIR to domain model
	observation := service.observationMapper.FromFHIR(fhirObservation)

	// Create in repository
	createdObservation, createError := service.observationRepository.Create(ctx, observation)
	if createError != nil {
		return nil, createError
	}

	// Convert back to FHIR
	return service.observationMapper.ToFHIR(createdObservation), nil
}

// GetObservationByID retrieves an observation by ID
func (service *ObservationService) GetObservationByID(ctx context.Context, observationID string) (*fhir.Observation, error) {
	// Get from repository
	observation, getError := service.observationRepository.GetByID(ctx, observationID)
	if getError != nil {
		return nil, getError
	}

	// Convert to FHIR
	return service.observationMapper.ToFHIR(observation), nil
}

// GetObservationsByPatientID retrieves all observations for a patient
func (service *ObservationService) GetObservationsByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*fhir.Observation, error) {
	// Get from repository
	observations, getError := service.observationRepository.GetByPatientID(ctx, patientID, limit, offset)
	if getError != nil {
		return nil, getError
	}

	// Convert to FHIR
	fhirObservations := make([]*fhir.Observation, 0, len(observations))
	for _, observation := range observations {
		fhirObservations = append(fhirObservations, service.observationMapper.ToFHIR(observation))
	}

	return fhirObservations, nil
}

// GetAllObservations retrieves all observations with pagination
func (service *ObservationService) GetAllObservations(ctx context.Context, limit int, offset int) ([]*fhir.Observation, error) {
	// Get from repository
	observations, getError := service.observationRepository.GetAll(ctx, limit, offset)
	if getError != nil {
		return nil, getError
	}

	// Convert to FHIR
	fhirObservations := make([]*fhir.Observation, 0, len(observations))
	for _, observation := range observations {
		fhirObservations = append(fhirObservations, service.observationMapper.ToFHIR(observation))
	}

	return fhirObservations, nil
}

// UpdateObservation updates an existing observation
func (service *ObservationService) UpdateObservation(ctx context.Context, observationID string, fhirObservation *fhir.Observation) (*fhir.Observation, error) {
	// Convert FHIR to domain model
	observation := service.observationMapper.FromFHIR(fhirObservation)
	observation.ID = observationID

	// Update in repository
	updatedObservation, updateError := service.observationRepository.Update(ctx, observation)
	if updateError != nil {
		return nil, updateError
	}

	// Convert back to FHIR
	return service.observationMapper.ToFHIR(updatedObservation), nil
}

// SearchObservations retrieves observations matching the search criteria
func (service *ObservationService) SearchObservations(ctx context.Context, searchParams *models.ObservationSearchParams) ([]*fhir.Observation, error) {
	// Search in repository
	observations, searchError := service.observationRepository.Search(ctx, searchParams)
	if searchError != nil {
		return nil, searchError
	}

	// Convert to FHIR
	fhirObservations := make([]*fhir.Observation, 0, len(observations))
	for _, observation := range observations {
		fhirObservations = append(fhirObservations, service.observationMapper.ToFHIR(observation))
	}

	return fhirObservations, nil
}

// DeleteObservation deletes an observation by ID
func (service *ObservationService) DeleteObservation(ctx context.Context, observationID string) error {
	return service.observationRepository.Delete(ctx, observationID)
}
