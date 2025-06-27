# Restaurant Ordering System - Backend

This is the backend service for a restaurant ordering system, built with Go and gRPC.

## Features

- Tab management for dining sessions
- Order management with item sharing
- Menu browsing with tag-based filtering
- Customer management with default name generation

## Technologies

- Go 1.24
- PostgreSQL 17
- gRPC
- SQLc
- OpenTelemetry (Observability)

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- Protocol Buffers compiler

## Development Setup

1. Clone the repository:
   ```bash
   git clone [repository-url]
   cd backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start the development environment:
   ```bash
   docker-compose up -d
   ```

4. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

## Project Structure

```
.
├── api/
│   └── proto/          # Protocol Buffers definitions
├── cmd/
│   ├── cli/           # CLI entry point
│   ├── server/        # Server entry point
│   └── tui/           # TUI entry point
├── configs/           # Configuration files
├── internal/
│   └── pkg/
│       ├── config/    # Configuration handling
│       ├── middleware/ # gRPC middleware
│       ├── model/     # Domain models
│       ├── repository/ # Data access layer
│       └── service/   # Business logic
├── pkg/              # Public packages
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Configuration

Configuration is handled through a JSON file located at `configs/config.json`. The following options are available:

```json
{
    "server": {
        "host": "0.0.0.0",
        "port": 50051
    },
    "database": {
        "host": "localhost",
        "port": 5432,
        "user": "postgres",
        "password": "postgres",
        "database": "restaurant",
        "sslMode": "disable"
    },
}
```

## API Documentation

The API is defined using Protocol Buffers and gRPC. For detailed API documentation, please refer to the proto files in the `api/proto` directory.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
