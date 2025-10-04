# Build stage
FROM golang:1.25.1-alpine AS builder

# Install build dependencies (git for version info)
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files (when they exist)
# COPY go.mod go.sum ./
# RUN go mod download

# Copy source code
COPY . .

# Build the application
# Pure Go build (modernc.org/sqlite is CGO-free)
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
#     -trimpath \
#     -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo 'dev')" \
#     -o momo ./cmd/momo

# Development stage (includes tools)
FROM golang:1.25.1-alpine AS development

# Install development tools
RUN apk add --no-cache \
    git \
    sqlite \
    make \
    curl \
    bash \
    vim \
    tmux \
    ca-certificates \
    tzdata

# Install Go development tools
RUN go install github.com/cosmtrek/air@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /app

# Copy source code
COPY . .

# Environment variables for development
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV MOMOROT_DATA_DIR=/app/data
ENV MOMOROT_CONFIG_PATH=/app/configs
ENV MOMOROT_LOG_LEVEL=debug
ENV ALPHAVANTAGE_API_KEY=

# Create necessary directories
RUN mkdir -p /app/data /app/configs /app/logs /app/exports

# Expose port for potential web UI or API (adjust as needed)
EXPOSE 8080

# Volume for persistent data
VOLUME ["/app/data", "/app/configs", "/app/exports"]

# Default command for development (can be overridden)
CMD ["bash"]

# Production stage (distroless for minimal attack surface)
FROM gcr.io/distroless/static:nonroot AS production

# Note: distroless includes ca-certificates and tzdata
# User nonroot (UID 65532) is built-in

WORKDIR /app

# Copy binary from builder (uncomment when build is working)
# COPY --from=builder /build/momo /app/momo

# Copy configs (ownership handled by distroless nonroot user)
COPY --chown=65532:65532 configs/ /app/configs/

# Environment variables for production
ENV MOMOROT_DATA_DIR=/data
ENV MOMOROT_CONFIG_PATH=/app/configs
ENV MOMOROT_LOG_LEVEL=info
ENV MOMOROT_EXPORT_DIR=/data/exports

# Already running as nonroot (65532) in distroless

# Volume for persistent data (writable by nonroot)
VOLUME ["/data"]

# Health check using ping subcommand
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/app/momo", "--ping"]

# Run the application (uncomment when binary is available)
# CMD ["/app/momo"]

# Distroless has no shell, use a sleep loop for placeholder
# This will be replaced with the actual binary
CMD ["/pause"]