package fetch

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// CSVRecord represents a parsed CSV record from Stooq or similar sources.
type CSVRecord struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// CSVImporter handles importing OHLCV data from CSV files.
type CSVImporter struct {
	// DateFormat is the expected date format in the CSV (default: "2006-01-02")
	DateFormat string
	// SkipHeader indicates whether to skip the first row (default: true)
	SkipHeader bool
}

// NewCSVImporter creates a new CSV importer with default settings.
func NewCSVImporter() *CSVImporter {
	return &CSVImporter{
		DateFormat: "2006-01-02",
		SkipHeader: true,
	}
}

// ImportFromFile imports OHLCV data from a CSV file.
// Expected format: Date,Open,High,Low,Close,Volume
// This matches Stooq's export format.
func (ci *CSVImporter) ImportFromFile(filePath string) ([]CSVRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	return ci.Import(file)
}

// Import reads OHLCV data from a CSV reader.
func (ci *CSVImporter) Import(reader io.Reader) ([]CSVRecord, error) {
	csvReader := csv.NewReader(reader)

	// Skip header if configured
	if ci.SkipHeader {
		if _, err := csvReader.Read(); err != nil {
			return nil, fmt.Errorf("failed to read CSV header: %w", err)
		}
	}

	var records []CSVRecord

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		// Validate row length
		if len(row) < 6 {
			return nil, fmt.Errorf("invalid CSV row (expected at least 6 columns, got %d): %v", len(row), row)
		}

		// Parse date
		date, err := time.Parse(ci.DateFormat, row[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", row[0], err)
		}

		// Parse OHLCV values
		open, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse open price %q: %w", row[1], err)
		}

		high, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse high price %q: %w", row[2], err)
		}

		low, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse low price %q: %w", row[3], err)
		}

		close, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse close price %q: %w", row[4], err)
		}

		volume, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse volume %q: %w", row[5], err)
		}

		records = append(records, CSVRecord{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return records, nil
}

// ValidateRecords performs basic validation on imported CSV records.
func ValidateRecords(records []CSVRecord) error {
	if len(records) == 0 {
		return fmt.Errorf("no records found in CSV")
	}

	for i, rec := range records {
		// Check for negative prices
		if rec.Open < 0 || rec.High < 0 || rec.Low < 0 || rec.Close < 0 {
			return fmt.Errorf("record %d has negative price values", i)
		}

		// Check for negative volume
		if rec.Volume < 0 {
			return fmt.Errorf("record %d has negative volume", i)
		}

		// Check high/low consistency
		if rec.High < rec.Low {
			return fmt.Errorf("record %d has high < low", i)
		}

		// Check that open/close are within high/low range
		if rec.Open > rec.High || rec.Open < rec.Low {
			return fmt.Errorf("record %d has open outside high/low range", i)
		}

		if rec.Close > rec.High || rec.Close < rec.Low {
			return fmt.Errorf("record %d has close outside high/low range", i)
		}
	}

	return nil
}
