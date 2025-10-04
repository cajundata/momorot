package db

import (
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// Migration represents a database migration with version tracking
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// Migrate runs all pending migrations to bring the database to the latest schema version
func (db *DB) Migrate() error {
	// Create migrations tracking table if it doesn't exist
	if err := db.initMigrationsTable(); err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}

	// Get current schema version
	currentVersion, err := db.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Define all migrations
	migrations := []Migration{
		{
			Version:     1,
			Description: "Initial schema with symbols, prices, indicators, runs, and fetch_log",
			Up:          schemaSQL,
			Down:        dropAllTables,
		},
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue // Already applied
		}

		if err := db.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// initMigrationsTable creates the migrations tracking table
func (db *DB) initMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations(
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		) STRICT
	`
	_, err := db.Exec(query)
	return err
}

// getCurrentVersion returns the current schema version
func (db *DB) getCurrentVersion() (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// applyMigration applies a single migration within a transaction
func (db *DB) applyMigration(m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(m.Up); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	recordQuery := `INSERT INTO schema_migrations (version, description) VALUES (?, ?)`
	if _, err := tx.Exec(recordQuery, m.Version, m.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the last applied migration
func (db *DB) Rollback() error {
	currentVersion, err := db.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the migration to rollback
	migrations := []Migration{
		{
			Version:     1,
			Description: "Initial schema",
			Up:          schemaSQL,
			Down:        dropAllTables,
		},
	}

	var targetMigration *Migration
	for _, m := range migrations {
		if m.Version == currentVersion {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration version %d not found", currentVersion)
	}

	// Apply rollback
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute down migration
	if _, err := tx.Exec(targetMigration.Down); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", currentVersion); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of all applied migrations
func (db *DB) GetAppliedMigrations() ([]MigrationRecord, error) {
	query := `SELECT version, description, applied_at FROM schema_migrations ORDER BY version`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var m MigrationRecord
		if err := rows.Scan(&m.Version, &m.Description, &m.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

// MigrationRecord represents a migration that has been applied
type MigrationRecord struct {
	Version     int
	Description string
	AppliedAt   string
}

// dropAllTables is the down migration for version 1
const dropAllTables = `
DROP INDEX IF EXISTS idx_fetch_log_failures;
DROP INDEX IF EXISTS idx_runs_status_date;
DROP INDEX IF EXISTS idx_symbols_active;
DROP INDEX IF EXISTS idx_indicators_date_rank;
DROP INDEX IF EXISTS idx_prices_symbol_date;
DROP TABLE IF EXISTS fetch_log;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS indicators;
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS symbols;
`
