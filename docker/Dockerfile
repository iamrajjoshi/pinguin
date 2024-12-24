FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server/main.go

# Development stage
FROM golang:1.21-alpine AS development
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
CMD ["go", "run", "./cmd/server/main.go"]

# Production stage
FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /app/server .
COPY migrations ./migrations
RUN apk add --no-cache ca-certificates
EXPOSE 8080
CMD ["./server"]