# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary (CGO disabled for pure Go SQLite)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Create data directory with correct ownership (nonroot UID=65532)
RUN mkdir -p /data && chown 65532:65532 /data

# Stage 2: Runtime (Distroless)
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /server /server

# Copy data directory with correct ownership
COPY --from=builder --chown=nonroot:nonroot /data /data

VOLUME /data

EXPOSE 8080

ENTRYPOINT ["/server"]
