# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o mdspace ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/mdspace .

# Copy static files
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Set environment defaults
ENV PORT=8080
ENV STATIC_DIR=./static

# Run the application
CMD ["./mdspace"]
