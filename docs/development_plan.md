# Development Plan - Momentum Screener TUI

## Project Overview
A containerized Go TUI application that fetches daily OHLCV data, computes momentum rankings across an ETF universe, and persists everything in SQLite for speed and repeatability.

---

## Phase 1: Foundation & Setup (Days 1-3)

### 1.1 Project Initialization
- [x] Initialize Go module: `go mod init github.com/malbwa/momorot`
- [x] Set up Go 1.25.1 environment
- [x] Create `.air.toml` for hot reload configuration
- [x] Create `configs/config.example.yaml` with all required fields
- [x] Set up `.env.example` for environment variable handling
- [x] Update `.gitignore` to exclude `config.yaml`, `.env`, `data/`, `exports/`

### 1.2 Core Dependencies
- [x] Install Bubble Tea v1.3.4 (TUI framework - stable v1.x series)
- [x] Install Bubbles v0.21.0 (TUI components)
- [x] Install Lip Gloss v1.0.0 (styling)
- [x] Install modernc.org/sqlite v1.39.0 (pure Go, cgo-free database)
- [x] Install Viper v1.21.0 (configuration management)
- [x] Install go-retryablehttp v0.7.8 (HTTP client with retry)
- [x] Install testify v1.11.1 (testing framework)
- [x] Run go mod tidy to organize dependencies

**Installed versions:**
```bash
# Core TUI framework (stable v1.x series)
github.com/charmbracelet/bubbletea v1.3.4
github.com/charmbracelet/bubbles v0.21.0
github.com/charmbracelet/lipgloss v1.0.0

# Database (pure Go, cgo-free)
modernc.org/sqlite v1.39.0

# Configuration
github.com/spf13/viper v1.21.0

# HTTP client with retry
github.com/hashicorp/go-retryablehttp v0.7.8

# Testing
github.com/stretchr/testify v1.11.1
```

**Note:** Bubble Tea v1.3.4 is used instead of v1.3.10 for compatibility with Bubbles v0.21.0 (latest stable release).

### 1.3 Database Layer (`internal/db/`)
- [x] Create SQLite connection manager with proper pragmas (WAL mode, STRICT tables)
- [x] Implement migrations system with version tracking
- [x] Create initial schema (symbols, prices, indicators, runs, fetch_log)
- [x] Configure pragmas: `journal_mode=WAL`, `synchronous=NORMAL`, `foreign_keys=ON`, `busy_timeout=5000`, `temp_store=MEMORY`
- [x] Create repository interfaces for each table (Symbol, Price, Indicator, Run, FetchLog)
- [x] Add indexes: `idx_prices_symbol_date`, `idx_indicators_date_rank`, `idx_symbols_active`, `idx_runs_status_date`, `idx_fetch_log_failures`
- [x] Write comprehensive unit tests (20 test cases, all passing)

**Implemented Files:**
- `internal/db/connection.go` - Database connection manager with WAL mode and optimized pragmas
- `internal/db/migrations.go` - Migration system with up/down support and version tracking
- `internal/db/schema.sql` - STRICT tables schema with foreign keys and indexes
- `internal/db/repositories.go` - Repository pattern for all tables with batch operations
- `internal/db/connection_test.go` - Connection and pragma tests
- `internal/db/migrations_test.go` - Migration and rollback tests
- `internal/db/repositories_test.go` - Repository operation tests

---

## Phase 2: Data Fetching Infrastructure (Days 4-6)

### 2.1 Configuration Management (`internal/config/`)
- [x] Implement config loader using Viper
- [x] Support YAML files and environment variables
- [x] Validate required fields (API key, universe)
- [x] Set defaults for optional parameters

### 2.2 Data Fetchers (`internal/fetch/`)
- [x] Create Alpha Vantage client with rate limiting (strict 25 req/day free tier)
- [x] Implement `TIME_SERIES_DAILY_ADJUSTED` endpoint for OHLCV + adjusted close
- [x] Implement CSV parser for Stooq manual imports (bootstrap/fallback)
- [x] Create fetch scheduler for universe management (stagger large universes)
- [x] Implement rate limiter with friendly quota messages
- [x] Use controlled worker pool to respect API budget
- [x] Write comprehensive tests for all Phase 2 components (30 tests, all passing)

**Implemented Files:**
- `internal/config/config.go` - Viper-based config loader with validation and defaults
- `internal/config/config_test.go` - Config loader tests (13 tests)
- `internal/fetch/alphavantage.go` - API client with 25/day limit enforcement
- `internal/fetch/csv_importer.go` - Stooq CSV parsing with validation
- `internal/fetch/scheduler.go` - Symbol scheduling and prioritization
- `internal/fetch/rate_limiter.go` - Request budget management with auto-reset
- `internal/fetch/rate_limiter_test.go` - Rate limiter tests (8 tests)
- `internal/fetch/csv_importer_test.go` - CSV importer tests (15 tests)
- `internal/fetch/scheduler_test.go` - Scheduler tests (7 tests)

