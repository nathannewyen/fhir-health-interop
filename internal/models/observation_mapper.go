package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// ObservationMapper converts between domain model and FHIR Observation
type ObservationMapper struct{}

// NewObservationMapper creates a new observation mapper instance
func NewObservationMapper() *ObservationMapper {
	return &ObservationMapper{}
}

// ToFHIR converts domain Observation to FHIR Observation
func (mapper *ObservationMapper) ToFHIR(observation *Observation) *fhir.Observation {
	fhirObservation := &fhir.Observation{}

	// Set ID
	if observation.ID != "" {
		fhirObservation.Id = &observation.ID
	}

	// Set status
	if observation.Status != "" {
		status := fhir.ObservationStatusFinal
		if observation.Status == "registered" {
			status = fhir.ObservationStatusRegistered
		} else if observation.Status == "preliminary" {
			status = fhir.ObservationStatusPreliminary
		} else if observation.Status == "final" {
			status = fhir.ObservationStatusFinal
		} else if observation.Status == "amended" {
			status = fhir.ObservationStatusAmended
		}
		fhirObservation.Status = status
	}

	// Set category
	if observation.Category != "" {
		fhirObservation.Category = []fhir.CodeableConcept{
			{
				Coding: []fhir.Coding{
					{
						Code:    &observation.Category,
						Display: &observation.Category,
					},
				},
			},
		}
	}

	// Set code
	fhirObservation.Code = fhir.CodeableConcept{
		Coding: []fhir.Coding{
			{
				System:  &observation.CodeSystem,
				Code:    &observation.Code,
				Display: &observation.CodeDisplay,
			},
		},
	}

	// Set subject (patient reference)
	if observation.PatientID != "" {
		patientReference := "Patient/" + observation.PatientID
		fhirObservation.Subject = &fhir.Reference{
			Reference: &patientReference,
		}
	}

	// Set effective date
	if observation.EffectiveDate != nil {
		effectiveDateString := observation.EffectiveDate.Format("2006-01-02T15:04:05Z")
		fhirObservation.EffectiveDateTime = &effectiveDateString
	}

	// Set issued date
	issuedInstant := observation.IssuedDate.Format(time.RFC3339)
	fhirObservation.Issued = &issuedInstant

	// Set value (quantity or string)
	if observation.ValueQuantity != nil {
		valueNumber := json.Number(strconv.FormatFloat(*observation.ValueQuantity, 'f', -1, 64))
		fhirObservation.ValueQuantity = &fhir.Quantity{
			Value: &valueNumber,
			Unit:  &observation.ValueUnit,
		}
	} else if observation.ValueString != "" {
		fhirObservation.ValueString = &observation.ValueString
	}

	// Set components
	if len(observation.Components) > 0 {
		fhirComponents := make([]fhir.ObservationComponent, 0, len(observation.Components))
		for _, component := range observation.Components {
			fhirComponent := fhir.ObservationComponent{
				Code: fhir.CodeableConcept{
					Coding: []fhir.Coding{
						{
							System:  &component.CodeSystem,
							Code:    &component.Code,
							Display: &component.CodeDisplay,
						},
					},
				},
			}

			if component.ValueQuantity != nil {
				componentNumber := json.Number(strconv.FormatFloat(*component.ValueQuantity, 'f', -1, 64))
				fhirComponent.ValueQuantity = &fhir.Quantity{
					Value: &componentNumber,
					Unit:  &component.ValueUnit,
				}
			} else if component.ValueString != "" {
				fhirComponent.ValueString = &component.ValueString
			}

			fhirComponents = append(fhirComponents, fhirComponent)
		}
		fhirObservation.Component = fhirComponents
	}

	return fhirObservation
}

// FromFHIR converts FHIR Observation to domain Observation
func (mapper *ObservationMapper) FromFHIR(fhirObservation *fhir.Observation) *Observation {
	// Map FHIR status to string
	statusString := "final"
	switch fhirObservation.Status {
	case fhir.ObservationStatusRegistered:
		statusString = "registered"
	case fhir.ObservationStatusPreliminary:
		statusString = "preliminary"
	case fhir.ObservationStatusFinal:
		statusString = "final"
	case fhir.ObservationStatusAmended:
		statusString = "amended"
	}

	observation := &Observation{
		Status:     statusString,
		IssuedDate: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Extract ID
	if fhirObservation.Id != nil {
		observation.ID = *fhirObservation.Id
	}

	// Extract category
	if len(fhirObservation.Category) > 0 && len(fhirObservation.Category[0].Coding) > 0 {
		if fhirObservation.Category[0].Coding[0].Code != nil {
			observation.Category = *fhirObservation.Category[0].Coding[0].Code
		}
	}

	// Extract code
	if len(fhirObservation.Code.Coding) > 0 {
		coding := fhirObservation.Code.Coding[0]
		if coding.Code != nil {
			observation.Code = *coding.Code
		}
		if coding.System != nil {
			observation.CodeSystem = *coding.System
		}
		if coding.Display != nil {
			observation.CodeDisplay = *coding.Display
		}
	}

	// Extract patient ID from subject reference
	if fhirObservation.Subject != nil && fhirObservation.Subject.Reference != nil {
		// Extract ID from "Patient/123" format
		reference := *fhirObservation.Subject.Reference
		if len(reference) > 8 && reference[:8] == "Patient/" {
			observation.PatientID = reference[8:]
		}
	}

	// Extract effective date
	if fhirObservation.EffectiveDateTime != nil {
		parsedTime, parseError := time.Parse("2006-01-02T15:04:05Z", *fhirObservation.EffectiveDateTime)
		if parseError == nil {
			observation.EffectiveDate = &parsedTime
		}
	}

	// Extract issued date
	if fhirObservation.Issued != nil {
		parsedTime, parseError := time.Parse(time.RFC3339, *fhirObservation.Issued)
		if parseError == nil {
			observation.IssuedDate = parsedTime
		}
	}

	// Extract value
	if fhirObservation.ValueQuantity != nil && fhirObservation.ValueQuantity.Value != nil {
		valueFloat, parseError := fhirObservation.ValueQuantity.Value.Float64()
		if parseError == nil {
			observation.ValueQuantity = &valueFloat
		}
		if fhirObservation.ValueQuantity.Unit != nil {
			observation.ValueUnit = *fhirObservation.ValueQuantity.Unit
		}
	} else if fhirObservation.ValueString != nil {
		observation.ValueString = *fhirObservation.ValueString
	}

	// Extract components
	if len(fhirObservation.Component) > 0 {
		components := make([]ObservationComponent, 0, len(fhirObservation.Component))
		for _, fhirComponent := range fhirObservation.Component {
			component := ObservationComponent{}

			if len(fhirComponent.Code.Coding) > 0 {
				coding := fhirComponent.Code.Coding[0]
				if coding.Code != nil {
					component.Code = *coding.Code
				}
				if coding.System != nil {
					component.CodeSystem = *coding.System
				}
				if coding.Display != nil {
					component.CodeDisplay = *coding.Display
				}
			}

			if fhirComponent.ValueQuantity != nil && fhirComponent.ValueQuantity.Value != nil {
				componentValueFloat, parseError := fhirComponent.ValueQuantity.Value.Float64()
				if parseError == nil {
					component.ValueQuantity = &componentValueFloat
				}
				if fhirComponent.ValueQuantity.Unit != nil {
					component.ValueUnit = *fhirComponent.ValueQuantity.Unit
				}
			} else if fhirComponent.ValueString != nil {
				component.ValueString = *fhirComponent.ValueString
			}

			components = append(components, component)
		}
		observation.Components = components
	}

	return observation
}
