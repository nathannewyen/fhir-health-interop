package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ObservationRepository defines the interface for observation data access
type ObservationRepository interface {
	Create(ctx context.Context, observation *models.Observation) (*models.Observation, error)
	GetByID(ctx context.Context, observationID string) (*models.Observation, error)
	GetByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*models.Observation, error)
	GetAll(ctx context.Context, limit int, offset int) ([]*models.Observation, error)
	Search(ctx context.Context, searchParams *models.ObservationSearchParams) ([]*models.Observation, error)
	Update(ctx context.Context, observation *models.Observation) (*models.Observation, error)
	Delete(ctx context.Context, observationID string) error
}

// MongoObservationRepository implements ObservationRepository using MongoDB
type MongoObservationRepository struct {
	collection *mongo.Collection
}

// NewMongoObservationRepository creates a new MongoDB observation repository
func NewMongoObservationRepository(database *mongo.Database) *MongoObservationRepository {
	collection := database.Collection("observations")
	return &MongoObservationRepository{
		collection: collection,
	}
}

// Create inserts a new observation into MongoDB
func (repository *MongoObservationRepository) Create(ctx context.Context, observation *models.Observation) (*models.Observation, error) {
	// Set timestamps
	observation.CreatedAt = time.Now()
	observation.UpdatedAt = time.Now()

	// Insert document
	result, insertError := repository.collection.InsertOne(ctx, observation)
	if insertError != nil {
		return nil, fmt.Errorf("failed to insert observation: %w", insertError)
	}

	// Set the generated ID
	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		observation.ID = objectID.Hex()
	}

	return observation, nil
}

// GetByID retrieves an observation by ID
func (repository *MongoObservationRepository) GetByID(ctx context.Context, observationID string) (*models.Observation, error) {
	// Convert string ID to ObjectID
	objectID, convertError := primitive.ObjectIDFromHex(observationID)
	if convertError != nil {
		return nil, fmt.Errorf("invalid observation ID: %w", convertError)
	}

	// Find document
	var observation models.Observation
	filter := bson.M{"_id": objectID}
	findError := repository.collection.FindOne(ctx, filter).Decode(&observation)
	if findError != nil {
		if findError == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("observation not found")
		}
		return nil, fmt.Errorf("failed to find observation: %w", findError)
	}

	return &observation, nil
}

// GetByPatientID retrieves all observations for a specific patient
func (repository *MongoObservationRepository) GetByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*models.Observation, error) {
	// Build filter
	filter := bson.M{"patient_id": patientID}

	// Set options
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))
	findOptions.SetSort(bson.M{"created_at": -1}) // Sort by newest first

	// Execute query
	cursor, findError := repository.collection.Find(ctx, filter, findOptions)
	if findError != nil {
		return nil, fmt.Errorf("failed to find observations: %w", findError)
	}
	defer cursor.Close(ctx)

	// Decode results
	observations := make([]*models.Observation, 0)
	if decodeError := cursor.All(ctx, &observations); decodeError != nil {
		return nil, fmt.Errorf("failed to decode observations: %w", decodeError)
	}

	return observations, nil
}

// GetAll retrieves all observations with pagination
func (repository *MongoObservationRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Observation, error) {
	// Set options
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))
	findOptions.SetSort(bson.M{"created_at": -1})

	// Execute query
	cursor, findError := repository.collection.Find(ctx, bson.M{}, findOptions)
	if findError != nil {
		return nil, fmt.Errorf("failed to find observations: %w", findError)
	}
	defer cursor.Close(ctx)

	// Decode results
	observations := make([]*models.Observation, 0)
	if decodeError := cursor.All(ctx, &observations); decodeError != nil {
		return nil, fmt.Errorf("failed to decode observations: %w", decodeError)
	}

	return observations, nil
}

