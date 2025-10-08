# Development Plan - Momentum Screener TUI

## Project Overview
A containerized Go TUI application that fetches daily OHLCV data, computes momentum rankings across an ETF universe, and persists everything in SQLite for speed and repeatability.

## Overall Progress

**Phase Completion Status:**
- ✅ **Phase 1: Foundation & Setup** - COMPLETE (100%)
- ✅ **Phase 2: Data Fetching Infrastructure** - COMPLETE (100%)
- ✅ **Phase 3: Analytics Engine** - COMPLETE (100%)
- ✅ **Phase 4: Terminal UI Implementation** - COMPLETE (100%)
- ⏳ **Phase 5: Export & Reporting** - PENDING
- ⏳ **Phase 6: Main Application Assembly** - PENDING
- ⏳ **Phase 7: Testing & Quality Assurance** - PENDING
- ⏳ **Phase 8: Containerization & Deployment** - PENDING
- ⏳ **Phase 9: Documentation & Polish** - PENDING
- ⏳ **Phase 10: Production Readiness** - PENDING

**Key Metrics:**
- **Total Tests**: 239 (all passing)
- **Test Coverage**: Analytics 66.1%, Config 82.1%, DB 82.0%, Fetch 47.4%, UI 57.0%, Screens 88.3%
- **Lines of Code**: ~6,400 (excluding tests)
- **Files Implemented**: 40 files (17 UI files: 7 core + 5 screens + 4 components + 1 test helper)
- **Current Milestone**: Phase 4 COMPLETE - Full TUI with 5 integrated screens, navigation, and 88.3% test coverage

---

## Phase 1: Foundation & Setup (Days 1-3)

**Note**: Placeholder directories created during setup: `internal/export/`, `internal/logx/`, `internal/version/` (currently empty, will be populated in later phases).

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

**Note:** Bubble Tea v1.3.4 is used instead of v1.3.10 for compatibility with Bubbles v0.21.0 (latest stable release). Lip Gloss v1.1.0 is used (v1.0.0 not available, fully compatible).

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
- [x] NYSE trading day calendar with 2024-2030 holidays
- [x] Holiday handling (10 major holidays per year)
- [x] Business day calculations (Next, Previous, Add, Count)

### 3.2 Indicators (`internal/analytics/indicators.go`)
- [x] Calculate multi-horizon total returns using `adj_close` (1M=21d, 3M=63d, 6M=126d, 12M=252d)
- [x] Compute rolling volatility: σ of daily log returns over 63/126 day windows (annualized)
- [x] Implement breadth filters (require positive N-of-M lookbacks)
- [x] Calculate average dollar volume (ADV) with minimum threshold ($5M default)

### 3.3 Scoring System (`internal/analytics/scoring.go`)
- [x] Composite momentum score: average return minus λ·volatility
- [x] Volatility penalty parameter λ = 0.35 (configurable)
- [x] Z-score normalization across universe
- [x] Deterministic tie-breaking: (1) higher score, (2) lower volatility, (3) higher liquidity, (4) alphabetical
- [x] Ranking algorithm producing deterministic, reproducible results

### 3.4 Orchestrator (`internal/analytics/orchestrator.go`)
- [x] Database integration for reading prices and writing indicators
- [x] Batch processing for all active symbols
- [x] End-to-end indicator computation and ranking pipeline

**Implemented Files:**
- `internal/analytics/types.go` - Shared data structures (PriceBar, Indicators, SymbolScore)
- `internal/analytics/calendar.go` - NYSE calendar with holiday detection
- `internal/analytics/calendar_test.go` - Calendar tests (11 tests, all passing)
- `internal/analytics/indicators.go` - Multi-horizon returns, volatility, ADV calculations
- `internal/analytics/indicators_test.go` - Indicators tests with golden vectors (20 tests)
- `internal/analytics/scoring.go` - Z-score normalization and deterministic ranking
- `internal/analytics/scoring_test.go` - Scoring tests with determinism validation (17 tests)
- `internal/analytics/orchestrator.go` - Database-integrated analytics pipeline

