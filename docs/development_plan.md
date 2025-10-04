# Development Plan - Momentum Screener TUI

## Project Overview
A containerized Go TUI application that fetches daily OHLCV data, computes momentum rankings across an ETF universe, and persists everything in SQLite for speed and repeatability.

---

## Phase 1: Foundation & Setup (Days 1-3)

### 1.1 Project Initialization
- [ ] Initialize Go module: `go mod init github.com/[username]/momorot-tui`
- [ ] Set up Go 1.21 environment
- [ ] Create `.air.toml` for hot reload configuration
- [ ] Create `config/config.example.yaml` with all required fields
- [ ] Set up environment variable handling for API keys

### 1.2 Core Dependencies
```bash
# Core TUI framework
go get github.com/charmbracelet/bubbletea@v1.2.4
go get github.com/charmbracelet/bubbles@v0.20.0
go get github.com/charmbracelet/lipgloss@v1.0.0

# Database
go get modernc.org/sqlite@v1.33.1

# Configuration
go get github.com/spf13/viper@latest

# HTTP client with retry
go get github.com/hashicorp/go-retryablehttp@latest

# Testing
go get github.com/stretchr/testify@latest
```

### 1.3 Database Layer (`internal/db/`)
- [ ] Create SQLite connection manager with proper pragmas
- [ ] Implement migrations system
- [ ] Create initial schema (symbols, prices, indicators, runs, fetch_log)
- [ ] Add connection pooling and WAL mode configuration
- [ ] Create repository interfaces for each table

**Key Files:**
- `internal/db/connection.go` - Database connection and pragmas
- `internal/db/migrations.go` - Migration runner
- `internal/db/schema.sql` - Initial schema
- `internal/db/repositories.go` - Data access layer

---

## Phase 2: Data Fetching Infrastructure (Days 4-6)

### 2.1 Configuration Management (`internal/config/`)
- [ ] Implement config loader using Viper
- [ ] Support YAML files and environment variables
- [ ] Validate required fields (API key, universe)
- [ ] Set defaults for optional parameters

### 2.2 Data Fetchers (`internal/fetch/`)
- [ ] Create Alpha Vantage client with rate limiting (25 req/day)
- [ ] Implement CSV parser for Stooq manual imports
- [ ] Add request caching and delta fetching logic
- [ ] Create fetch scheduler for universe management
- [ ] Implement exponential backoff for retries

**Key Files:**
- `internal/fetch/alphavantage.go` - API client
- `internal/fetch/csv_importer.go` - CSV parsing
- `internal/fetch/scheduler.go` - Symbol scheduling
- `internal/fetch/rate_limiter.go` - Request budget management

---

## Phase 3: Analytics Engine (Days 7-10)

### 3.1 Market Calendar (`internal/analytics/calendar.go`)
- [ ] NYSE trading day calendar
- [ ] Holiday handling
- [ ] Business day calculations

### 3.2 Indicators (`internal/analytics/indicators.go`)
- [ ] Calculate multi-horizon returns (1M, 3M, 6M, 12M)
- [ ] Compute rolling volatility (63 & 126 day windows)
- [ ] Implement breadth filters (positive N-of-M lookbacks)
- [ ] Calculate average dollar volume (ADV)

### 3.3 Scoring System (`internal/analytics/scoring.go`)
- [ ] Composite momentum score with volatility penalty
- [ ] Z-score normalization
- [ ] Deterministic tie-breaking rules
- [ ] Ranking algorithm across universe

**Testing Requirements:**
- Unit tests with golden vectors for all calculations
- Deterministic output validation
- Performance benchmarks for large universes

---

## Phase 4: Terminal UI Implementation (Days 11-15)

### 4.1 Core TUI Structure (`internal/ui/`)
- [ ] Main Bubble Tea program setup
- [ ] Screen navigation system
- [ ] Keyboard shortcuts handler
- [ ] Theme and styling with Lip Gloss

### 4.2 Individual Screens

#### Dashboard Screen (`internal/ui/screens/dashboard.go`)
- [ ] Last run status display
- [ ] Cache health metrics
- [ ] API quota status
- [ ] Quick action buttons

#### Leaders Screen (`internal/ui/screens/leaders.go`)
- [ ] Top-5/Top-N ranking table
- [ ] Columns: Rank, Symbol, Score, Returns, Volatility, ADV
- [ ] Sortable columns
- [ ] Drill-down to symbol detail

#### Universe Screen (`internal/ui/screens/universe.go`)
- [ ] Full symbol list with search
- [ ] Active/inactive toggle
- [ ] Bulk operations
- [ ] Add/remove symbols

#### Symbol Detail Screen (`internal/ui/screens/symbol.go`)
- [ ] Price sparkline chart
- [ ] Return metrics table
- [ ] Volatility history
- [ ] Raw vs adjusted close comparison

