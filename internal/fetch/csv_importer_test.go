package fetch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSVImporter(t *testing.T) {
	importer := NewCSVImporter()

	assert.NotNil(t, importer)
	assert.Equal(t, "2006-01-02", importer.DateFormat)
	assert.True(t, importer.SkipHeader)
}

func TestCSVImporter_Import_ValidData(t *testing.T) {
	csvData := `Date,Open,High,Low,Close,Volume
2024-01-01,100.50,102.00,99.50,101.00,1000000
2024-01-02,101.00,103.50,100.00,103.00,1200000
2024-01-03,103.00,105.00,102.50,104.50,1500000`

	importer := NewCSVImporter()
	records, err := importer.Import(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Len(t, records, 3)

	// Verify first record
	assert.Equal(t, "2024-01-01", records[0].Date.Format("2006-01-02"))
	assert.Equal(t, 100.50, records[0].Open)
	assert.Equal(t, 102.00, records[0].High)
	assert.Equal(t, 99.50, records[0].Low)
	assert.Equal(t, 101.00, records[0].Close)
	assert.Equal(t, 1000000.0, records[0].Volume)

	// Verify second record
	assert.Equal(t, "2024-01-02", records[1].Date.Format("2006-01-02"))
	assert.Equal(t, 101.00, records[1].Open)
	assert.Equal(t, 103.50, records[1].High)

	// Verify third record
	assert.Equal(t, "2024-01-03", records[2].Date.Format("2006-01-02"))
	assert.Equal(t, 104.50, records[2].Close)
}

func TestCSVImporter_Import_NoHeader(t *testing.T) {
	csvData := `2024-01-01,100.50,102.00,99.50,101.00,1000000
2024-01-02,101.00,103.50,100.00,103.00,1200000`

	importer := NewCSVImporter()
	importer.SkipHeader = false

	records, err := importer.Import(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestCSVImporter_Import_InvalidDate(t *testing.T) {
	csvData := `Date,Open,High,Low,Close,Volume
invalid-date,100.50,102.00,99.50,101.00,1000000`

	importer := NewCSVImporter()
	_, err := importer.Import(strings.NewReader(csvData))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse date")
}

func TestCSVImporter_Import_InvalidNumber(t *testing.T) {
	csvData := `Date,Open,High,Low,Close,Volume
2024-01-01,invalid,102.00,99.50,101.00,1000000`

	importer := NewCSVImporter()
	_, err := importer.Import(strings.NewReader(csvData))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse open price")
}

func TestCSVImporter_Import_InsufficientColumns(t *testing.T) {
	csvData := `Date,Open,High,Low,Close
2024-01-01,100.50,102.00,99.50,101.00`

	importer := NewCSVImporter()
	_, err := importer.Import(strings.NewReader(csvData))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected at least 6 columns")
}

func TestCSVImporter_ImportFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_data.csv")

	csvData := `Date,Open,High,Low,Close,Volume
2024-01-01,100.50,102.00,99.50,101.00,1000000
2024-01-02,101.00,103.50,100.00,103.00,1200000`

	err := os.WriteFile(csvPath, []byte(csvData), 0644)
	require.NoError(t, err)

	importer := NewCSVImporter()
	records, err := importer.ImportFromFile(csvPath)

	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestCSVImporter_ImportFromFile_FileNotFound(t *testing.T) {
	importer := NewCSVImporter()
	_, err := importer.ImportFromFile("/nonexistent/path/data.csv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open CSV file")
}

func TestValidateRecords_Valid(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100.0,
			High:   105.0,
			Low:    99.0,
			Close:  103.0,
			Volume: 1000000,
		},
		{
			Date:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Open:   103.0,
			High:   108.0,
			Low:    102.0,
			Close:  107.0,
			Volume: 1200000,
		},
	}

	err := ValidateRecords(records)
	assert.NoError(t, err)
}

func TestValidateRecords_Empty(t *testing.T) {
	records := []CSVRecord{}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no records found")
}

func TestValidateRecords_NegativePrice(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   -100.0,
			High:   105.0,
			Low:    99.0,
			Close:  103.0,
			Volume: 1000000,
		},
	}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative price values")
}

func TestValidateRecords_NegativeVolume(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100.0,
			High:   105.0,
			Low:    99.0,
			Close:  103.0,
			Volume: -1000000,
		},
	}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative volume")
}

func TestValidateRecords_HighLowInconsistent(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100.0,
			High:   95.0, // High < Low
			Low:    99.0,
			Close:  98.0,
			Volume: 1000000,
		},
	}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "high < low")
}

func TestValidateRecords_OpenOutOfRange(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   110.0, // Open > High
			High:   105.0,
			Low:    99.0,
			Close:  103.0,
			Volume: 1000000,
		},
	}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open outside high/low range")
}

func TestValidateRecords_CloseOutOfRange(t *testing.T) {
	records := []CSVRecord{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100.0,
			High:   105.0,
			Low:    99.0,
			Close:  108.0, // Close > High
			Volume: 1000000,
		},
	}

	err := ValidateRecords(records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close outside high/low range")
}