**Phase 3 Test Summary:**
- 61 analytics tests, all passing
- Analytics coverage: 66.1% of statements
- Golden vectors validated for all mathematical calculations
- Deterministic ordering verified (same input → same output)
- Performance benchmarks included

**Overall Test Summary (Phases 1-3):**
- Total: 112 tests (excluding UI tests), all passing
- DB: 82.0% coverage
- Config: 82.1% coverage
- Analytics: 66.1% coverage
- Fetch: 47.4% coverage

---

## Phase 4: Terminal UI Implementation (Days 11-15)

### 4.1 Core TUI Structure (`internal/ui/`) ✅ **COMPLETE (100%)**
- [x] Main Bubble Tea v1.3.4 program setup (stable v1.x, NOT v2 beta)
- [x] Screen navigation system with `←/→` tab navigation
- [x] Keyboard shortcuts: `r` refresh, `/` search, `e` export CSV, `q` quit
- [x] Theme and styling with Lip Gloss v1.1.0
- [x] Core model, update, and view architecture
- [x] Comprehensive key binding system with help text
- [x] Window size handling and responsive layout
- [x] Status bar with loading, error, and help states
- [x] Tab-based screen navigation framework
- [x] Core TUI tests (13 tests, 57.0% coverage)

**Implemented Files (Core Foundation - Phase 4.1 COMPLETE):**
- `internal/ui/model.go` - Main Bubble Tea model with screen state, navigation, and dependencies
- `internal/ui/update.go` - Update function with message routing and screen-specific handlers
- `internal/ui/view.go` - View rendering with header, tabs, status bar, and screen routing
- `internal/ui/keys.go` - Full key binding system with help text generation
- `internal/ui/theme.go` - Comprehensive theming system with color palettes and styles
- `internal/ui/deps.go` - Dependency placeholder to ensure Bubble Tea packages stay in go.mod
- `internal/ui/model_test.go` - Core TUI tests (13 tests, 57.0% coverage, all passing)

**Key Features Implemented (Phase 4.1):**
- ✅ Screen enum and navigation state machine (Dashboard, Leaders, Universe, Symbol, Logs)
- ✅ Forward/backward navigation with history stack
- ✅ Global key bindings using Bubble Tea's key package
- ✅ Theme system with color palette and pre-built styles
- ✅ Status bar with loading, error, and help text states
- ✅ Tab-based screen navigation with arrow keys
- ✅ Window size handling and responsive layout
- ✅ Full test coverage for core model operations

### 4.2 Individual Screens ✅ **COMPLETE (100%)**

#### Dashboard Screen (`internal/ui/screens/dashboard.go`) ✅
- [x] Create screens directory and dashboard file
- [x] Last run status display (from `runs` table)
- [x] Cache health metrics (last fetched date per symbol)
- [x] API quota status (25 req/day budget, next reset)
- [x] Quick action buttons
- **Tests**: 18 tests, 100% passing
- **Test file**: `internal/ui/screens/dashboard_test.go`

#### Leaders Screen (`internal/ui/screens/leaders.go`) ✅
- [x] Top-5/Top-N ranking table
- [x] Columns: Rank, Symbol, Score, R1M, R3M, R6M, Vol, ADV
- [x] Data from `indicators` table joined with `prices`
- [x] Drill-down to symbol detail
- **Tests**: 18 tests, 100% passing
- **Test file**: `internal/ui/screens/leaders_test.go`

#### Universe Screen (`internal/ui/screens/universe.go`) ✅
- [x] Full symbol list with `/` search
- [x] Active/inactive toggle (updates `symbols.active` column)
- [x] Display asset_type (ETF/STOCK/INDEX)
- [x] Search mode toggle with keyboard navigation
- **Tests**: 20 tests, 100% passing (85.4% coverage)
- **Test file**: `internal/ui/screens/universe_test.go`

