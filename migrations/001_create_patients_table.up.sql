-- Migration: Create patients table for FHIR Patient resources
-- This table stores patient data that maps to FHIR R4 Patient resource

CREATE TABLE IF NOT EXISTS patients (
    -- Primary key using UUID for distributed systems compatibility
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- FHIR identifier system and value (e.g., MRN, SSN)
    identifier_system VARCHAR(255),
    identifier_value VARCHAR(255),

    -- Patient active status
    active BOOLEAN DEFAULT true,

    -- Name fields (supporting single name for simplicity)
    family_name VARCHAR(255) NOT NULL,
    given_name VARCHAR(255) NOT NULL,

    -- Administrative gender (male, female, other, unknown)
    gender VARCHAR(20),

    -- Birth date in ISO format (YYYY-MM-DD)
    birth_date DATE,

    -- Audit fields for tracking changes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups by identifier
CREATE INDEX idx_patients_identifier ON patients(identifier_system, identifier_value);

-- Index for name searches
CREATE INDEX idx_patients_name ON patients(family_name, given_name);

-- Comment explaining the table purpose
COMMENT ON TABLE patients IS 'Stores FHIR R4 Patient resources for healthcare interoperability';
