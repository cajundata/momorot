package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/cajundata/momorot/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.New(db.Config{Path: dbPath})
	require.NoError(t, err)

	// Run migrations
	err = database.Migrate()
	require.NoError(t, err)

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

func setupTestData(t *testing.T, database *db.DB) {
	t.Helper()

	// Add test symbols
	symbolRepo := db.NewSymbolRepository(database)
	symbols := []*db.Symbol{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
		{Symbol: "QQQ", Name: "Invesco QQQ", AssetType: "ETF", Active: true},
		{Symbol: "IWM", Name: "iShares Russell 2000", AssetType: "ETF", Active: true},
	}
	for _, sym := range symbols {
		require.NoError(t, symbolRepo.Create(sym))
	}

	// Add price data
	priceRepo := db.NewPriceRepository(database)
	for _, symbol := range []string{"SPY", "QQQ", "IWM"} {
		adjClose := 450.0
		volume := int64(1000000)
		err := priceRepo.Create(&db.Price{
			Symbol:   symbol,
			Date:     "2025-10-08",
			Open:     445.0,
			High:     452.0,
			Low:      444.0,
			Close:    450.0,
			AdjClose: &adjClose,
			Volume:   &volume,
		})
		require.NoError(t, err)
	}

	// Add indicator data
	indicatorRepo := db.NewIndicatorRepository(database)
	r1m1, r3m1, r6m1, r12m1 := 0.15, 0.25, 0.35, 0.45
	vol3m1, vol6m1 := 0.12, 0.15
	adv1 := 5000000.0
	score1 := 1.5
	rank1 := 1

	r1m2, r3m2, r6m2, r12m2 := 0.12, 0.22, 0.32, 0.42
	vol3m2, vol6m2 := 0.10, 0.13
	adv2 := 4000000.0
	score2 := 1.2
	rank2 := 2

	r1m3, r3m3, r6m3, r12m3 := 0.10, 0.20, 0.30, 0.40
	vol3m3, vol6m3 := 0.14, 0.17
	adv3 := 3000000.0
	score3 := 1.0
	rank3 := 3

	indicators := []db.Indicator{
		{
			Symbol: "SPY",
			Date:   "2025-10-08",
			R1M:    &r1m1,
			R3M:    &r3m1,
			R6M:    &r6m1,
			R12M:   &r12m1,
			Vol3M:  &vol3m1,
			Vol6M:  &vol6m1,
			ADV:    &adv1,
			Score:  &score1,
			Rank:   &rank1,
		},
		{
			Symbol: "QQQ",
			Date:   "2025-10-08",
			R1M:    &r1m2,
			R3M:    &r3m2,
			R6M:    &r6m2,
			R12M:   &r12m2,
			Vol3M:  &vol3m2,
			Vol6M:  &vol6m2,
			ADV:    &adv2,
			Score:  &score2,
			Rank:   &rank2,
		},
		{
			Symbol: "IWM",
			Date:   "2025-10-08",
			R1M:    &r1m3,
			R3M:    &r3m3,
			R6M:    &r6m3,
			R12M:   &r12m3,
			Vol3M:  &vol3m3,
			Vol6M:  &vol6m3,
			ADV:    &adv3,
			Score:  &score3,
			Rank:   &rank3,
		},
	}
	require.NoError(t, indicatorRepo.UpsertBatch(indicators))

	// Add run data
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("test run")
	require.NoError(t, err)
	require.NoError(t, runRepo.Finish(runID, "OK", 3, 0))
}

func TestNew(t *testing.T) {
	database := setupTestDB(t)

	exporter := New(database, "./test-exports")

	assert.NotNil(t, exporter)
	assert.Equal(t, "./test-exports", exporter.exportDir)
	assert.NotNil(t, exporter.database)
}

func TestExportLeaders(t *testing.T) {
	database := setupTestDB(t)


	setupTestData(t, database)

	// Create temp directory for exports
	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export top 2 leaders
	filename, err := exporter.ExportLeaders(2, "2025-10-08")
	require.NoError(t, err)
	assert.Contains(t, filename, "leaders-")
	assert.Contains(t, filename, ".csv")

	// Verify file exists
	assert.FileExists(t, filename)

	// Read and verify CSV content
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Check header
	assert.Equal(t, 12, len(records[0]), "Header should have 12 columns")
	assert.Equal(t, "Rank", records[0][0])
	assert.Equal(t, "Symbol", records[0][1])
	assert.Equal(t, "Score", records[0][4])

	// Check data rows (should have 2 rows + 1 header = 3 total)
	assert.Equal(t, 3, len(records), "Should have header + 2 data rows")

	// Verify first leader is SPY (rank 1)
	assert.Equal(t, "1", records[1][0])
	assert.Equal(t, "SPY", records[1][1])
	assert.Equal(t, "SPDR S&P 500", records[1][2])

	// Verify second leader is QQQ (rank 2)
	assert.Equal(t, "2", records[2][0])
	assert.Equal(t, "QQQ", records[2][1])
}