#### Symbol Detail Screen (`internal/ui/screens/symbol.go`) ✅
- [x] Price sparkline chart (ASCII visualization)
- [x] Return metrics cards (R1M, R3M, R6M, R12M)
- [x] Volatility history (3M, 6M windows)
- [x] ADV and momentum score display
- **Tests**: 20 tests, 100% passing (87.0% coverage)
- **Test file**: `internal/ui/screens/symbol_test.go`

#### Runs/Logs Screen (`internal/ui/screens/logs.go`) ✅
- [x] Run history table (from `runs` table)
- [x] Error log viewer (from `fetch_log` where ok=0)
- [x] Tab-switching focus between tables
- [x] Filter by status (OK/ERROR/RUNNING)
- **Tests**: 21 tests, 100% passing (88.3% coverage)
- **Test file**: `internal/ui/screens/logs_test.go`

**Phase 4.2 Summary:**
- **Total Tests**: 97 screen tests (all passing)
- **Average Coverage**: 88.3%
- **Lines of Code**: ~2,800 (screens + tests)
- **Completion**: All 5 screens implemented and tested

### 4.3 Common Components (`internal/ui/components/`) ✅ **COMPLETE (100%)**

- [x] Create components directory
- [x] Table widget (`table.go`) - wrapper around Bubbles table with focus/blur
- [x] Sparkline chart (`sparkline.go`) - ASCII chart with 8-level block characters
- [x] Search input (`search.go`) - wrapper around Bubbles textinput with theming
- [x] Status bar (integrated into main Model)

**Implemented Components:**
- `internal/ui/components/table.go` - Table wrapper with custom styling and focus states
- `internal/ui/components/sparkline.go` - ASCII price charts with statistics
- `internal/ui/components/search.go` - Search input with Focus/Blur/Reset methods

**Note**: Components are tested indirectly through screen integration tests (88.3% overall coverage)

### 4.4 Main Model Integration ✅ **COMPLETE (100%)**

- [x] Update main Model to use concrete screen types instead of interface
- [x] Replace placeholder interface{} with typed screens
- [x] Implement screen navigation state machine
- [x] Wire up screen switching logic (Tab for next, Shift+Tab for prev)
- [x] Pass database connection to screen constructors
- [x] Handle screen initialization and data loading
- [x] Implement screen-to-screen navigation (Leaders/Universe → Symbol Detail)
- [x] Add drill-down navigation with back/breadcrumb support (Esc key)
- [x] Update status bar to show current screen name
- [x] Test full navigation flow between all screens

**Implemented Changes:**
- `internal/ui/model.go` - Concrete screen types, NavigateToSymbol(), NavigateToSymbolMsg
- `internal/ui/update.go` - Screen message delegation, WindowSizeMsg forwarding
- `internal/ui/view.go` - Screen View() delegation, getScreenName() helper
- `internal/ui/model_test.go` - Updated Init() test for batch commands

**Phase 4.4 Summary:**
- All screens fully integrated into main Model
- Tab-based navigation between 5 screens
- Drill-down navigation with back support
- Status bar shows "[Screen Name]" indicator
- 239 total tests passing (all UI tests green)

---

## Phase 5: Export & Reporting (Days 16-17)

**Note**: `internal/export/` directory exists but is empty (placeholder from initial setup).

### 5.1 Export Module (`internal/export/`)
- [ ] Implement CSV export for Top-5 leaders (leaders-YYYYMMDD.csv)
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

### Milestone 1: Data Layer Complete (End of Phase 3) ✅ **COMPLETED**
- ✅ SQLite database operational with WAL mode
- ✅ Data fetching from Alpha Vantage working with rate limiting
- ✅ Analytics calculations validated with golden vectors
- ✅ 112 tests passing (Phases 1-3, excluding UI)
- ✅ Deterministic ranking algorithm verified
- ✅ Test coverage: DB 82.0%, Config 82.1%, Analytics 66.1%, Fetch 47.4%

