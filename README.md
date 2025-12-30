# QR Code Generator

A minimal, distroless Docker container hosting a simple web UI for generating QR codes with SQLite persistence.

## Features

- Generate QR codes from any text or URL
- History table with all generated QR codes
- Editable labels for organization
- Click to view/download full-size QR images
- Persistent storage via SQLite
- Tiny distroless container (~5MB)

## Quick Start

```bash
# Using Docker Compose
docker compose up -d

# Or run directly
docker run -d -p 8080:8080 -v qr-data:/data ghcr.io/OWNER/qr-code-generator:latest
```

Access the web UI at http://localhost:8080

## Development

### Prerequisites

- Go 1.22+
- Docker

### Local Development

```bash
# Run tests
go test -v ./...

# Build and run locally
go run ./cmd/server

# Build Docker image
docker build -t qr-code-generator .
```

### Project Structure

```
qr-code-generator/
├── cmd/server/           # Application entrypoint
│   ├── main.go
│   └── templates/        # HTML templates (embedded)
├── internal/
│   ├── handler/          # HTTP handlers
│   ├── qrcode/           # QR generation
│   └── storage/          # SQLite storage
├── .github/workflows/    # CI/CD
├── Dockerfile            # Multi-stage distroless build
└── docker-compose.yml    # Local development
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Main page with form and history |
| POST | `/generate` | Generate new QR code |
| GET | `/qr/{id}` | Get QR code image |
| PUT | `/qr/{id}` | Update QR code label |
| DELETE | `/qr/{id}` | Delete QR code |
| GET | `/health` | Health check |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `/data/qrcodes.db` | SQLite database path |

## License

MIT
