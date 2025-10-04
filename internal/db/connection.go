package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// DB wraps the database connection with additional metadata
type DB struct {
	*sql.DB
	Path string
}

// Config holds database connection configuration
type Config struct {
	Path string // Full path to SQLite database file
}

// New creates a new database connection with proper pragmas configured
func New(cfg Config) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	sqlDB.SetMaxIdleConns(1)

	db := &DB{
		DB:   sqlDB,
		Path: cfg.Path,
	}

	// Apply SQLite pragmas for optimal performance and reliability
	if err := db.applyPragmas(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to apply pragmas: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// applyPragmas configures SQLite with optimal settings for the application
func (db *DB) applyPragmas() error {
	pragmas := []string{
		// Enable WAL mode for better concurrency (readers don't block writers)
		"PRAGMA journal_mode=WAL",

		// NORMAL synchronous mode is safe with WAL and much faster
		"PRAGMA synchronous=NORMAL",

		// Enforce foreign key constraints
		"PRAGMA foreign_keys=ON",

		// Wait up to 5 seconds when the database is locked
		"PRAGMA busy_timeout=5000",

		// Store temp tables in memory for performance
		"PRAGMA temp_store=MEMORY",

		// Cache size: -2000 means 2MB of cache (negative = KB)
		"PRAGMA cache_size=-2000",

		// Memory-mapped I/O for better performance (16MB)
		"PRAGMA mmap_size=16777216",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute %q: %w", pragma, err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// GetInfo returns database metadata for diagnostics
func (db *DB) GetInfo() (map[string]string, error) {
	info := make(map[string]string)

	queries := map[string]string{
		"journal_mode":  "PRAGMA journal_mode",
		"synchronous":   "PRAGMA synchronous",
		"foreign_keys":  "PRAGMA foreign_keys",
		"page_size":     "PRAGMA page_size",
		"page_count":    "PRAGMA page_count",
		"schema_version": "PRAGMA schema_version",
	}

	for key, query := range queries {
		var value string
		if err := db.QueryRow(query).Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", key, err)
		}
		info[key] = value
	}

	return info, nil
}
