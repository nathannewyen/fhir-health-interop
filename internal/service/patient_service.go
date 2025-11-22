package service

import (
	"context"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"github.com/nathannewyen/fhir-health-interop/internal/repository"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// PatientService handles business logic for Patient operations
type PatientService struct {
	patientRepository repository.PatientRepository
	patientMapper     *models.PatientMapper
}

// NewPatientService creates a new instance of PatientService
func NewPatientService(patientRepository repository.PatientRepository) *PatientService {
	return &PatientService{
		patientRepository: patientRepository,
		patientMapper:     models.NewPatientMapper(),
	}
}

// CreatePatient creates a new patient from FHIR Patient resource
func (service *PatientService) CreatePatient(ctx context.Context, fhirPatient *fhir.Patient) (*fhir.Patient, error) {
	// Convert FHIR Patient to domain model
	domainPatient := service.patientMapper.FromFHIR(fhirPatient)

	// Set default active status if not provided
	if fhirPatient.Active == nil {
		domainPatient.Active = true
	}

	// Save to database
	createdPatient, createError := service.patientRepository.Create(ctx, domainPatient)
	if createError != nil {
		return nil, createError
	}

	// Convert back to FHIR format and return
	return service.patientMapper.ToFHIR(createdPatient), nil
}

// GetPatientByID retrieves a patient by ID and returns as FHIR Patient
func (service *PatientService) GetPatientByID(ctx context.Context, patientID string) (*fhir.Patient, error) {
	// Get from database
	domainPatient, getError := service.patientRepository.GetByID(ctx, patientID)
	if getError != nil {
		return nil, getError
	}

	// Convert to FHIR format and return
	return service.patientMapper.ToFHIR(domainPatient), nil
}

// GetAllPatients retrieves all patients with pagination
func (service *PatientService) GetAllPatients(ctx context.Context, limit int, offset int) ([]*fhir.Patient, error) {
	// Get from database
	domainPatients, getError := service.patientRepository.GetAll(ctx, limit, offset)
	if getError != nil {
		return nil, getError
	}

	// Convert each patient to FHIR format
	fhirPatients := make([]*fhir.Patient, len(domainPatients))
	for index, domainPatient := range domainPatients {
		fhirPatients[index] = service.patientMapper.ToFHIR(domainPatient)
	}

	return fhirPatients, nil
}

// UpdatePatient updates an existing patient
func (service *PatientService) UpdatePatient(ctx context.Context, patientID string, fhirPatient *fhir.Patient) (*fhir.Patient, error) {
	// Convert FHIR Patient to domain model
	domainPatient := service.patientMapper.FromFHIR(fhirPatient)
	domainPatient.ID = patientID

	// Update in database
	updatedPatient, updateError := service.patientRepository.Update(ctx, domainPatient)
	if updateError != nil {
		return nil, updateError
	}

	// Convert back to FHIR format and return
	return service.patientMapper.ToFHIR(updatedPatient), nil
}

// SearchPatients retrieves patients matching the search criteria
func (service *PatientService) SearchPatients(ctx context.Context, searchParams *models.PatientSearchParams) ([]*fhir.Patient, error) {
	// Search in database
	domainPatients, searchError := service.patientRepository.Search(ctx, searchParams)
	if searchError != nil {
		return nil, searchError
	}

	// Convert each patient to FHIR format
	fhirPatients := make([]*fhir.Patient, len(domainPatients))
	for index, domainPatient := range domainPatients {
		fhirPatients[index] = service.patientMapper.ToFHIR(domainPatient)
	}

	return fhirPatients, nil
}

// DeletePatient removes a patient by ID
func (service *PatientService) DeletePatient(ctx context.Context, patientID string) error {
	return service.patientRepository.Delete(ctx, patientID)
}
