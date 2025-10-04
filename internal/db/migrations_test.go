package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	err = db.Migrate()
	require.NoError(t, err)

	// Verify schema_migrations table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count) // One migration applied

	// Verify current version is 1
	version, err := db.getCurrentVersion()
	require.NoError(t, err)
	assert.Equal(t, 1, version)

	// Verify all tables were created
	tables := []string{"symbols", "prices", "indicators", "runs", "fetch_log"}
	for _, table := range tables {
		err = db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		assert.NoError(t, err, "Table %s should exist", table)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Run migrations twice
	err = db.Migrate()
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	// Should still have only one migration record
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetAppliedMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	err = db.Migrate()
	require.NoError(t, err)

	migrations, err := db.GetAppliedMigrations()
	require.NoError(t, err)
	require.Len(t, migrations, 1)

	assert.Equal(t, 1, migrations[0].Version)
	assert.Contains(t, migrations[0].Description, "Initial schema")
	assert.NotEmpty(t, migrations[0].AppliedAt)
}

func TestRollback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Apply migration
	err = db.Migrate()
	require.NoError(t, err)

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&count)
	require.NoError(t, err)

	// Rollback
	err = db.Rollback()
	require.NoError(t, err)

	// Verify tables are gone
	err = db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&count)
	assert.Error(t, err) // Table should not exist

	// Verify version is 0
	version, err := db.getCurrentVersion()
	require.NoError(t, err)
	assert.Equal(t, 0, version)
}

func TestRollbackNoMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Initialize migrations table but don't apply migrations
	err = db.initMigrationsTable()
	require.NoError(t, err)

	// Rollback should fail when there are no migrations to rollback
	err = db.Rollback()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no migrations to rollback")
}

func TestSchemaIndexes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	err = db.Migrate()
	require.NoError(t, err)

	// Verify indexes were created
	indexes := []string{
		"idx_prices_symbol_date",
		"idx_indicators_date_rank",
		"idx_symbols_active",
		"idx_runs_status_date",
		"idx_fetch_log_failures",
	}

	for _, idx := range indexes {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?"
		err = db.QueryRow(query, idx).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Index %s should exist", idx)
	}
}
