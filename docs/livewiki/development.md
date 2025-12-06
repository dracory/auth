---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Development

## Prerequisites

*   Go 1.21+
*   Git

## running Tests

Run the full test suite:

```bash
go test -v ./...
```

For coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

*   `internal/api`: API handlers.
*   `internal/ui`: HTML page handlers.
*   `internal/middlewares`: Logic for rate limiting, CSRF.
*   `types`: Shared type definitions.

## Contributing

1.  Fork the repository.
2.  Create a feature branch.
3.  Add tests for your changes.
4.  Ensure `go test ./...` passes.
5.  Submit a Pull Request.
