-- Momentum Screener TUI Database Schema
-- SQLite with STRICT tables for type safety
-- WAL mode enabled via pragmas in connection.go

-- Symbols table: tracks all tickers in the universe
CREATE TABLE IF NOT EXISTS symbols(
  symbol TEXT PRIMARY KEY,
  name   TEXT,
  asset_type TEXT CHECK(asset_type IN ('ETF','STOCK','INDEX')) DEFAULT 'ETF',
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
) STRICT;

-- Prices table: historical OHLCV data
CREATE TABLE IF NOT EXISTS prices(
  symbol TEXT NOT NULL REFERENCES symbols(symbol) ON DELETE CASCADE,
  date   TEXT NOT NULL,                     -- ISO yyyy-mm-dd
  open   REAL NOT NULL,
  high   REAL NOT NULL,
  low    REAL NOT NULL,
  close  REAL NOT NULL,
  adj_close REAL,                           -- Adjusted close for splits/dividends
  volume INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY(symbol, date)
) STRICT;

-- Indicators table: derived momentum metrics
CREATE TABLE IF NOT EXISTS indicators(
  symbol TEXT NOT NULL,
  date   TEXT NOT NULL,
  r_1m   REAL,                              -- 1-month return (21 days)
  r_3m   REAL,                              -- 3-month return (63 days)
  r_6m   REAL,                              -- 6-month return (126 days)
  r_12m  REAL,                              -- 12-month return (252 days)
  vol_3m REAL,                              -- 3-month volatility
  vol_6m REAL,                              -- 6-month volatility
  adv    REAL,                              -- Average dollar volume
  score  REAL,                              -- Composite momentum score
  rank   INTEGER,                           -- Rank within universe
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY(symbol, date),
  FOREIGN KEY(symbol, date) REFERENCES prices(symbol, date) ON DELETE CASCADE
) STRICT;

-- Runs table: tracks each refresh/computation run
CREATE TABLE IF NOT EXISTS runs(
  run_id     INTEGER PRIMARY KEY AUTOINCREMENT,
  started_at TEXT NOT NULL DEFAULT (datetime('now')),
  finished_at TEXT,
  status     TEXT CHECK(status IN ('RUNNING','OK','ERROR')) DEFAULT 'RUNNING',
  symbols_processed INTEGER DEFAULT 0,
  symbols_failed INTEGER DEFAULT 0,
  notes      TEXT
) STRICT;

-- Fetch log: detailed record of each symbol fetch attempt
CREATE TABLE IF NOT EXISTS fetch_log(
  run_id  INTEGER NOT NULL REFERENCES runs(run_id) ON DELETE CASCADE,
  symbol  TEXT NOT NULL,
  from_dt TEXT,                             -- Start date of fetched range
  to_dt   TEXT,                             -- End date of fetched range
  rows    INTEGER DEFAULT 0,                -- Number of rows fetched
  ok      INTEGER NOT NULL DEFAULT 1,       -- Success flag (1=success, 0=failure)
  msg     TEXT,                             -- Error message or notes
  fetched_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY(run_id, symbol)
) STRICT;

-- Indexes for query performance

-- Fast lookup of prices by symbol and date (descending for latest first)
CREATE INDEX IF NOT EXISTS idx_prices_symbol_date
  ON prices(symbol, date DESC);

-- Fast retrieval of latest indicators sorted by rank
CREATE INDEX IF NOT EXISTS idx_indicators_date_rank
  ON indicators(date DESC, rank);

-- Fast lookup of active symbols
CREATE INDEX IF NOT EXISTS idx_symbols_active
  ON symbols(active) WHERE active = 1;

-- Fast lookup of runs by status and date
CREATE INDEX IF NOT EXISTS idx_runs_status_date
  ON runs(status, started_at DESC);

-- Fast lookup of failed fetches
CREATE INDEX IF NOT EXISTS idx_fetch_log_failures
  ON fetch_log(ok, run_id) WHERE ok = 0;
