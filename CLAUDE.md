# FHIR Health Interop - Project Instructions

## Project Overview
A FHIR-compliant Healthcare Data Aggregator API built in Go for healthcare interoperability. This project demonstrates FHIR R4 resources, design patterns, and enterprise-grade code quality.

## Tech Stack
- **Language:** Go 1.25+
- **Router:** Chi (github.com/go-chi/chi/v5)
- **Database:** PostgreSQL 15 (Docker)
- **FHIR Library:** github.com/samply/golang-fhir-models/fhir-models/fhir

## Project Structure
```
cmd/server/          # Application entry point (main.go)
internal/
  ├── handlers/      # HTTP request handlers
  ├── models/        # Domain models and FHIR mappers
  ├── repository/    # Database access layer (Repository pattern)
  ├── service/       # Business logic layer
  ├── middleware/    # HTTP middleware
  └── database/      # Database connection helpers
migrations/          # SQL migration files
```

## Code Conventions
- Use descriptive variable names (no single letters except loops)
- Add comments explaining the purpose of functions and complex logic
- Do NOT use fallback operators "||" - handle errors explicitly
- Keep variable names consistent with database column names
- Follow Go standard project layout

## Testing
- Tests live next to source files (*_test.go)
- Run tests: `go test -v ./...`
- All new code must have corresponding tests

## Database
- Connection: `localhost:5432`
- User: `fhir_user`
- Password: `fhir_password`
- Database: `fhir_health_db`
- Start: `docker-compose up -d postgres`

## Common Commands
```bash
# Run server
go run ./cmd/server

# Run all tests
go test -v ./...

# Build
go build -o bin/server ./cmd/server

# Start database
docker-compose up -d postgres

# Run migrations
docker exec -i fhir-postgres psql -U fhir_user -d fhir_health_db < migrations/001_create_patients_table.up.sql
```

## Design Patterns Used
- **Repository Pattern** - Data access abstraction
- **Factory Pattern** - FHIR resource creation (planned)
- **Strategy Pattern** - Data transformation (planned)

## FHIR Resources Implemented
- [x] Patient (basic CRUD)
- [ ] Observation
- [ ] Condition
- [ ] MedicationRequest
