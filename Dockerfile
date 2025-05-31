# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base tesseract-ocr-dev leptonica-dev pkgconfig

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o superclass

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    leptonica

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/superclass .

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the default port
EXPOSE 8080

# Environment variables with defaults
ENV PORT=8080
ENV MODEL_TYPE=gpt-4
ENV MODEL_PROVIDER=openai
ENV MAX_COST=0.1
ENV MAX_LATENCY=30

# Command to run
ENTRYPOINT ["./superclass"] 