#### Runs/Logs Screen (`internal/ui/screens/logs.go`)
- [ ] Run history table
- [ ] Error log viewer
- [ ] Export functionality
- [ ] Log filtering

### 4.3 Common Components (`internal/ui/components/`)
- [ ] Table widget with sorting
- [ ] Sparkline chart
- [ ] Progress spinner
- [ ] Search input
- [ ] Status bar

---

## Phase 5: Export & Reporting (Days 16-17)

### 5.1 Export Module (`internal/export/`)
- [ ] CSV export for rankings
- [ ] CSV export for full universe metrics
- [ ] Run metadata export
- [ ] Configurable output directory

### 5.2 Report Templates
- [ ] Daily leaders report
- [ ] Full ranking report
- [ ] Symbol detail report
- [ ] Run summary report

---

## Phase 6: Main Application Assembly (Days 18-19)

### 6.1 CLI Entry Point (`cmd/momo/main.go`)
- [ ] Command-line argument parsing
- [ ] Subcommands: run, refresh, export, ping
- [ ] Version information embedding
- [ ] Graceful shutdown handling

### 6.2 Application Lifecycle
- [ ] Database initialization
- [ ] Config loading
- [ ] TUI launch
- [ ] Background refresh worker
- [ ] Signal handling

---

## Phase 7: Testing & Quality Assurance (Days 20-22)

### 7.1 Unit Tests
- [ ] Math functions (returns, volatility)
- [ ] Database operations
- [ ] API client mocking
- [ ] CSV parsing

### 7.2 Integration Tests
- [ ] End-to-end data flow
- [ ] Database migrations
- [ ] API fixtures/recording
- [ ] Export verification

### 7.3 Performance Tests
- [ ] 25-symbol universe benchmark
- [ ] Database query optimization
- [ ] Memory profiling
- [ ] Concurrent fetch testing

---

## Phase 8: Containerization & Deployment (Days 23-24)

### 8.1 Docker Setup
- [ ] Multi-stage Dockerfile optimization
- [ ] Distroless runtime image
- [ ] Volume configuration for data persistence
- [ ] Security hardening (non-root, read-only FS)

### 8.2 Docker Compose
- [ ] Development environment
- [ ] Production configuration
- [ ] Health checks
- [ ] Log aggregation

---

## Phase 9: Documentation & Polish (Days 25-26)

### 9.1 User Documentation
- [ ] README.md with quickstart
- [ ] Configuration guide
- [ ] API key setup instructions
- [ ] Docker usage guide

### 9.2 Developer Documentation
- [ ] Architecture overview
- [ ] Contributing guidelines
- [ ] API documentation
- [ ] Testing guide

---

## Phase 10: Production Readiness (Days 27-28)

### 10.1 Operational Features
- [ ] Health check endpoint
- [ ] Metrics collection
- [ ] Structured logging
- [ ] Backup automation

### 10.2 Final Validation
- [ ] Acceptance criteria verification
- [ ] Performance benchmarks
- [ ] Security audit
- [ ] Docker image optimization

---

## Technical Milestones

### Milestone 1: Data Layer Complete (End of Phase 3)
- SQLite database operational
- Data fetching from Alpha Vantage working
- Analytics calculations validated

### Milestone 2: TUI Functional (End of Phase 4)
- All screens navigable
- Real-time data display
- User interactions working

### Milestone 3: MVP Complete (End of Phase 6)
- Full application flow working
- Data persistence functional
- Export capabilities ready

### Milestone 4: Production Ready (End of Phase 10)
- Containerized and deployable
- Fully tested and documented
- Performance optimized

---

## Risk Mitigation

### Technical Risks
1. **API Rate Limits**: Implement aggressive caching, delta fetching
2. **SQLite Performance**: Use WAL mode, proper indexing
3. **TUI Complexity**: Start with simple screens, iterate
4. **Container Size**: Use multi-stage builds, distroless images

### Schedule Risks
1. **Buffer Time**: Each phase includes 20% buffer
2. **Parallel Work**: Some phases can overlap
3. **Early Testing**: Write tests alongside implementation

---

## Success Criteria

- [ ] Handles 25-symbol universe within API limits
- [ ] Deterministic rankings given same data
- [ ] TUI responsive and navigable
- [ ] Docker image < 30MB
- [ ] Full refresh < 10s (excluding network)
- [ ] 80%+ test coverage
- [ ] Zero critical security issues

---

## Next Steps

1. **Immediate**: Set up Go module and install dependencies
2. **Day 1**: Create database schema and connection manager
3. **Day 2**: Implement configuration system
4. **Day 3**: Begin Alpha Vantage client development

---

## Notes

- All dates assume full-time development
- Phases can be adjusted based on progress
- Some components can be developed in parallel
- Regular testing throughout all phases
- Daily commits to track progress