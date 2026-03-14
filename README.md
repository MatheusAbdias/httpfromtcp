# HTTP from TCP

A custom HTTP server implementation built from scratch using raw TCP sockets in Go, without relying on the standard library's `net/http` package.

## Overview

This project demonstrates how HTTP works at the protocol level by implementing:
- HTTP request parsing (request line, headers, body)
- HTTP response construction (status lines, headers, body)
- TCP server with graceful shutdown handling

## Project Structure

```
.
├── cmd/
│   ├── httpserver/       # Main HTTP server application
│   └── tcplistener/      # TCP listener utility
├── internal/
│   ├── constants/        # HTTP constants (CRLF, headers)
│   ├── headers/          # HTTP header parsing and management
│   ├── request/          # HTTP request parsing
│   ├── response/         # HTTP response building
│   └── server/           # TCP server implementation
└── assets/               # Static files (e.g., video content)
```

## Features

- **Custom HTTP Parser**: Parses HTTP requests manually from raw TCP data
- **Response Writer**: Build HTTP responses (status lines, headers, body)
- **Chunked Transfer Encoding**: Support for chunked body transfer with trailers
- **Graceful Shutdown**: Handles SIGINT/SIGTERM for clean server shutdown
- **Multiple Endpoints**:
  - `/` - Returns 200 OK with HTML response
  - `/yourproblem` - Returns 400 Bad Request
  - `/myproblem` - Returns 500 Internal Server Error
  - `/video` - Serves binary video file
  - `/httpbin/*` - Proxies requests to httpbin.org with chunked transfer encoding

## Running the Server

```bash
go run cmd/httpserver/main.go
```

The server starts on port 42069.

## Testing

Run the unit tests:

```bash
go test ./...
```

## Technical Details

- Uses Go's `net` package for TCP socket handling
- Implements HTTP/1.1 protocol
- Dynamic buffer resizing for request parsing
- Supports Content-Length and Transfer-Encoding: chunked
