.PHONY: up down build test migrate

# Start all services
up:
	docker-compose up --build -d

# Stop all services
down:
	docker-compose down

# Remove volumes via docker-compose
remove-volumes:
	docker-compose down -v


# Build the Go application
build:
	go build -o bin/server ./cmd/server/main.go

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
	docker-compose exec db psql -U postgres -d uptime