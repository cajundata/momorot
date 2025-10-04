package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	// Verify connection works
	err = db.Ping()
	assert.NoError(t, err)

	// Verify path is stored correctly
	assert.Equal(t, dbPath, db.Path)
}

func TestApplyPragmas(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Verify WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)

	// Verify foreign keys are enabled
	var foreignKeys int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	require.NoError(t, err)
	assert.Equal(t, 1, foreignKeys)

	// Verify busy timeout is set
	var busyTimeout int
	err = db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	require.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout)
}

func TestGetInfo(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	info, err := db.GetInfo()
	require.NoError(t, err)

	// Verify expected keys are present
	assert.Contains(t, info, "journal_mode")
	assert.Contains(t, info, "synchronous")
	assert.Contains(t, info, "foreign_keys")
	assert.Contains(t, info, "page_size")

	// Verify WAL mode
	assert.Equal(t, "wal", info["journal_mode"])

	// Verify foreign keys enabled
	assert.Equal(t, "1", info["foreign_keys"])
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)

	// Close should succeed
	err = db.Close()
	assert.NoError(t, err)

	// Closing again should not panic
	err = db.Close()
	assert.NoError(t, err)
}

func TestDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	// Use nested directory that doesn't exist yet
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// Verify directories were created
	_, err = os.Stat(filepath.Dir(dbPath))
	assert.NoError(t, err)
}
