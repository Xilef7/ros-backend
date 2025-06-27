---
applyTo: "**/*.go"
---
# Project coding standards for Go

Apply the [general coding guidelines](./general-coding.instructions.md) to all code.
Use Go 1.24 or later.
Use Protobuf Opaque API.

## Style Guide
- Use Effective Go as the style guide

## Dependencies
- Use `net/http` for HTTP server.
- Use `log/slog` for logging.
- Use https://github.com/spf13/cobra for CLI
- Use https://github.com/charmbracelet/bubbletea for TUI
- Use https://github.com/grpc/grpc-go for gRPC
- Use https://github.com/gorilla/websocket for WebSocket
