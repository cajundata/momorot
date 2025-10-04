# Momorot TUI

A Terminal User Interface (TUI) application built with Go for managing and analyzing data through an intuitive command-line interface.

## Overview

Momorot TUI provides a powerful terminal-based interface for data management, analytics, and visualization. Built with Go for performance and reliability, it offers a modern CLI experience with rich interactive features.

## Features

- 🖥️ **Interactive Terminal UI** - Clean and intuitive terminal interface
- 📊 **Analytics & Metrics** - Built-in analytics and metrics functionality
- 💾 **Local Data Storage** - SQLite-based local data persistence
- 📤 **Data Export** - Export data in multiple formats
- ⚙️ **Configurable** - Flexible configuration management
- 🚀 **Cross-Platform** - Works on Linux, macOS, and Windows

## Project Structure

```
momorot-tui/
├── cmd/momo/           # Main application entry point
├── internal/           # Internal packages
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

## Prerequisites

- Go 1.23 or higher (developed with Go 1.25.1)
- gcc/build tools (for SQLite compilation)

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/cajundata/momorot.git
cd momorot-tui
```

2. Build the application:
```bash
go build -o momorot-tui ./cmd/momo
```

3. (Optional) Install globally:
```bash
go install ./cmd/momo
```

## Usage

### Basic Usage

Run the application:
```bash
./momorot-tui
```

Or if installed globally:
```bash
momo
```

### Configuration

The application supports multiple configuration methods:

- Configuration files in `configs/` directory
- Environment variables (prefix: `MOMOROT_`)
- Command-line flags

Example environment variables:
```bash
export MOMOROT_CONFIG_PATH=/path/to/config
export MOMOROT_DATA_DIR=/path/to/data
export MOMOROT_LOG_LEVEL=debug
```

## Development

### Running in Development

```bash
# Run with hot reload (requires air)
air

# Or run directly
go run ./cmd/momo
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### Building

```bash
# Build for current platform
go build -o momorot-tui ./cmd/momo

# Build with version information
go build -ldflags "-X main.version=$(git describe --tags --always)" -o momorot-tui ./cmd/momo

# Cross-compile for different platforms
# Linux
GOOS=linux GOARCH=amd64 go build -o momorot-tui-linux-amd64 ./cmd/momo

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o momorot-tui-darwin-amd64 ./cmd/momo

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o momorot-tui-darwin-arm64 ./cmd/momo

# Windows
GOOS=windows GOARCH=amd64 go build -o momorot-tui.exe ./cmd/momo
```

The beauty of Go is that it compiles to a single static binary - no runtime dependencies needed! Just build and run.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Core TUI framework implementation
- [ ] Database schema and migrations
- [ ] Basic CRUD operations
- [ ] Analytics dashboard
- [ ] Export functionality (CSV, JSON, Excel)
- [ ] Configuration management
- [ ] Plugin system
- [ ] API integration
- [ ] Advanced visualization features
- [ ] Performance optimizations

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

Built with:
- [Go](https://golang.org/) - The Go Programming Language
- [SQLite](https://www.sqlite.org/) - Self-contained SQL database engine
- TUI Framework (TBD - Bubble Tea, Termui, or Tview)

## Contact

For questions, issues, or suggestions, please open an issue on GitHub.

---

**Note:** This project is currently under active development. Features and APIs may change.