package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// TestObservationMapper_ToFHIR verifies conversion from domain to FHIR
func TestObservationMapper_ToFHIR(t *testing.T) {
	mapper := NewObservationMapper()

	effectiveDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	valueQuantity := 120.5
	observation := &Observation{
		ID:            "test-id-123",
		PatientID:     "patient-456",
		Status:        "final",
		Category:      "vital-signs",
		Code:          "85354-9",
		CodeSystem:    "http://loinc.org",
		CodeDisplay:   "Blood pressure",
		ValueQuantity: &valueQuantity,
		ValueUnit:     "mmHg",
		EffectiveDate: &effectiveDate,
		IssuedDate:    time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	fhirObservation := mapper.ToFHIR(observation)

	if fhirObservation.Id == nil || *fhirObservation.Id != "test-id-123" {
		t.Error("Expected ID to be set")
	}
	if fhirObservation.Status != fhir.ObservationStatusFinal {
		t.Errorf("Expected status final, got %v", fhirObservation.Status)
	}
	if len(fhirObservation.Code.Coding) == 0 || *fhirObservation.Code.Coding[0].Code != "85354-9" {
		t.Error("Expected code to be set")
	}
	if fhirObservation.Subject == nil || !strings.Contains(*fhirObservation.Subject.Reference, "patient-456") {
		t.Error("Expected subject reference to contain patient ID")
	}
	if fhirObservation.ValueQuantity == nil {
		t.Fatal("Expected value quantity to be set")
	}
	valueFloat, _ := fhirObservation.ValueQuantity.Value.Float64()
	if valueFloat != 120.5 {
		t.Errorf("Expected value 120.5, got %f", valueFloat)
	}
}

// TestObservationMapper_ToFHIR_WithComponents verifies component conversion
func TestObservationMapper_ToFHIR_WithComponents(t *testing.T) {
	mapper := NewObservationMapper()

	systolic := 120.0
	diastolic := 80.0
	observation := &Observation{
		ID:     "test-id",
		Status: "final",
		Code:   "85354-9",
		Components: []ObservationComponent{
			{
				Code:          "8480-6",
				CodeSystem:    "http://loinc.org",
				CodeDisplay:   "Systolic",
				ValueQuantity: &systolic,
				ValueUnit:     "mmHg",
			},
			{
				Code:          "8462-4",
				CodeSystem:    "http://loinc.org",
				CodeDisplay:   "Diastolic",
				ValueQuantity: &diastolic,
				ValueUnit:     "mmHg",
			},
		},
	}

	fhirObservation := mapper.ToFHIR(observation)

	if len(fhirObservation.Component) != 2 {
		t.Errorf("Expected 2 components, got %d", len(fhirObservation.Component))
	}
	if fhirObservation.Component[0].ValueQuantity == nil {
		t.Error("Expected component value quantity")
	}
}

// TestObservationMapper_ToFHIR_WithValueString verifies string value conversion
func TestObservationMapper_ToFHIR_WithValueString(t *testing.T) {
	mapper := NewObservationMapper()

	observation := &Observation{
		ID:          "test-id",
		Status:      "final",
		Code:        "test-code",
		ValueString: "Normal",
	}

	fhirObservation := mapper.ToFHIR(observation)

	if fhirObservation.ValueString == nil || *fhirObservation.ValueString != "Normal" {
		t.Error("Expected value string to be set")
	}
}

// TestObservationMapper_ToFHIR_StatusMapping verifies all status mappings
func TestObservationMapper_ToFHIR_StatusMapping(t *testing.T) {
	mapper := NewObservationMapper()

	testCases := []struct {
		status   string
		expected fhir.ObservationStatus
	}{
		{"registered", fhir.ObservationStatusRegistered},
		{"preliminary", fhir.ObservationStatusPreliminary},
		{"final", fhir.ObservationStatusFinal},
		{"amended", fhir.ObservationStatusAmended},
		{"unknown", fhir.ObservationStatusFinal}, // default
	}

	for _, tc := range testCases {
		observation := &Observation{Status: tc.status, Code: "test"}
		fhirObservation := mapper.ToFHIR(observation)
		if fhirObservation.Status != tc.expected {
			t.Errorf("Status %s: expected %v, got %v", tc.status, tc.expected, fhirObservation.Status)
		}
	}
}

// TestObservationMapper_FromFHIR verifies conversion from FHIR to domain
func TestObservationMapper_FromFHIR(t *testing.T) {
	mapper := NewObservationMapper()

	id := "fhir-id-123"
	category := "vital-signs"
	code := "85354-9"
	codeSystem := "http://loinc.org"
	codeDisplay := "Blood pressure"
	patientRef := "Patient/patient-456"
	effectiveDateTime := "2024-01-15T10:30:00Z"
	issued := "2024-01-15T11:00:00Z"
	valueNumber := json.Number("120.5")
	unit := "mmHg"

	fhirObservation := &fhir.Observation{
		Id:     &id,
		Status: fhir.ObservationStatusFinal,
		Category: []fhir.CodeableConcept{
			{
				Coding: []fhir.Coding{
					{Code: &category},
				},
			},
		},
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{
				{
					System:  &codeSystem,
					Code:    &code,
					Display: &codeDisplay,
				},
			},
		},
		Subject: &fhir.Reference{
			Reference: &patientRef,
		},
		EffectiveDateTime: &effectiveDateTime,
		Issued:            &issued,
		ValueQuantity: &fhir.Quantity{
			Value: &valueNumber,
			Unit:  &unit,
		},
	}

	observation := mapper.FromFHIR(fhirObservation)

	if observation.ID != "fhir-id-123" {
		t.Errorf("Expected ID fhir-id-123, got %s", observation.ID)
	}
	if observation.Status != "final" {
		t.Errorf("Expected status final, got %s", observation.Status)
	}
	if observation.PatientID != "patient-456" {
		t.Errorf("Expected patient ID patient-456, got %s", observation.PatientID)
	}
	if observation.Code != "85354-9" {
		t.Errorf("Expected code 85354-9, got %s", observation.Code)
	}
	if observation.ValueQuantity == nil || *observation.ValueQuantity != 120.5 {
		t.Error("Expected value quantity 120.5")
	}
}

