# Build stage
FROM golang:1.23-alpine AS builder

# Build arguments for version information
ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ARG GO_VERSION=unknown

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code and version file
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-X 'certificate-monkey/internal/version.Version=${VERSION}' \
              -X 'certificate-monkey/internal/version.BuildTime=${BUILD_TIME}' \
              -X 'certificate-monkey/internal/version.GitCommit=${GIT_COMMIT}' \
              -X 'certificate-monkey/internal/version.GoVersion=${GO_VERSION}' \
              -s -w" \
    -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for health checks
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy VERSION file for runtime access
COPY --from=builder /app/VERSION .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check using curl instead of wget
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Command to run
CMD ["./main"]
