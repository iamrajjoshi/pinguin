.PHONY: up down build test migrate

# Start all services
up:
	docker-compose -f docker/docker-compose-core.yml -f docker/docker-compose-dev.yml up --build -d

# Stop all services
down:
	docker-compose -f docker/docker-compose-core.yml -f docker/docker-compose-dev.yml down

# Remove volumes via docker-compose
remove-volumes:
	docker-compose -f docker/docker-compose-core.yml -f docker/docker-compose-dev.yml down -v


# Build the Go application
build:
	go build -o bin/server ./cmd/server/main.go

# Run tests
test:
	go test ./...

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Apply migrations
migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/uptime?sslmode=disable" up

# Rollback migrations
migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/uptime?sslmode=disable" down

# DB shell
db-shell:
	docker-compose -f docker/docker-compose-core.yml -f docker/docker-compose-dev.yml exec db psql -U postgres -d uptime