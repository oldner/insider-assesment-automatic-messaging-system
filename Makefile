.PHONY: run build test swag docker-build docker-up docker-down lint

# Run the application locally
run:
	go run cmd/server/main.go

# Build the application
build:
	go build -o main cmd/server/main.go

# Run tests
test:
	go test ./...

# Generate Swagger documentation
swag:
	swag init -g cmd/server/main.go

# Build Docker images
docker-build:
	docker-compose build

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Run linter (requires golangci-lint installed)
lint:
	golangci-lint run

# Seed database with test data
seed:
	cat scripts/seed.sql | docker-compose exec -T postgres psql -U user -d insider_db