---

## Phase 3: Analytics Engine (Days 7-10)

### 3.1 Market Calendar (`internal/analytics/calendar.go`)
- [ ] NYSE trading day calendar
- [ ] Holiday handling
- [ ] Business day calculations

### 3.2 Indicators (`internal/analytics/indicators.go`)
- [ ] Calculate multi-horizon total returns using `adj_close` (1M=21d, 3M=63d, 6M=126d, 12M=252d)
- [ ] Compute rolling volatility: σ of daily log returns over 63/126 day windows
- [ ] Implement breadth filters (require positive N-of-M lookbacks)
- [ ] Calculate average dollar volume (ADV) with minimum threshold ($5M default)

### 3.3 Scoring System (`internal/analytics/scoring.go`)
- [ ] Composite momentum score: z-score or min-max of lookback returns minus λ·volatility
- [ ] Volatility penalty parameter λ = 0.35 (configurable)
- [ ] Z-score normalization across universe
- [ ] Deterministic tie-breaking: lower volatility first, then higher liquidity (ADV)
- [ ] Ranking algorithm producing deterministic, reproducible results

**Testing Requirements:**
- Unit tests with golden vectors for all calculations
- Deterministic output validation (byte-identical given same inputs)
- Performance benchmarks for 25-symbol universe
- Idempotent re-run verification

---

## Phase 4: Terminal UI Implementation (Days 11-15)

### 4.1 Core TUI Structure (`internal/ui/`)
- [ ] Main Bubble Tea v1.3.4 program setup (stable v1.x, NOT v2 beta)
- [ ] Screen navigation system with `←/→` tab navigation
- [ ] Keyboard shortcuts: `r` refresh, `/` search, `e` export CSV, `q` quit
- [ ] Theme and styling with Lip Gloss v1.0.0
- [ ] Non-blocking refresh with spinner (use Bubbles v0.21.0 spinner)

### 4.2 Individual Screens

#### Dashboard Screen (`internal/ui/screens/dashboard.go`)
- [ ] Last run status display (from `runs` table)
- [ ] Cache health metrics (last fetched date per symbol)
- [ ] API quota status (25 req/day budget, next reset)
- [ ] Quick action buttons

#### Leaders Screen (`internal/ui/screens/leaders.go`)
- [ ] Top-5/Top-N ranking table
- [ ] Columns: Rank, Symbol, Score, R1M, R3M, R6M, Vol, ADV
- [ ] Data from `indicators` table joined with `prices`
- [ ] Drill-down to symbol detail

#### Universe Screen (`internal/ui/screens/universe.go`)
- [ ] Full symbol list with `/` search
- [ ] Active/inactive toggle (updates `symbols.active` column)
- [ ] Display asset_type (ETF/STOCK/INDEX)
- [ ] Add/remove symbols

#### Symbol Detail Screen (`internal/ui/screens/symbol.go`)
- [ ] Price sparkline chart (using Bubbles components)
- [ ] Return metrics table (R1M, R3M, R6M, R12M)
- [ ] Volatility history (3M, 6M windows)
- [ ] Raw vs adjusted close comparison

#### Runs/Logs Screen (`internal/ui/screens/logs.go`)
- [ ] Run history table (from `runs` table)
- [ ] Error log viewer (from `fetch_log` where ok=0)
- [ ] Export logs functionality
- [ ] Filter by run_id, symbol, status

### 4.3 Common Components (`internal/ui/components/`)
- [ ] Table widget with sorting (using Bubbles v0.21.0 table)
- [ ] Sparkline chart (custom or Bubbles component)
- [ ] Progress spinner (Bubbles spinner for non-blocking refresh)
- [ ] Search input (Bubbles textinput)
- [ ] Status bar with help text

---

## Phase 5: Export & Reporting (Days 16-17)

### 5.1 Export Module (`internal/export/`)
- [ ] CSV export for Top-5 leaders (leaders-YYYYMMDD.csv)
- [ ] CSV export for full universe rankings
- [ ] Run metadata export (runs.csv)
- [ ] Output to `./exports` or configurable `export_dir`
- [ ] Auto-generate on successful refresh

### 5.2 Report Templates
- [ ] Daily leaders report: Rank, Symbol, Score, R1M, R3M, R6M, Vol, ADV
- [ ] Full ranking report: all symbols with indicators
- [ ] Symbol detail report: full time series and metrics
- [ ] Run summary report: status, timing, symbols fetched

---

## Phase 6: Main Application Assembly (Days 18-19)

