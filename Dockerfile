# Stage 1: Build binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git & ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependency definitions first for Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile static binary with symbol stripping (-s -w) for minimal size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api/main.go

# Stage 2: Minimal runtime environment
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (certificates & timezone data)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy compiled binary and Swagger docs from builder stage
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/docs /app/docs

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

USER appuser

# Expose HTTP port
EXPOSE 8080

# Environment variables
ENV PORT=8080 \
    GIN_MODE=release

# Run API with automatic mock seeder by default
ENTRYPOINT ["/app/api"]
CMD ["--seed"]

