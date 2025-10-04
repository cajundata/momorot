# Momentum Screener TUI — Minimal Software Requirements (v0.2, updated Oct 3, 2025)

> A tiny, containerized Go TUI that fetches free daily OHLCV, computes momentum ranks across an ETF universe, and stores everything in **SQLite** for speed and repeatability.

---

## 1) Goals & Scope

**Primary goals**
- Pull historical daily prices (close/adjusted close, volume) for a configurable list of tickers.
- Compute the momentum ranking system we defined (multi-horizon returns, volatility penalty, breadth filter, tie-breaks).
- Present results in a fast terminal UI with simple navigation, search, and export.
- Persist all raw series, derived indicators, and run metadata to **SQLite**.
- Run reproducibly in Docker on Linux/amd64 and arm64.

**Out of scope (initial)**
- Real-time or intraday data.
- Brokerage connectivity or order routing.
- Option chains or Greeks.
- Backtesting engine; only rolling, point-in-time screen.

---

## 2) Users & Workflows

**Personas**
- *You (power user)*: kicks off a daily refresh, reviews Top-5, inspects constituents, exports CSV.
- *Future “batch” user*: schedules via cron/Kubernetes to refresh nightly and publish an artifact.

**Happy path**
1. Launch TUI → last run status and cache health.
2. Press `r` (Refresh) → app fetches missing daily bars (respects API limits), updates indicators/scores.
3. View *Leaders* screen → Top-5 with detail; drill into ticker page for series & diagnostics.
4. Export Top-5 and full ranks as CSV.

---

## 3) Data Sources (Free Tier)

- **Primary**: Alpha Vantage `TIME_SERIES_DAILY_ADJUSTED` for equities/ETFs; free plan is 25 requests/day (daily batch fits if we cache aggressively and batch universes).
- **Bootstrap / Manual fallback**: Stooq downloadable CSVs from each symbol’s *Historical data* page (“Download data in .csv file…”). (No official API; manual/scripted download only.)

**Implications**
- Tight request budget → must cache and only fetch *deltas* (latest missing days) per symbol.
- Universe size in free mode should be ≤25 symbols/day (or staggered schedule).

---

## 4) Tech Stack (pin to current stable)

- **Go**: **1.25.1** (latest stable).
- **TUI framework**: **Bubble Tea v1.3.10** (latest stable); v2 is beta, not used for prod yet.
- **TUI components**: **Bubbles v0.21.0** (latest release).
- **Styling**: **Lip Gloss v1.0.0**.
- **DB (embedded)**: **SQLite** via **modernc.org/sqlite v1.39.0** (pure Go, cgo-free).  
  - Upstream SQLite current series is **3.50.x**.
- **Containers**
  - **Docker Engine**: 28.x GA track (current series).
  - **Docker Compose**: **v2.40.0** (latest release).
  - **Runtime image**: `gcr.io/distroless/static:nonroot`.

---

## 5) Architecture

**Process model**
- Single process binary with three subsystems:
  1. **Fetcher**: HTTP client + rate limiter; symbol scheduler; CSV importers.
  2. **Analytics**: indicator/score calculators (pure Go); deterministic, idempotent.
  3. **TUI**: Bubble Tea program; screens: Dashboard, Leaders, Universe, Symbol, Runs/Logs.

**Storage**
- One SQLite file at `./data/momentum.db` (mounted volume in Docker).

**Concurrency**
- Controlled worker pool for network calls (respect request budget).
- DB access via `database/sql` with WAL enabled (read-mostly TUI + background writes).

---

## 6) SQLite Details (chosen store)

**Why SQLite here**
- Single binary, zero ops, strong SQL, transactional; cgo-free driver keeps Docker minimal.

**DB pragmas (on open)**
```sql
PRAGMA journal_mode=WAL;         -- concurrent readers
PRAGMA synchronous=NORMAL;       -- durability/speed trade-off for WAL
PRAGMA foreign_keys=ON;          -- enforce referential integrity
PRAGMA busy_timeout=5000;        -- ms; cooperate with long reads
PRAGMA temp_store=MEMORY;        -- faster temp ops
```
(Upstream 3.50.x is current; WAL & these pragmas are standard patterns.)

