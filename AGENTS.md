# AGENTS.md - Coding Agent Instructions for temp-recorder

## Project Overview

Go application that reads temperature data from a serial port and stores it in a MariaDB database. Runs as an OCI container. Code comments, log messages, and error messages are written in **German**.

- **Language:** Go 1.21
- **Module:** `temp-recorder`
- **Layout:** Standard Go project structure (`cmd/` for entrypoints, `internal/` for private packages)
- **Packages:** `config`, `database`, `serial` (all under `internal/`)
- **Dependencies:** `github.com/go-sql-driver/mysql`, `github.com/tarm/serial`

## Build Commands

```bash
# Install dependencies
go mod download

# Build the binary
go build -o temp-recorder ./cmd/temp-recorder

# Build optimized (production, matches Dockerfile)
CGO_ENABLED=0 go build -ldflags="-s -w" -o temp-recorder ./cmd/temp-recorder

# Run locally (DB_PASSWORD is required)
DB_PASSWORD=test ./temp-recorder

# Build Docker image
docker build -t temp-recorder .

# Start full stack (app + MariaDB + Adminer)
docker compose up -d
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test by name
go test -v -run TestFunctionName ./internal/package/

# Run tests for a single package
go test -v ./internal/config/

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

Note: No tests exist yet. When adding tests, follow Go conventions: place `*_test.go` files alongside the source files in the same package.

## Lint / Vet

No linter configuration exists. Use standard Go tooling:

```bash
# Vet (static analysis)
go vet ./...

# Format code
gofmt -w .

# Simplify code
gofmt -s -w .
```

If adding a linter, prefer `golangci-lint` with a `.golangci.yml` config.

## Code Style Guidelines

### Project Structure

```
cmd/
  temp-recorder/
    main.go            # Application entrypoint
internal/
  config/
    config.go          # Environment-based configuration
  database/
    database.go        # MariaDB connection and queries
  serial/
    reader.go          # Serial port reading logic
```

### Import Order

Use three groups separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"

    // 2. Internal packages
    "temp-recorder/internal/config"

    // 3. Third-party packages
    "github.com/tarm/serial"
)
```

Blank import for side effects uses underscore prefix with a comment if needed:

```go
_ "github.com/go-sql-driver/mysql"
```

### Naming Conventions

- **Exported types/functions:** PascalCase (`Config`, `NewReader`, `SaveTemperature`)
- **Unexported types/functions:** camelCase (`getEnv`, `parseTemperature`, `getEnvAsInt`)
- **Struct fields:** PascalCase for exported (`SerialPort`, `BaudRate`)
- **Variables:** camelCase (`tempChan`, `sigChan`, `cfg`)
- **Receiver names:** Short, typically 1-2 letters matching the type (`db *DB`, `r *Reader`, `c *Config`)
- **Interfaces:** Name by behavior, use `-er` suffix when appropriate
- **Config fields:** Match the environment variable names in PascalCase (e.g., `DB_HOST` -> `DBHost`)

### Error Handling

- Always wrap errors with `fmt.Errorf("context: %w", err)` to preserve the error chain
- Error message prefix describes what failed, written in **German** (lowercase start):
  ```go
  return nil, fmt.Errorf("fehler beim Öffnen der Datenbankverbindung: %w", err)
  ```
- Use `log.Fatalf` only in `main()` for unrecoverable startup errors
- Use `log.Printf` for non-fatal runtime errors (continue operation)
- Return errors rather than panicking; never use `panic` in library code
- Check errors immediately after the call that produces them

### Types and Structs

- Struct types get a doc comment in German explaining their purpose:
  ```go
  // DB ist ein Wrapper für die Datenbankverbindung
  type DB struct { ... }
  ```
- Constructor functions follow the `New` or `NewXxx` pattern and return `(*Type, error)`
- Types that hold resources (DB connections, ports) must implement a `Close() error` method
- Use `defer obj.Close()` at the call site immediately after successful creation

### Configuration

- All configuration is via environment variables (no config files)
- Use helper functions `getEnv(key, defaultValue)` and `getEnvAsInt(key, defaultValue)` for reading env vars
- Provide sensible defaults; fail explicitly when required values (like `DB_PASSWORD`) are missing
- Validate configuration in the `Load()` function before returning

### Comments and Documentation

- All doc comments, inline comments, log messages, and error messages are in **German**
- Exported functions and types must have a doc comment starting with the name:
  ```go
  // Load lädt die Konfiguration aus Umgebungsvariablen
  func Load() (*Config, error) { ... }
  ```

### Concurrency Patterns

- Use `context.Context` for cancellation and graceful shutdown
- Use buffered channels for data pipelines (`make(chan float64, 100)`)
- Handle `ctx.Done()` in select statements for clean goroutine termination
- Signal handling via `os/signal` with `SIGINT` and `SIGTERM`

### Formatting

- Use `gofmt` standard formatting (tabs for indentation, no trailing whitespace)
- No custom formatter configuration exists; standard Go formatting applies
- SQL queries: use backtick multi-line strings, indented for readability

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `SERIAL_PORT` | Serial device path | `/dev/ttyUSB0` |
| `BAUD_RATE` | Serial baud rate | `9600` |
| `READ_INTERVAL` | Debounce interval in seconds | `60` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database user | `tempuser` |
| `DB_PASSWORD` | Database password | *(required)* |
| `DB_NAME` | Database name | `temperatures` |

### Docker

- Multi-stage build: `golang:1.21-alpine` builder, `alpine:3.19` runtime
- Binary runs as non-root `appuser` in `dialout` group (for serial device access)
- CGO is disabled (`CGO_ENABLED=0`) for static binary
- Container requires `--device` flag for serial port passthrough
