# Database Connection Testing

Comprehensive integration tests for PostgreSQL and MongoDB connection infrastructure.

## Overview

The database package now has **95.5% test coverage** with 21 comprehensive integration tests that validate connection handling, error scenarios, and configuration.

## Test Statistics

- **Total Tests**: 21
- **PostgreSQL Tests**: 10
- **MongoDB Tests**: 11
- **Coverage**: 95.5%
- **All Tests**: ✅ Passing

## PostgreSQL Connection Tests

### Successful Connection Scenarios (2)
1. **TestNewPostgresConnection_Success** - Valid connection with correct credentials
2. **TestNewPostgresConnection_ConnectionPooling** - Connection pooling and reuse

### Failure Scenarios (4)
3. **TestNewPostgresConnection_InvalidCredentials** - Wrong password
4. **TestNewPostgresConnection_InvalidHost** - Non-existent host
5. **TestNewPostgresConnection_InvalidPort** - Wrong port number
6. **TestNewPostgresConnection_InvalidDatabase** - Non-existent database

### Edge Cases (4)
7. **TestNewPostgresConnection_EmptyConfig** - Empty configuration
8. **TestNewPostgresConnection_MultipleConnections** - Multiple simultaneous connections
9. **TestPostgresConfig_AllFieldsSet** - Configuration struct validation

## MongoDB Connection Tests

### Successful Connection Scenarios (3)
1. **TestNewMongoConnection_Success** - Valid connection with correct credentials
2. **TestNewMongoConnection_DifferentDatabases** - Multiple database connections
3. **TestNewMongoConnection_CollectionOperations** - Collection operations

### Failure Scenarios (5)
4. **TestNewMongoConnection_InvalidCredentials** - Wrong password
5. **TestNewMongoConnection_InvalidHost** - Non-existent host
6. **TestNewMongoConnection_InvalidPort** - Wrong port number
7. **TestNewMongoConnection_ConnectionTimeout** - Timeout handling
8. **TestNewMongoConnection_EmptyConfig** - Empty configuration

### Edge Cases (3)
9. **TestNewMongoConnection_URIFormat** - URI formatting validation
10. **TestMongoConfig_AllFieldsSet** - Configuration struct validation
11. **TestNewMongoConnection_MultipleConnections** - Multiple simultaneous connections

## Test Design

### Integration Tests
These are **integration tests** that connect to actual databases running in Docker containers:
- Uses real PostgreSQL (localhost:5432)
- Uses real MongoDB (localhost:27017)
- Tests actual connection logic (no mocking)
- Validates error messages and behavior

### Graceful Degradation
Tests use `t.Skipf()` to skip gracefully when databases aren't available:
```go
if connectionError != nil {
    t.Skipf("Skipping test: MongoDB not available - %v", connectionError)
    return
}
```

### Fast Execution
- Average test duration: < 1 second (success cases)
- Timeout tests: ~5 seconds (expected)
- Total suite: ~15 seconds

## What Gets Tested

### Connection Success
✅ Valid credentials connect successfully
✅ Connection can ping database
✅ Connection can be closed cleanly
✅ Multiple connections work simultaneously
✅ Connection pooling works correctly

### Connection Failures
✅ Invalid credentials fail with proper error
✅ Invalid host fails with proper error
✅ Invalid port fails with proper error
✅ Invalid database fails with proper error
✅ Empty config fails with proper error
✅ Timeouts are respected

### Configuration
✅ Connection string/URI formatting
✅ Config struct field validation
✅ Database name selection
✅ Collection operations (MongoDB)

## Running the Tests

### All Database Tests
```bash
go test -v ./internal/database
```

### With Coverage
```bash
go test ./internal/database -coverprofile=database_coverage.out
go tool cover -html=database_coverage.out
```

### Individual Tests
```bash
# PostgreSQL only
go test -v ./internal/database -run TestNewPostgresConnection

# MongoDB only
go test -v ./internal/database -run TestNewMongoConnection

# Specific test
go test -v ./internal/database -run TestNewPostgresConnection_Success
```

## Coverage Report

```
mongodb.go:22:     NewMongoConnection        100.0%
postgres.go:21:    NewPostgresConnection     87.5%
total:             (statements)              95.5%
```

### Why Not 100%?
The 4.5% uncovered code is in rarely-executed error paths that would require:
- Database driver internal failures
- Specific race conditions
- Edge cases in third-party libraries

These scenarios are difficult to reproduce in integration tests and represent acceptable coverage for production code.

## Benefits for Portfolio

### Demonstrates Professional Skills
1. **Integration Testing** - Testing real infrastructure code
2. **Error Handling** - Comprehensive failure scenario coverage
3. **Test Design** - Graceful degradation when dependencies unavailable
4. **Production Ready** - Tests critical connection logic
5. **Documentation** - Well-documented test purpose and scenarios

### Code Quality Metrics
- ✅ 95.5% code coverage
- ✅ 21 comprehensive tests
- ✅ Tests both success and failure paths
- ✅ Fast execution (< 20 seconds)
- ✅ No flaky tests
- ✅ Clear test names and documentation

## Test Output Example

```
=== RUN   TestNewPostgresConnection_Success
--- PASS: TestNewPostgresConnection_Success (0.01s)
=== RUN   TestNewPostgresConnection_InvalidCredentials
--- PASS: TestNewPostgresConnection_InvalidCredentials (0.01s)
=== RUN   TestNewMongoConnection_Success
--- PASS: TestNewMongoConnection_Success (0.01s)
...
PASS
ok  	github.com/nathannewyen/fhir-health-interop/internal/database	15.605s
```

## Maintenance

### Adding New Tests
1. Follow existing test patterns
2. Use descriptive test names: `TestNewXConnection_Scenario`
3. Use `t.Skipf()` for optional dependencies
4. Test both success and failure cases
5. Clean up connections with `defer`

### Updating Tests
- Update credentials if Docker config changes
- Adjust timeouts if needed
- Add new tests for new connection features

## Integration with CI/CD

These tests are designed to work in CI/CD pipelines that have Docker available:

```yaml
# GitHub Actions example
- name: Start Docker Services
  run: docker-compose up -d

- name: Run Tests
  run: go test ./... -v

- name: Stop Docker Services
  run: docker-compose down
```

## Related Files

- `postgres.go` - PostgreSQL connection implementation
- `mongodb.go` - MongoDB connection implementation
- `postgres_test.go` - PostgreSQL tests (10 tests)
- `mongodb_test.go` - MongoDB tests (11 tests)

---

**Coverage**: 95.5% | **Tests**: 21 | **Status**: ✅ All Passing
