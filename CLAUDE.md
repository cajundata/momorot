# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**momorot-tui** is a Go-based Terminal User Interface (TUI) application. The project is currently in initial setup phase with the following planned architecture:

## Project Structure

```
momorot-tui/
├── cmd/momo/           # Main application entry point
├── internal/           # Internal packages (not exposed to external imports)
│   ├── analytics/      # Analytics and metrics functionality
│   ├── config/         # Configuration management
│   ├── db/            # Database layer (SQLite)
│   ├── export/        # Data export functionality
│   ├── fetch/         # Data fetching/retrieval
│   ├── logx/          # Logging utilities
│   ├── ui/            # Terminal UI components
│   └── version/       # Version management
├── configs/           # Configuration files
├── data/             # Data storage directory
├── test/             # Test files and fixtures
└── docs/             # Documentation
```

## Development Commands

### Build Commands
```bash
# Build the application
go build -o momorot-tui ./cmd/momo

# Build with version information
go build -ldflags "-X main.version=$(git describe --tags --always)" -o momorot-tui ./cmd/momo

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o momorot-tui-linux ./cmd/momo
GOOS=darwin GOARCH=amd64 go build -o momorot-tui-darwin ./cmd/momo
GOOS=windows GOARCH=amd64 go build -o momorot-tui.exe ./cmd/momo
```

### Development Commands
```bash
# Run the application
go run ./cmd/momo

# Run with hot reload (requires air)
air

# Format code
go fmt ./...
gofmt -w .

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...

# Run staticcheck (requires staticcheck)
staticcheck ./...
```

### Test Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests in verbose mode
go test -v ./...

# Run a specific test
go test -run TestName ./internal/package

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Docker Commands
```bash
# Build Docker image
docker build -t momorot-tui .

# Run with Docker Compose
docker compose up

# Run in detached mode
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## Architecture Notes

### Database
- Uses SQLite for local data persistence
- Database files should be stored in the `data/` directory
- Connection handling should be managed in `internal/db/`

### UI Framework
- TUI components are located in `internal/ui/`
- Consider using popular Go TUI libraries like:
  - Bubble Tea (github.com/charmbracelet/bubbletea)
  - Termui (github.com/gizak/termui)
  - Tview (github.com/rivo/tview)

### Configuration
- Configuration management in `internal/config/`
- Support for multiple config formats (YAML, TOML, JSON)
- Environment variable overrides should be supported

### Logging
- Custom logging utilities in `internal/logx/`
- Should support different log levels (DEBUG, INFO, WARN, ERROR)
- Consider structured logging for better debugging

## Key Development Patterns

1. **Package Structure**: Follow Go's internal package pattern - packages in `internal/` cannot be imported by external projects

2. **Error Handling**: Use wrapped errors with context:
   ```go
   return fmt.Errorf("failed to connect to database: %w", err)
   ```

3. **Testing**: Each package should have corresponding test files following Go's naming convention (`*_test.go`)

4. **Dependencies**: Use Go modules for dependency management. Initialize with:
   ```bash
   go mod init github.com/username/momorot-tui
   go mod tidy
   ```

## Common Tasks

### Adding a New Feature
1. Create the package in `internal/` if it's internal logic
2. Add UI components in `internal/ui/`
3. Update configuration if needed in `internal/config/`
4. Write tests alongside the implementation
5. Update documentation

### Database Migrations
- Place migration files in `data/migrations/` or `internal/db/migrations/`
- Use a migration tool like golang-migrate or goose

### Adding Dependencies
```bash
go get package-name
go mod tidy
```

### Removing Unused Dependencies
```bash
go mod tidy
```

## Environment Variables

Common environment variables to consider:
- `MOMOROT_CONFIG_PATH` - Path to configuration file
- `MOMOROT_DATA_DIR` - Data directory location
- `MOMOROT_LOG_LEVEL` - Logging level
- `MOMOROT_DB_PATH` - SQLite database path