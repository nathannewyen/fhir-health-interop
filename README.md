# FHIR Health Interop API

A **FHIR R4 compliant** healthcare data aggregator API built with **Go**, demonstrating multi-database architecture, clean code practices, and comprehensive testing.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Test Coverage](https://img.shields.io/badge/Coverage-97%25-brightgreen)
![FHIR](https://img.shields.io/badge/FHIR-R4-orange)

## 🎯 Project Overview

A portfolio project showcasing **RESTful API development in Go** for healthcare interoperability. Implements FHIR (Fast Healthcare Interoperability Resources) standard for exchanging healthcare data between systems.

**Built to demonstrate:**
- Clean architecture with Repository and Service patterns
- Multi-database design (PostgreSQL + MongoDB)
- Comprehensive search/filtering capabilities
- 97% test coverage (service layer)
- FHIR R4 compliance

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP/JSON
       ▼
┌─────────────────────────────────────────┐
│           API Server (Chi)              │
├─────────────────────────────────────────┤
│  Handlers (HTTP Layer)                  │
│    ├── Patient Handler                  │
│    └── Observation Handler              │
├─────────────────────────────────────────┤
│  Services (Business Logic)              │
│    ├── Patient Service                  │
│    └── Observation Service              │
├─────────────────────────────────────────┤
│  Repositories (Data Access)             │
│    ├── Patient Repository               │
│    └── Observation Repository           │
└──────┬──────────────────┬───────────────┘
       │                  │
       ▼                  ▼
┌─────────────┐    ┌─────────────┐
│ PostgreSQL  │    │   MongoDB   │
│  (Patient)  │    │(Observation)│
└─────────────┘    └─────────────┘
```

### Why Two Databases?

- **PostgreSQL** for Patient data: Structured, relational, ACID compliance
- **MongoDB** for Observation data: Flexible schema, handles varied clinical observations

## 🚀 Quick Start

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- (Optional) Bruno API Client for testing

### 1. Start Databases

```bash
# Start PostgreSQL
docker run -d \
  --name fhir-postgres \
  -e POSTGRES_USER=fhir_user \
  -e POSTGRES_PASSWORD=fhir_password \
  -e POSTGRES_DB=fhir_health_db \
  -p 5432:5432 \
  postgres:15

# Start MongoDB
docker run -d \
  --name fhir-mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=fhir_user \
  -e MONGO_INITDB_ROOT_PASSWORD=fhir_password \
  -p 27017:27017 \
  mongo:7
```

### 2. Initialize Database Schema

```bash
psql -h localhost -U fhir_user -d fhir_health_db -f scripts/init-postgres.sql
```

### 3. Run the Server

```bash
# Install dependencies
go mod download

# Run server
go run cmd/server/main.go
```

Server starts at `http://localhost:8080`

### 4. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a patient
curl -X POST http://localhost:8080/fhir/Patient \
  -H "Content-Type: application/fhir+json" \
  -d '{
    "resourceType": "Patient",
    "name": [{"family": "Doe", "given": ["John"]}],
    "gender": "male",
    "birthDate": "1990-01-01"
  }'

# Search patients by gender
curl "http://localhost:8080/fhir/Patient?gender=male"
```

## 📚 API Endpoints

### Patient Resource (PostgreSQL)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/fhir/Patient` | Create patient |
| GET | `/fhir/Patient/{id}` | Get patient by ID |
| GET | `/fhir/Patient` | Search patients (supports filters) |
| PUT | `/fhir/Patient/{id}` | Update patient |
| DELETE | `/fhir/Patient/{id}` | Delete patient |

**Search Parameters:**
- `?name=Smith` - Search by name
- `?gender=male` - Filter by gender
- `?birthdate=ge1990-01-01` - Birth date >= 1990
- `?active=true` - Filter active patients
- `?_sort=-created_at` - Sort descending
- `?_count=20&_offset=0` - Pagination

### Observation Resource (MongoDB)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/fhir/Observation` | Create observation |
| GET | `/fhir/Observation/{id}` | Get observation by ID |
| GET | `/fhir/Observation` | Search observations (supports filters) |
| PUT | `/fhir/Observation/{id}` | Update observation |
| DELETE | `/fhir/Observation/{id}` | Delete observation |

**Search Parameters:**
- `?patient=123` - Filter by patient ID
- `?code=8480-6` - Filter by LOINC code
- `?category=vital-signs` - Filter by category
- `?status=final` - Filter by status
- `?date=ge2024-01-01` - Effective date >= 2024
- `?_sort=-effective_date` - Sort descending

## 🧪 Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Detailed coverage by package
go test ./internal/service -coverprofile=coverage/service.out
go tool cover -func=coverage/service.out

go test ./internal/repository -coverprofile=coverage/repo.out
go tool cover -func=coverage/repo.out
```

### Test Coverage

- **Service Layer:** 97.2% ✅
- **Repository Layer:** 91.2% ✅
- **Total Tests:** 80 tests passing

### Test Types

- **Unit Tests:** Service layer with mocked repositories
- **Integration Tests:** Repository layer with real databases
- **Search Tests:** Comprehensive query testing (33 tests)

## 📁 Project Structure

```
fhir-health-interop/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/                # Database connections
│   │   ├── postgres.go          # PostgreSQL connection
│   │   ├── mongodb.go           # MongoDB connection
│   │   └── *_test.go            # Database tests (95.5% coverage)
│   ├── handlers/                # HTTP handlers
│   │   ├── patient.go           # Patient CRUD endpoints
│   │   ├── observation.go       # Observation CRUD endpoints
│   │   └── *_test.go            # Handler tests
│   ├── service/                 # Business logic
│   │   ├── patient_service.go   # Patient business logic
│   │   ├── observation_service.go
│   │   └── *_test.go            # Service tests (97.2% coverage)
│   ├── repository/              # Data access
│   │   ├── patient_repository.go
│   │   ├── observation_repository.go
│   │   ├── *_search_test.go     # Search tests (33 tests)
│   │   └── *_test.go            # Repository tests (91.2% coverage)
│   ├── models/                  # Domain models
│   │   ├── patient.go
│   │   ├── observation.go
│   │   └── search_params.go     # Search parameter structs
│   ├── mappers/                 # FHIR ↔ Domain conversion
│   │   ├── patient_mapper.go
│   │   └── observation_mapper.go
│   ├── middleware/              # HTTP middleware
│   │   ├── logger.go
│   │   ├── error_handler.go
│   │   └── validator.go
│   ├── errors/                  # Custom error types
│   └── utils/                   # Utilities
│       └── query_parser.go      # HTTP query parser
├── scripts/
│   └── init-postgres.sql        # Database schema
├── bruno-collections/           # API test collection (23 requests)
│   └── fhir-health-interop/
│       ├── Patient/             # Patient endpoints
│       ├── Observation/         # Observation endpoints
│       └── README.md            # Collection documentation
├── go.mod
└── README.md
```

## 🛠️ Technology Stack

### Core
- **Language:** Go 1.25+
- **Router:** Chi v5 (lightweight, fast HTTP router)
- **Logging:** Zerolog (structured logging)

### Databases
- **PostgreSQL 15:** Patient resource storage
- **MongoDB 7:** Observation resource storage

### FHIR
- **Library:** github.com/samply/golang-fhir-models
- **Version:** FHIR R4

### Testing
- **Framework:** Go testing package
- **Coverage:** 97% service, 91% repository
- **Integration:** Docker-based database tests

### API Testing
- **Tool:** Bruno API Client
- **Requests:** 23 comprehensive test requests

## 🔍 Key Features

### 1. FHIR R4 Compliance
- Proper FHIR resource structure
- FHIR search parameters
- Date comparison prefixes (ge, le, gt, lt, eq)
- FHIR-compliant error responses

### 2. Advanced Search
- Multi-parameter filtering
- Date range queries
- Sorting (ascending/descending)
- Pagination
- Case-insensitive search
- Partial string matching

### 3. Clean Architecture
- **Separation of Concerns:** Handler → Service → Repository
- **Interface-based Design:** Easy to mock for testing
- **Repository Pattern:** Database abstraction
- **Dependency Injection:** Testable components

### 4. Production Practices
- Structured logging with correlation IDs
- Comprehensive error handling
- Request validation
- Health check endpoint
- Graceful error responses

## 💡 What I Learned

Building this project helped me learn:

### Go Programming
- Building REST APIs with Chi router
- Working with interfaces and dependency injection
- Go testing patterns (unit + integration)
- Proper error handling in Go
- Struct tags and JSON marshaling

### Database Management
- PostgreSQL with `database/sql`
- MongoDB with official Go driver
- Database connection pooling
- Transaction handling
- Multi-database architecture

### Software Architecture
- Repository pattern implementation
- Service layer design
- Clean architecture principles
- Mapper pattern for data transformation
- Middleware chain design

### Healthcare Domain
- FHIR R4 resource specifications
- Healthcare data interoperability
- Clinical observation modeling
- Patient demographic management

### Testing
- Unit testing with mocks
- Integration testing with Docker
- Test coverage analysis
- Table-driven tests in Go

## 🚀 Running in Production

### Build Binary

```bash
go build -o bin/fhir-api cmd/server/main.go
```

### Environment Variables

```bash
# PostgreSQL
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=fhir_user
export POSTGRES_PASSWORD=fhir_password
export POSTGRES_DB=fhir_health_db

# MongoDB
export MONGO_HOST=localhost
export MONGO_PORT=27017
export MONGO_USER=fhir_user
export MONGO_PASSWORD=fhir_password

# Server
export SERVER_PORT=8080
```

### Run Binary

```bash
./bin/fhir-api
```

## 📊 Performance

- **Startup Time:** < 1 second
- **Memory Usage:** ~20MB idle
- **Request Latency:** < 10ms (local databases)
- **Search Performance:** Optimized with database indexes

## 📝 License

MIT License - feel free to use for learning or portfolio purposes.

## 🤝 Contact

**Nathan Newyen**
- GitHub: [github.com/nathannewyen](https://github.com/nathannewyen)
- LinkedIn: [linkedin.com/in/nathannewyen](https://www.linkedin.com/in/nhannguyen3112/)

---

**Built with ❤️ using Go** | Portfolio Project | 2024
