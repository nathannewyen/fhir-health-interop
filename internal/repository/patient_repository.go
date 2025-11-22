package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
)

// PatientRepository defines the interface for patient data operations
// This interface allows for different implementations (PostgreSQL, mock, etc.)
type PatientRepository interface {
	// Create inserts a new patient record and returns the created patient with ID
	Create(ctx context.Context, patient *models.Patient) (*models.Patient, error)

	// GetByID retrieves a patient by their unique identifier
	GetByID(ctx context.Context, patientID string) (*models.Patient, error)

	// GetAll retrieves all patients with optional pagination
	GetAll(ctx context.Context, limit int, offset int) ([]*models.Patient, error)

	// Search retrieves patients matching the search criteria
	Search(ctx context.Context, searchParams *models.PatientSearchParams) ([]*models.Patient, error)

	// Update modifies an existing patient record
	Update(ctx context.Context, patient *models.Patient) (*models.Patient, error)

	// Delete removes a patient record by ID
	Delete(ctx context.Context, patientID string) error
}

// PostgresPatientRepository implements PatientRepository using PostgreSQL
type PostgresPatientRepository struct {
	// Database connection pool
	databaseConnection *sql.DB
}

// NewPostgresPatientRepository creates a new PostgreSQL patient repository instance
func NewPostgresPatientRepository(databaseConnection *sql.DB) *PostgresPatientRepository {
	return &PostgresPatientRepository{
		databaseConnection: databaseConnection,
	}
}

// Create inserts a new patient record into the database
func (repository *PostgresPatientRepository) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	// SQL query to insert a new patient and return the generated ID and timestamps
	insertQuery := `
		INSERT INTO patients (identifier_system, identifier_value, active, family_name, given_name, gender, birth_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	// Execute the insert query and scan the returned values
	scanError := repository.databaseConnection.QueryRowContext(
		ctx,
		insertQuery,
		patient.IdentifierSystem,
		patient.IdentifierValue,
		patient.Active,
		patient.FamilyName,
		patient.GivenName,
		patient.Gender,
		patient.BirthDate,
	).Scan(&patient.ID, &patient.CreatedAt, &patient.UpdatedAt)

	if scanError != nil {
		return nil, scanError
	}

	return patient, nil
}

// GetByID retrieves a patient by their unique identifier
func (repository *PostgresPatientRepository) GetByID(ctx context.Context, patientID string) (*models.Patient, error) {
	// SQL query to select a patient by ID
	selectQuery := `
		SELECT id, identifier_system, identifier_value, active, family_name, given_name, gender, birth_date, created_at, updated_at
		FROM patients
		WHERE id = $1
	`

	// Create a new patient instance to hold the result
	patient := &models.Patient{}

	// Execute the query and scan the result into the patient struct
	scanError := repository.databaseConnection.QueryRowContext(ctx, selectQuery, patientID).Scan(
		&patient.ID,
		&patient.IdentifierSystem,
		&patient.IdentifierValue,
		&patient.Active,
		&patient.FamilyName,
		&patient.GivenName,
		&patient.Gender,
		&patient.BirthDate,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if scanError != nil {
		return nil, scanError
	}

	return patient, nil
}

// GetAll retrieves all patients with pagination support
func (repository *PostgresPatientRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Patient, error) {
	// SQL query to select all patients with limit and offset for pagination
	selectAllQuery := `
		SELECT id, identifier_system, identifier_value, active, family_name, given_name, gender, birth_date, created_at, updated_at
		FROM patients
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// Execute the query
	rows, queryError := repository.databaseConnection.QueryContext(ctx, selectAllQuery, limit, offset)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	// Create a slice to hold the results
	patients := []*models.Patient{}

	// Iterate through the rows and scan each into a patient struct
	for rows.Next() {
		patient := &models.Patient{}
		scanError := rows.Scan(
			&patient.ID,
			&patient.IdentifierSystem,
			&patient.IdentifierValue,
			&patient.Active,
			&patient.FamilyName,
			&patient.GivenName,
			&patient.Gender,
			&patient.BirthDate,
			&patient.CreatedAt,
			&patient.UpdatedAt,
		)
		if scanError != nil {
			return nil, scanError
		}
		patients = append(patients, patient)
	}

	// Check for errors during iteration
	if rowsError := rows.Err(); rowsError != nil {
		return nil, rowsError
	}

	return patients, nil
}

// Update modifies an existing patient record in the database
func (repository *PostgresPatientRepository) Update(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	// SQL query to update a patient and return the updated timestamp
	updateQuery := `
		UPDATE patients
		SET identifier_system = $1, identifier_value = $2, active = $3, family_name = $4, given_name = $5, gender = $6, birth_date = $7, updated_at = $8
		WHERE id = $9
		RETURNING updated_at
	`

	// Set the updated timestamp
	patient.UpdatedAt = time.Now()

	// Execute the update query
	scanError := repository.databaseConnection.QueryRowContext(
		ctx,
		updateQuery,
		patient.IdentifierSystem,
		patient.IdentifierValue,
		patient.Active,
		patient.FamilyName,
		patient.GivenName,
		patient.Gender,
		patient.BirthDate,
		patient.UpdatedAt,
		patient.ID,
	).Scan(&patient.UpdatedAt)

	if scanError != nil {
		return nil, scanError
	}

	return patient, nil
}