// TestObservationMapper_FromFHIR_WithComponents verifies component conversion
func TestObservationMapper_FromFHIR_WithComponents(t *testing.T) {
	mapper := NewObservationMapper()

	code1 := "8480-6"
	code2 := "8462-4"
	system := "http://loinc.org"
	display1 := "Systolic"
	display2 := "Diastolic"
	value1 := json.Number("120")
	value2 := json.Number("80")
	unit := "mmHg"

	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code1}},
		},
		Component: []fhir.ObservationComponent{
			{
				Code: fhir.CodeableConcept{
					Coding: []fhir.Coding{
						{System: &system, Code: &code1, Display: &display1},
					},
				},
				ValueQuantity: &fhir.Quantity{Value: &value1, Unit: &unit},
			},
			{
				Code: fhir.CodeableConcept{
					Coding: []fhir.Coding{
						{System: &system, Code: &code2, Display: &display2},
					},
				},
				ValueQuantity: &fhir.Quantity{Value: &value2, Unit: &unit},
			},
		},
	}

	observation := mapper.FromFHIR(fhirObservation)

	if len(observation.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(observation.Components))
	}
	if observation.Components[0].ValueQuantity == nil || *observation.Components[0].ValueQuantity != 120 {
		t.Error("Expected first component value 120")
	}
}

// TestObservationMapper_FromFHIR_WithValueString verifies string value
func TestObservationMapper_FromFHIR_WithValueString(t *testing.T) {
	mapper := NewObservationMapper()

	code := "test-code"
	valueString := "Normal"
	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
		ValueString: &valueString,
	}

	observation := mapper.FromFHIR(fhirObservation)

	if observation.ValueString != "Normal" {
		t.Errorf("Expected value string Normal, got %s", observation.ValueString)
	}
}

// TestObservationMapper_FromFHIR_StatusMapping verifies status conversion
func TestObservationMapper_FromFHIR_StatusMapping(t *testing.T) {
	mapper := NewObservationMapper()

	testCases := []struct {
		status   fhir.ObservationStatus
		expected string
	}{
		{fhir.ObservationStatusRegistered, "registered"},
		{fhir.ObservationStatusPreliminary, "preliminary"},
		{fhir.ObservationStatusFinal, "final"},
		{fhir.ObservationStatusAmended, "amended"},
	}

	code := "test"
	for _, tc := range testCases {
		fhirObservation := &fhir.Observation{
			Status: tc.status,
			Code: fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: &code}},
			},
		}
		observation := mapper.FromFHIR(fhirObservation)
		if observation.Status != tc.expected {
			t.Errorf("Expected status %s, got %s", tc.expected, observation.Status)
		}
	}
}

// TestObservationMapper_FromFHIR_ComponentWithValueString verifies component string values
func TestObservationMapper_FromFHIR_ComponentWithValueString(t *testing.T) {
	mapper := NewObservationMapper()

	code := "test-code"
	componentCode := "component-code"
	valueString := "Normal"

	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
		Component: []fhir.ObservationComponent{
			{
				Code: fhir.CodeableConcept{
					Coding: []fhir.Coding{{Code: &componentCode}},
				},
				ValueString: &valueString,
			},
		},
	}

	observation := mapper.FromFHIR(fhirObservation)

	if len(observation.Components) != 1 {
		t.Fatal("Expected 1 component")
	}
	if observation.Components[0].ValueString != "Normal" {
		t.Errorf("Expected component value string Normal, got %s", observation.Components[0].ValueString)
	}
}

// TestNewObservationMapper verifies constructor
func TestNewObservationMapper(t *testing.T) {
	mapper := NewObservationMapper()
	if mapper == nil {
		t.Error("Expected non-nil mapper")
	}
}