func TestExportLeadersEmptyDate(t *testing.T) {
	database := setupTestDB(t)


	setupTestData(t, database)

	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export without specifying date (should use today's date in query)
	// Since test data uses 2025-10-08 and today is 2025-10-08, it will find data
	filename, err := exporter.ExportLeaders(5, "")
	require.NoError(t, err)
	assert.FileExists(t, filename)

	// File should exist
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + data (if today matches test data date)
	assert.GreaterOrEqual(t, len(records), 1, "Should have at least header row")
}

func TestExportFullRankings(t *testing.T) {
	database := setupTestDB(t)


	setupTestData(t, database)

	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export all rankings
	filename, err := exporter.ExportFullRankings("2025-10-08")
	require.NoError(t, err)
	assert.Contains(t, filename, "rankings-")
	assert.FileExists(t, filename)

	// Read and verify CSV content
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Check header + 3 data rows = 4 total
	assert.Equal(t, 4, len(records))

	// Verify all 3 symbols are present in rank order
	assert.Equal(t, "1", records[1][0])
	assert.Equal(t, "SPY", records[1][1])
	assert.Equal(t, "2", records[2][0])
	assert.Equal(t, "QQQ", records[2][1])
	assert.Equal(t, "3", records[3][0])
	assert.Equal(t, "IWM", records[3][1])
}

func TestExportRuns(t *testing.T) {
	database := setupTestDB(t)


	setupTestData(t, database)

	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export runs
	filename, err := exporter.ExportRuns()
	require.NoError(t, err)
	assert.Contains(t, filename, "runs-")
	assert.FileExists(t, filename)

	// Read and verify CSV content
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Check header
	assert.Equal(t, 7, len(records[0]), "Header should have 7 columns")
	assert.Equal(t, "RunID", records[0][0])
	assert.Equal(t, "Status", records[0][3])

	// Check data rows (at least 1 run)
	assert.GreaterOrEqual(t, len(records), 2, "Should have header + at least 1 data row")

	// Verify run data
	assert.Equal(t, "1", records[1][0]) // RunID
	assert.Equal(t, "OK", records[1][3]) // Status
	assert.Equal(t, "3", records[1][4])  // SymbolsProcessed
	assert.Equal(t, "0", records[1][5])  // SymbolsFailed
}

func TestExportSymbolDetail(t *testing.T) {
	database := setupTestDB(t)


	setupTestData(t, database)

	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export symbol detail for SPY
	filename, err := exporter.ExportSymbolDetail("SPY")
	require.NoError(t, err)
	assert.Contains(t, filename, "symbol-SPY-")
	assert.FileExists(t, filename)

	// Read and verify CSV content
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Check header
	assert.Equal(t, 16, len(records[0]), "Header should have 16 columns")
	assert.Equal(t, "Date", records[0][0])
	assert.Equal(t, "Close", records[0][4])
	assert.Equal(t, "R1M", records[0][7])

	// Check data rows (at least 1 price entry)
	assert.GreaterOrEqual(t, len(records), 2, "Should have header + at least 1 data row")

	// Verify price data
	assert.Equal(t, "2025-10-08", records[1][0]) // Date
	assert.Equal(t, "450.00", records[1][4])      // Close
	assert.Equal(t, "15.00%", records[1][7])      // R1M (0.15 * 100)
}

func TestEnsureExportDir(t *testing.T) {
	database := setupTestDB(t)


	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "nested", "export", "dir")
	exporter := New(database, exportPath)

	// Directory shouldn't exist initially
	_, err := os.Stat(exportPath)
	assert.True(t, os.IsNotExist(err))

	// Export should create the directory
	_, err = exporter.ExportRuns()
	require.NoError(t, err)

	// Directory should now exist
	_, err = os.Stat(exportPath)
	assert.NoError(t, err)
}

func TestFormatHelpers(t *testing.T) {
	// Test formatFloat
	val1 := 123.456789
	assert.Equal(t, "123.46", formatFloat(&val1, 2))
	assert.Equal(t, "123.457", formatFloat(&val1, 3))
	assert.Equal(t, "", formatFloat(nil, 2))

	// Test formatPercent
	val2 := 0.1234
	assert.Equal(t, "12.34%", formatPercent(&val2))
	assert.Equal(t, "", formatPercent(nil))

	// Test formatInt
	val3 := 42
	assert.Equal(t, "42", formatInt(&val3))
	assert.Equal(t, "", formatInt(nil))

	// Test formatInt64
	val4 := int64(1000000)
	assert.Equal(t, "1000000", formatInt64(&val4))
	assert.Equal(t, "", formatInt64(nil))

	// Test formatString
	val5 := "test"
	assert.Equal(t, "test", formatString(&val5))
	assert.Equal(t, "", formatString(nil))
}

func TestExportLeadersNoData(t *testing.T) {
	database := setupTestDB(t)


	// No data setup

	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export should succeed even with no data
	filename, err := exporter.ExportLeaders(5, "2025-10-08")
	require.NoError(t, err)
	assert.FileExists(t, filename)

	// File should have only header
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Equal(t, 1, len(records), "Should have only header row")
}

func TestExportSymbolDetailNoData(t *testing.T) {
	database := setupTestDB(t)


	tempDir := t.TempDir()
	exporter := New(database, tempDir)

	// Export should succeed even with no data
	filename, err := exporter.ExportSymbolDetail("NOTFOUND")
	require.NoError(t, err)
	assert.FileExists(t, filename)

	// File should have only header
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Equal(t, 1, len(records), "Should have only header row")
}