// Search retrieves patients matching the search criteria with dynamic filtering
func (repository *PostgresPatientRepository) Search(ctx context.Context, searchParams *models.PatientSearchParams) ([]*models.Patient, error) {
	// Build dynamic query with WHERE clauses based on search parameters
	baseQuery := `
		SELECT id, identifier_system, identifier_value, active, family_name, given_name, gender, birth_date, created_at, updated_at
		FROM patients
		WHERE 1=1
	`

	// Store query parameters for prepared statement
	queryParameters := []interface{}{}
	parameterIndex := 1

	// Add name filter (searches both given_name and family_name)
	if searchParams.Name != "" {
		baseQuery += ` AND (LOWER(given_name) LIKE $` + fmt.Sprint(parameterIndex) + ` OR LOWER(family_name) LIKE $` + fmt.Sprint(parameterIndex) + `)`
		queryParameters = append(queryParameters, "%"+strings.ToLower(searchParams.Name)+"%")
		parameterIndex++
	}

	// Add family name filter
	if searchParams.FamilyName != "" {
		baseQuery += ` AND LOWER(family_name) LIKE $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, "%"+strings.ToLower(searchParams.FamilyName)+"%")
		parameterIndex++
	}

	// Add given name filter
	if searchParams.GivenName != "" {
		baseQuery += ` AND LOWER(given_name) LIKE $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, "%"+strings.ToLower(searchParams.GivenName)+"%")
		parameterIndex++
	}

	// Add gender filter
	if searchParams.Gender != "" {
		baseQuery += ` AND gender = $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, searchParams.Gender)
		parameterIndex++
	}

	// Add birth date exact filter
	if searchParams.BirthDate != nil {
		baseQuery += ` AND birth_date = $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, searchParams.BirthDate)
		parameterIndex++
	}

	// Add birth date greater than or equal filter
	if searchParams.BirthDateGreaterThan != nil {
		baseQuery += ` AND birth_date >= $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, searchParams.BirthDateGreaterThan)
		parameterIndex++
	}

	// Add birth date less than or equal filter
	if searchParams.BirthDateLessThan != nil {
		baseQuery += ` AND birth_date <= $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, searchParams.BirthDateLessThan)
		parameterIndex++
	}

	// Add active status filter
	if searchParams.Active != nil {
		baseQuery += ` AND active = $` + fmt.Sprint(parameterIndex)
		queryParameters = append(queryParameters, *searchParams.Active)
		parameterIndex++
	}

	// Add sorting
	sortBy := "created_at"
	if searchParams.SortBy != "" {
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]bool{
			"name":       true,
			"family_name": true,
			"given_name": true,
			"birthdate":  true,
			"gender":     true,
			"created_at": true,
		}
		if validSortFields[searchParams.SortBy] {
			if searchParams.SortBy == "name" {
				sortBy = "family_name"
			} else if searchParams.SortBy == "birthdate" {
				sortBy = "birth_date"
			} else {
				sortBy = searchParams.SortBy
			}
		}
	}

	sortOrder := "DESC"
	if searchParams.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	baseQuery += ` ORDER BY ` + sortBy + ` ` + sortOrder

	// Add pagination
	baseQuery += ` LIMIT $` + fmt.Sprint(parameterIndex) + ` OFFSET $` + fmt.Sprint(parameterIndex+1)
	queryParameters = append(queryParameters, searchParams.Limit, searchParams.Offset)

	// Execute the query
	rows, queryError := repository.databaseConnection.QueryContext(ctx, baseQuery, queryParameters...)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	// Scan results into patient structs
	patients := []*models.Patient{}
	for rows.Next() {
		patient := &models.Patient{}
		scanError := rows.Scan(
			&patient.ID,
			&patient.IdentifierSystem,
			&patient.IdentifierValue,
			&patient.Active,
			&patient.FamilyName,
			&patient.GivenName,
			&patient.Gender,
			&patient.BirthDate,
			&patient.CreatedAt,
			&patient.UpdatedAt,
		)
		if scanError != nil {
			return nil, scanError
		}
		patients = append(patients, patient)
	}

	// Check for errors during iteration
	if rowsError := rows.Err(); rowsError != nil {
		return nil, rowsError
	}

	return patients, nil
}

// Delete removes a patient record from the database by ID
func (repository *PostgresPatientRepository) Delete(ctx context.Context, patientID string) error {
	// SQL query to delete a patient by ID
	deleteQuery := `DELETE FROM patients WHERE id = $1`

	// Execute the delete query
	_, execError := repository.databaseConnection.ExecContext(ctx, deleteQuery, patientID)

	return execError
}