**Schema (initial)**
```sql
-- strict typing, no dupes, time-series friendly
CREATE TABLE IF NOT EXISTS symbols(
  symbol TEXT PRIMARY KEY,
  name   TEXT,
  asset_type TEXT CHECK(asset_type IN ('ETF','STOCK','INDEX')) DEFAULT 'ETF',
  active INTEGER NOT NULL DEFAULT 1
) STRICT;

CREATE TABLE IF NOT EXISTS prices(
  symbol TEXT NOT NULL REFERENCES symbols(symbol) ON DELETE CASCADE,
  date   TEXT NOT NULL,                     -- ISO yyyy-mm-dd
  open   REAL NOT NULL, high REAL NOT NULL, low REAL NOT NULL,
  close  REAL NOT NULL, adj_close REAL, volume INTEGER,
  PRIMARY KEY(symbol,date)
) STRICT;

-- Derived metrics (persist to avoid recomputation)
CREATE TABLE IF NOT EXISTS indicators(
  symbol TEXT NOT NULL,
  date   TEXT NOT NULL,
  r_1m   REAL, r_3m REAL, r_6m REAL, r_12m REAL,
  vol_3m REAL, vol_6m REAL,
  score  REAL,
  rank   INTEGER,
  PRIMARY KEY(symbol,date),
  FOREIGN KEY(symbol,date) REFERENCES prices(symbol,date) ON DELETE CASCADE
) STRICT;

-- Run metadata and logs
CREATE TABLE IF NOT EXISTS runs(
  run_id     INTEGER PRIMARY KEY AUTOINCREMENT,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  status     TEXT CHECK(status IN ('OK','ERROR')) DEFAULT 'OK',
  notes      TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS fetch_log(
  run_id  INTEGER NOT NULL REFERENCES runs(run_id) ON DELETE CASCADE,
  symbol  TEXT NOT NULL,
  from_dt TEXT, to_dt TEXT,
  rows    INTEGER DEFAULT 0,
  ok      INTEGER NOT NULL DEFAULT 1,
  msg     TEXT,
  PRIMARY KEY(run_id,symbol)
) STRICT;
```

**Indexes**
```sql
CREATE INDEX IF NOT EXISTS idx_prices_symbol_date ON prices(symbol,date DESC);
CREATE INDEX IF NOT EXISTS idx_indicators_date_rank ON indicators(date DESC, rank);
```

**Backups**
- Use SQLite online backup (VACUUM INTO) post-run; optional Litestream if remote object storage is desired.

---

## 7) Analytics (implementation notes)

- Use trading-day calendars (NYSE holiday file) but tolerate gaps (ETFs often still daily).
- All momentum math is *total return* over lookbacks using `adj_close` when available.
- Volatility penalty: rolling σ of daily log returns over 63/126 days (configurable).
- Composite score: z-score or min-max of lookback returns minus λ·vol; deterministic tie-breaks (lower vol, higher liquidity).
- Breadth screen: require positive N-of-M lookbacks and minimum dollar ADV.

*The exact formulae and steps follow the methodology we documented earlier; this doc focuses on implementation.*

---

## 8) Rate Limiting & Fetch Strategy (free tier)

- Global budget: **25 API calls/day** on Alpha Vantage free plan.
- **Strategy**:
  - Cache last fetched date per symbol; only call when the latest stored `date` < last trading day.
  - Stagger large universes across days (e.g., 25/day).
  - Prefer “full series once” on first run, then “compact/delta” (if provider supports).
  - Backfill via Stooq CSV imports when bootstrapping manually.

---

## 9) TUI Requirements

- **Screens**:
  - *Dashboard*: last run status, cache stats, next quota reset, quick actions.
  - *Leaders*: Top-5 / Top-N with columns [Rank, Symbol, Score, R1M/R3M/R6M, Vol, ADV].
  - *Universe*: filter/search; toggle active symbols.
  - *Symbol detail*: sparkline, return table, volatility, raw vs. adjusted close.
  - *Runs/Logs*: past runs, errors, export logs.

- **Controls**: `r` refresh, `/` search, `e` export CSV, `←/→` tab screens, `q` quit.
- **UX**: Non-blocking refresh (spinner), errors routed to Runs/Logs.

