# Multi-stage build for minimal image size
FROM fcr.fmr.com/library/golang:1.23-alpine AS builder

# Install build dependencies (including for BLS/blst which requires CGO)
RUN apk add --no-cache git make gcc musl-dev linux-headers

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled (required for BLS crypto)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o avax-remote-signer ./cmd/signer

# Final stage - use golang-alpine as runtime to avoid package installation issues
FROM fcr.fmr.com/library/golang:1.23-alpine

# Create non-root user
RUN addgroup -g 1000 signer && \
    adduser -D -u 1000 -G signer signer

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/avax-remote-signer .

# Change ownership
RUN chown -R signer:signer /app

# Switch to non-root user
USER signer

# Expose HTTP API port
EXPOSE 9090

# Health check using curl to hit the /health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

ENTRYPOINT ["/app/avax-remote-signer"]
CMD ["--config", "/config/config.yaml"]