### 6.1 CLI Entry Point (`cmd/momo/main.go`)
- [ ] Command-line argument parsing (flags package or cobra)
- [ ] Subcommands: `run` (TUI), `refresh`, `export`, `ping` (for healthcheck)
- [ ] Version information embedding via ldflags
- [ ] Graceful shutdown handling (SIGINT, SIGTERM)

### 6.2 Application Lifecycle
- [ ] Database initialization at `./data/momentum.db` (or configurable `data_dir`)
- [ ] Config loading from `config.yaml` or env vars (ALPHAVANTAGE_API_KEY)
- [ ] TUI launch with Bubble Tea program
- [ ] Background refresh worker with controlled concurrency
- [ ] Signal handling and cleanup (flush DB, close connections)

---

## Phase 7: Testing & Quality Assurance (Days 20-22)

### 7.1 Unit Tests
- [ ] Math functions (returns, volatility)
- [ ] Database operations
- [ ] API client mocking
- [ ] CSV parsing

### 7.2 Integration Tests
- [ ] End-to-end data flow (fetch → compute → display → export)
- [ ] Database migrations up/down idempotency
- [ ] API fixtures/recording for Alpha Vantage (record/replay)
- [ ] CSV import from Stooq fixtures
- [ ] Export verification (CSV format and content)

### 7.3 Performance Tests
- [ ] 25-symbol universe full refresh < 10s (excluding network)
- [ ] Database query optimization with EXPLAIN QUERY PLAN
- [ ] Memory profiling (target: minimal footprint)
- [ ] Concurrent fetch testing with worker pool
- [ ] Determinism test: byte-identical output for fixed DB snapshot

---

## Phase 8: Containerization & Deployment (Days 23-24)

### 8.1 Docker Setup
- [ ] Multi-stage Dockerfile using `golang:1.25.1` as builder
- [ ] Runtime image: `gcr.io/distroless/static:nonroot`
- [ ] Build static binary with `CGO_ENABLED=0` (pure Go)
- [ ] Volume configuration for `/data` persistence
- [ ] Security hardening: non-root user (65532), read-only FS, `cap_drop: ALL`
- [ ] Target image size: ≤ 20-30 MB

### 8.2 Docker Compose
- [ ] Use Docker Compose v2.40.0 syntax (no `version:` field)
- [ ] Development environment configuration
- [ ] Production configuration
- [ ] Health check using `--ping` subcommand
- [ ] Structured JSON logging output
- [ ] Docker Engine 28.x compatibility

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
- [ ] Health check: `--ping` subcommand exits 0 if healthy
- [ ] Structured JSON logging with run_id, symbol, action, outcome, latency
- [ ] Backup automation: SQLite VACUUM INTO or optional Litestream
- [ ] Friendly error messages for quota exhaustion

### 10.2 Final Validation
- [ ] Acceptance criteria verification from requirements doc
- [ ] Performance benchmarks (25-symbol < 10s)
- [ ] Security audit: no secrets committed, read-only FS, non-root
- [ ] Docker image optimization: verify ≤ 20-30 MB final size
- [ ] Cross-platform build: linux/amd64 and linux/arm64

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

- [ ] Handles 25-symbol universe within Alpha Vantage free tier (25 req/day)
- [ ] Deterministic, byte-identical rankings given same data and config
- [ ] TUI responsive and navigable with all screens functional
- [ ] Docker image ≤ 20-30 MB (distroless + static binary)
- [ ] Full delta refresh < 10s for 25 symbols (excluding network I/O)
- [ ] 80%+ test coverage with golden vectors
- [ ] Zero critical security issues (secrets, permissions, container hardening)
- [ ] Pure Go build (CGO_ENABLED=0) for portability
- [ ] SQLite with WAL mode, STRICT tables, proper indexes

---

## Next Steps

1. **Immediate**: Initialize Go 1.25.1 module and install core dependencies
2. **Day 1**: Create SQLite schema with STRICT tables, WAL pragmas, and migrations
3. **Day 2**: Implement configuration system (Viper + env vars)
4. **Day 3**: Begin Alpha Vantage client with 25 req/day rate limiter
5. **Day 4**: Set up Bubble Tea v1.3.10 TUI skeleton with navigation

---

## Notes

- All dates assume full-time development
- Phases can be adjusted based on progress
- Some components can be developed in parallel
- Regular testing throughout all phases (aim for 80%+ coverage)
- Daily commits to track progress
- **Tech stack pinned to stable versions**: Go 1.25.1, Bubble Tea v1.3.4 (NOT v2 beta), Bubbles v0.21.0, Lip Gloss v1.0.0, modernc.org/sqlite v1.39.0
- **Container stack**: Docker Engine 28.x, Compose v2.40.0, distroless runtime
- **Critical**: Pure Go build (CGO_ENABLED=0) for cross-platform compatibility and minimal container size