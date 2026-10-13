.PHONY: build run test clean deps

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run with hot reload (requires air)
dev:
	air

# Database setup - run the DDL
db-setup:
	psql -h localhost -U postgres -d auth_db -f ddl/init.sql

# Create database
db-create:
	psql -h localhost -U postgres -c "CREATE DATABASE auth_db;"