---

## 10) Configuration

- `config.yaml` (or env vars)  
  - `alpha_vantage.api_key` (env: `ALPHAVANTAGE_API_KEY`)  
  - `universe: [ "QQQ", "IWM", ... ]`  
  - `lookbacks: { r1m:21, r3m:63, r6m:126, r12m:252 }`  
  - `vol_windows: { short:63, long:126 }`  
  - `penalty_lambda: 0.35`  
  - `min_adv_usd: 5_000_000`  
  - `data_dir: "./data"`  
  - `export_dir: "./exports"`

---

## 11) Containerization

**Dockerfile (multi-stage)**
- *Builder*: `golang:1.25.1` → build static binary (`CGO_ENABLED=0`, pure Go).
- *Runner*: `gcr.io/distroless/static:nonroot` with `WORKDIR /app` and `USER 65532`.
- Mount `/data` volume for the DB and exports.

**Compose**
- Compose v2 syntax (no `version:` field; it’s obsolete).
- Healthcheck: basic `CMD` that pings the app’s `--ping` subcommand (exits 0).

---

## 12) Security & Secrets

- API keys from env or Docker secrets (never commit).
- Read-only runtime FS; write only to `/data`.
- Run as non-root in container; drop extra Linux capabilities (Compose `cap_drop: ALL`).

---

## 13) Telemetry & Logging

- Structured logs (JSON) with run_id, symbol, action, outcome, latency.
- Export `runs.csv` and `leaders-YYYYMMDD.csv` on each successful refresh.

---

## 14) Testing

- **Unit**: math for returns/volatility/scoring (golden vectors).
- **DB**: migration up/down; idempotent re-runs; WAL behavior.
- **Integration**: record/replay HTTP fixtures for Alpha Vantage; CSV ingest for Stooq.
- **Determinism**: given a fixed DB snapshot and config, screen must be byte-identical.

---

## 15) Non-Functional Requirements

- **Performance**: With a 25-symbol universe, full delta refresh < 10s on a laptop (excluding network).
- **Portability**: Linux/macOS; containers for Windows via Docker Desktop (keep Desktop patched for security).
- **Footprint**: Final image ≤ 20–30 MB typical for static Go + distroless.
- **Reliability**: Survive provider timeouts; exponential backoff; partial progress persisted per symbol.

---

## 16) Deliverables

- `cmd/momo/` main (TUI)
- `internal/fetch/` Alpha Vantage + CSV adapters
- `internal/db/` migrations & queries
- `internal/analytics/` indicators & scoring
- `internal/ui/` Bubble Tea models/views
- `configs/config.example.yaml`
- `Dockerfile`, `docker-compose.yml`
- `docs/quickstart.md`
- `test/` (unit + integration fixtures)

---

## 17) Acceptance Criteria

- Runs in Docker, creates `/data/momentum.db`, and persists at least 1 symbol.
- Given a 10-symbol universe and recorded fixtures, *Leaders* shows deterministic Top-5.
- Export commands create CSVs in `/data/exports`.
- Hitting the Alpha Vantage daily cap triggers a friendly message and partial results are still usable.

---

## 18) Change Log (v0.2)

- **Storage**: Standardized on **SQLite** (removed fs/JSON option); added WAL, STRICT tables, schema above.
- **Stack versions**: Updated to Go **1.25.1**, Bubble Tea **v1.3.10**, Bubbles **v0.21.0**, Lip Gloss **v1.0.0**, modernc.org/sqlite **v1.39.0**, Docker Engine **28.x**, Compose **v2.40.0**.

---

## References

- Go 1.25.1 downloads (stable).
- Bubble Tea releases (v1.3.x stable; v2 pre-release).
- Bubbles v0.21.0 release.
- Lip Gloss v1.0.0.
- modernc.org/sqlite v1.39.0 (pkg.go.dev).
- SQLite 3.50.x release.
- Docker Engine v28 series notes.
- Docker Compose v2.40.0 release & install docs; Compose v2 syntax (no `version:`).
- Distroless images & `:nonroot` practice.
- Alpha Vantage free 25 req/day support page.
- Stooq CSV download.
- Docker Desktop security patch context (keep Desktop current).