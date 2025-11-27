# Testing Guide

## Running Tests

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestBlockIP ./...

# Run with coverage
go test -v -cover ./... 

# Generate coverage report
go test -v -coverprofile=coverage.out ./... 
go tool cover -html=coverage.out