package models

import (
	"time"
)

// Observation represents a clinical observation (vitals, lab results, etc.)
// Stored in MongoDB due to variable structure
type Observation struct {
	ID             string                 `bson:"_id,omitempty"`
	PatientID      string                 `bson:"patient_id"`
	Status         string                 `bson:"status"`
	Category       string                 `bson:"category"`
	Code           string                 `bson:"code"`
	CodeSystem     string                 `bson:"code_system"`
	CodeDisplay    string                 `bson:"code_display"`
	ValueQuantity  *float64               `bson:"value_quantity,omitempty"`
	ValueUnit      string                 `bson:"value_unit,omitempty"`
	ValueString    string                 `bson:"value_string,omitempty"`
	EffectiveDate  *time.Time             `bson:"effective_date,omitempty"`
	IssuedDate     time.Time              `bson:"issued_date"`
	Components     []ObservationComponent `bson:"components,omitempty"`
	CreatedAt      time.Time              `bson:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at"`
}

// ObservationComponent represents a component of a complex observation
// Example: Blood pressure has systolic and diastolic components
type ObservationComponent struct {
	Code          string   `bson:"code"`
	CodeSystem    string   `bson:"code_system"`
	CodeDisplay   string   `bson:"code_display"`
	ValueQuantity *float64 `bson:"value_quantity,omitempty"`
	ValueUnit     string   `bson:"value_unit,omitempty"`
	ValueString   string   `bson:"value_string,omitempty"`
}