**Status**: Phases 1-3 fully implemented and tested. All core data infrastructure complete.

### Milestone 2: TUI Functional (End of Phase 4) 🟡 **IN PROGRESS (~35%)**
- ✅ Phase 4.1: Core TUI framework complete (100%)
  - ✅ Bubble Tea v1.3.4 program setup
  - ✅ Navigation system with tab switching
  - ✅ Key bindings and theme system
  - ✅ Model, update, view architecture
  - ✅ 13 UI tests (57.0% coverage)
- ⏳ Phase 4.2: Individual screens (0/5 complete)
  - ⏳ Dashboard screen
  - ⏳ Leaders screen
  - ⏳ Universe screen
  - ⏳ Symbol detail screen
  - ⏳ Runs/logs screen
- ⏳ Phase 4.3: Common components (0/5 complete)
  - ⏳ Table widget
  - ⏳ Sparkline chart
  - ⏳ Progress spinner
  - ⏳ Search input
  - ⏳ Status bar component

**Status**: Phase 4.1 complete (core TUI foundation). Ready to implement screens and components (Phases 4.2-4.3).

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

- [x] Handles 25-symbol universe within Alpha Vantage free tier (25 req/day)
- [x] Deterministic, byte-identical rankings given same data and config
- [ ] TUI responsive and navigable with all screens functional (4.1 ✅ complete, 4.2-4.3 ⏳ pending)
- [ ] Docker image ≤ 20-30 MB (distroless + static binary)
- [ ] Full delta refresh < 10s for 25 symbols (excluding network I/O)
- [x] High test coverage with golden vectors (currently 125 tests passing, 47-82% coverage across modules)
- [ ] Zero critical security issues (secrets, permissions, container hardening)
- [x] Pure Go build (CGO_ENABLED=0) for portability
- [x] SQLite with WAL mode, STRICT tables, proper indexes

---

## Next Steps

### Current Phase: Phase 4 - Terminal UI Implementation

**Completed:**
1. ✅ Phase 4.1: Core TUI structure complete (100%)
   - ✅ Model, update, view architecture
   - ✅ Key bindings system with help text
   - ✅ Theme and styling system with Lip Gloss
   - ✅ Screen navigation framework
   - ✅ Status bar and window handling
   - ✅ Core TUI tests (13 tests passing)

**Next Steps (Phase 4.2 & 4.3):**
2. ⏳ Create `internal/ui/components/` directory and build reusable components
   - ⏳ Table widget (using Bubbles table)
   - ⏳ Sparkline chart component
   - ⏳ Progress spinner component
   - ⏳ Search input component
   - ⏳ Status bar component enhancements
3. ⏳ Create `internal/ui/screens/` directory and implement individual screens
   - ⏳ Dashboard screen with run status and metrics
   - ⏳ Leaders screen with Top-N rankings table
   - ⏳ Universe screen with symbol management
   - ⏳ Symbol Detail screen with sparkline and metrics
   - ⏳ Runs/Logs screen with history viewer
4. ⏳ Integrate database queries with all screens
5. ⏳ Write tests for all screens and components
6. ⏳ End-to-end TUI testing and bug fixes

---

## Notes

- All dates assume full-time development
- Phases can be adjusted based on progress
- Some components can be developed in parallel
- Regular testing throughout all phases (aim for 80%+ coverage)
- Daily commits to track progress
- **Tech stack pinned to stable versions**: Go 1.25.1, Bubble Tea v1.3.4 (NOT v2 beta), Bubbles v0.21.0, Lip Gloss v1.1.0, modernc.org/sqlite v1.39.0
- **Container stack**: Docker Engine 28.x, Compose v2.40.0, distroless runtime
- **Critical**: Pure Go build (CGO_ENABLED=0) for cross-platform compatibility and minimal container size