// Search retrieves observations matching the search criteria with dynamic filtering
func (repository *MongoObservationRepository) Search(ctx context.Context, searchParams *models.ObservationSearchParams) ([]*models.Observation, error) {
	// Build dynamic filter based on search parameters
	filter := bson.M{}

	// Add patient ID filter
	if searchParams.PatientID != "" {
		filter["patient_id"] = searchParams.PatientID
	}

	// Add code filter
	if searchParams.Code != "" {
		filter["code"] = searchParams.Code
	}

	// Add category filter
	if searchParams.Category != "" {
		filter["category"] = searchParams.Category
	}

	// Add status filter
	if searchParams.Status != "" {
		filter["status"] = searchParams.Status
	}

	// Add date range filters
	if searchParams.DateGreaterThan != nil {
		if filter["effective_date"] == nil {
			filter["effective_date"] = bson.M{}
		}
		filter["effective_date"].(bson.M)["$gte"] = searchParams.DateGreaterThan
	}

	if searchParams.DateLessThan != nil {
		if filter["effective_date"] == nil {
			filter["effective_date"] = bson.M{}
		}
		filter["effective_date"].(bson.M)["$lte"] = searchParams.DateLessThan
	}

	// Build options for sorting and pagination
	findOptions := options.Find()
	findOptions.SetLimit(int64(searchParams.Limit))
	findOptions.SetSkip(int64(searchParams.Offset))

	// Add sorting
	sortBy := "created_at"
	sortOrder := -1 // -1 for descending, 1 for ascending

	if searchParams.SortBy != "" {
		// Validate sort field
		validSortFields := map[string]string{
			"effective_date": "effective_date",
			"code":           "code",
			"status":         "status",
			"created_at":     "created_at",
		}
		if field, valid := validSortFields[searchParams.SortBy]; valid {
			sortBy = field
		}
	}

	if searchParams.SortOrder == "asc" {
		sortOrder = 1
	}

	findOptions.SetSort(bson.M{sortBy: sortOrder})

	// Execute query
	cursor, findError := repository.collection.Find(ctx, filter, findOptions)
	if findError != nil {
		return nil, fmt.Errorf("failed to search observations: %w", findError)
	}
	defer cursor.Close(ctx)

	// Decode results
	observations := make([]*models.Observation, 0)
	if decodeError := cursor.All(ctx, &observations); decodeError != nil {
		return nil, fmt.Errorf("failed to decode observations: %w", decodeError)
	}

	return observations, nil
}

// Update modifies an existing observation
func (repository *MongoObservationRepository) Update(ctx context.Context, observation *models.Observation) (*models.Observation, error) {
	// Convert string ID to ObjectID
	objectID, convertError := primitive.ObjectIDFromHex(observation.ID)
	if convertError != nil {
		return nil, fmt.Errorf("invalid observation ID: %w", convertError)
	}

	// Update timestamp
	observation.UpdatedAt = time.Now()

	// Build filter and update (exclude _id field as it's immutable in MongoDB)
	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"patient_id":     observation.PatientID,
			"status":         observation.Status,
			"category":       observation.Category,
			"code":           observation.Code,
			"code_system":    observation.CodeSystem,
			"code_display":   observation.CodeDisplay,
			"value_quantity": observation.ValueQuantity,
			"value_unit":     observation.ValueUnit,
			"value_string":   observation.ValueString,
			"effective_date": observation.EffectiveDate,
			"issued_date":    observation.IssuedDate,
			"components":     observation.Components,
			"updated_at":     observation.UpdatedAt,
		},
	}

	// Execute update
	updateResult, updateError := repository.collection.UpdateOne(ctx, filter, update)
	if updateError != nil {
		return nil, fmt.Errorf("failed to update observation: %w", updateError)
	}

	if updateResult.MatchedCount == 0 {
		return nil, fmt.Errorf("observation not found")
	}

	return observation, nil
}

// Delete removes an observation by ID
func (repository *MongoObservationRepository) Delete(ctx context.Context, observationID string) error {
	// Convert string ID to ObjectID
	objectID, convertError := primitive.ObjectIDFromHex(observationID)
	if convertError != nil {
		return fmt.Errorf("invalid observation ID: %w", convertError)
	}

	// Delete document
	filter := bson.M{"_id": objectID}
	deleteResult, deleteError := repository.collection.DeleteOne(ctx, filter)
	if deleteError != nil {
		return fmt.Errorf("failed to delete observation: %w", deleteError)
	}

	if deleteResult.DeletedCount == 0 {
		return fmt.Errorf("observation not found")
	}

	return nil
}
