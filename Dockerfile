# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy files from builder
COPY --from=builder /app/server .
COPY --from=builder /app/configs/config.json ./configs/
COPY --from=builder /app/certs/server_cert.pem ./certs/
COPY --from=builder /app/certs/server_key.pem ./certs/

# Expose gRPC port
EXPOSE 50051

# Run the binary
CMD ["./server"]
