package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cajundata/momorot/internal/db"
)

// Exporter handles data export operations.
type Exporter struct {
	database  *db.DB
	exportDir string
}

// New creates a new Exporter instance.
func New(database *db.DB, exportDir string) *Exporter {
	return &Exporter{
		database:  database,
		exportDir: exportDir,
	}
}

// ensureExportDir creates the export directory if it doesn't exist.
func (e *Exporter) ensureExportDir() error {
	return os.MkdirAll(e.exportDir, 0755)
}

// ExportLeaders exports the top N leaders to a CSV file.
// Filename format: leaders-YYYYMMDD.csv
func (e *Exporter) ExportLeaders(topN int, date string) (string, error) {
	if err := e.ensureExportDir(); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// If no date provided, use today
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Query top N leaders for the given date
	query := `
		SELECT
			i.rank,
			i.symbol,
			s.name,
			s.asset_type,
			i.score,
			i.r_1m,
			i.r_3m,
			i.r_6m,
			i.r_12m,
			i.vol_3m,
			i.vol_6m,
			i.adv
		FROM indicators i
		JOIN symbols s ON i.symbol = s.symbol
		WHERE i.date = ? AND i.rank IS NOT NULL
		ORDER BY i.rank ASC
		LIMIT ?
	`

	rows, err := e.database.Query(query, date, topN)
	if err != nil {
		return "", fmt.Errorf("failed to query leaders: %w", err)
	}
	defer rows.Close()

	// Create output file
	dateStr := time.Now().Format("20060102")
	filename := filepath.Join(e.exportDir, fmt.Sprintf("leaders-%s.csv", dateStr))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Rank", "Symbol", "Name", "Asset Type", "Score",
		"R1M", "R3M", "R6M", "R12M", "Vol3M", "Vol6M", "ADV",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	rowCount := 0
	for rows.Next() {
		var rank *int
		var symbol, name, assetType string
		var score, r1m, r3m, r6m, r12m, vol3m, vol6m, adv *float64

		err := rows.Scan(
			&rank, &symbol, &name, &assetType, &score,
			&r1m, &r3m, &r6m, &r12m, &vol3m, &vol6m, &adv,
		)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		row := []string{
			formatInt(rank),
			symbol,
			name,
			assetType,
			formatFloat(score, 3),
			formatPercent(r1m),
			formatPercent(r3m),
			formatPercent(r6m),
			formatPercent(r12m),
			formatPercent(vol3m),
			formatPercent(vol6m),
			formatFloat(adv, 0),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	return filename, nil
}

// ExportFullRankings exports all ranked symbols to a CSV file.
// Filename format: rankings-YYYYMMDD.csv
func (e *Exporter) ExportFullRankings(date string) (string, error) {
	if err := e.ensureExportDir(); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// If no date provided, use today
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Query all rankings for the given date
	query := `
		SELECT
			i.rank,
			i.symbol,
			s.name,
			s.asset_type,
			i.score,
			i.r_1m,
			i.r_3m,
			i.r_6m,
			i.r_12m,
			i.vol_3m,
			i.vol_6m,
			i.adv
		FROM indicators i
		JOIN symbols s ON i.symbol = s.symbol
		WHERE i.date = ? AND i.rank IS NOT NULL
		ORDER BY i.rank ASC
	`

	rows, err := e.database.Query(query, date)
	if err != nil {
		return "", fmt.Errorf("failed to query rankings: %w", err)
	}
	defer rows.Close()

	// Create output file
	dateStr := time.Now().Format("20060102")
	filename := filepath.Join(e.exportDir, fmt.Sprintf("rankings-%s.csv", dateStr))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Rank", "Symbol", "Name", "Asset Type", "Score",
		"R1M", "R3M", "R6M", "R12M", "Vol3M", "Vol6M", "ADV",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	rowCount := 0
	for rows.Next() {
		var rank *int
		var symbol, name, assetType string
		var score, r1m, r3m, r6m, r12m, vol3m, vol6m, adv *float64

		err := rows.Scan(
			&rank, &symbol, &name, &assetType, &score,
			&r1m, &r3m, &r6m, &r12m, &vol3m, &vol6m, &adv,
		)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		row := []string{
			formatInt(rank),
			symbol,
			name,
			assetType,
			formatFloat(score, 3),
			formatPercent(r1m),
			formatPercent(r3m),
			formatPercent(r6m),
			formatPercent(r12m),
			formatPercent(vol3m),
			formatPercent(vol6m),
			formatFloat(adv, 0),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	return filename, nil
}

// ExportRuns exports run metadata to a CSV file.
// Filename format: runs-YYYYMMDD.csv
func (e *Exporter) ExportRuns() (string, error) {
	if err := e.ensureExportDir(); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Query all runs
	query := `
		SELECT
			run_id,
			started_at,
			finished_at,
			status,
			symbols_processed,
			symbols_failed,
			notes
		FROM runs
		ORDER BY started_at DESC
	`

	rows, err := e.database.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	// Create output file
	dateStr := time.Now().Format("20060102")
	filename := filepath.Join(e.exportDir, fmt.Sprintf("runs-%s.csv", dateStr))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"RunID", "StartedAt", "FinishedAt", "Status",
		"SymbolsProcessed", "SymbolsFailed", "Notes",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for rows.Next() {
		var runID int64
		var startedAt, finishedAt, status *string
		var symbolsProcessed, symbolsFailed *int
		var notes *string

		err := rows.Scan(
			&runID, &startedAt, &finishedAt, &status,
			&symbolsProcessed, &symbolsFailed, &notes,
		)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		row := []string{
			fmt.Sprintf("%d", runID),
			formatString(startedAt),
			formatString(finishedAt),
			formatString(status),
			formatIntPtr(symbolsProcessed),
			formatIntPtr(symbolsFailed),
			formatString(notes),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	return filename, nil
}

// ExportSymbolDetail exports detailed metrics for a specific symbol.
// Filename format: symbol-SYMBOL-YYYYMMDD.csv
func (e *Exporter) ExportSymbolDetail(symbol string) (string, error) {
	if err := e.ensureExportDir(); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Query price history and indicators for the symbol
	query := `
		SELECT
			p.date,
			p.open,
			p.high,
			p.low,
			p.close,
			p.adj_close,
			p.volume,
			i.r_1m,
			i.r_3m,
			i.r_6m,
			i.r_12m,
			i.vol_3m,
			i.vol_6m,
			i.adv,
			i.score,
			i.rank
		FROM prices p
		LEFT JOIN indicators i ON p.symbol = i.symbol AND p.date = i.date
		WHERE p.symbol = ?
		ORDER BY p.date DESC
		LIMIT 365
	`

	rows, err := e.database.Query(query, symbol)
	if err != nil {
		return "", fmt.Errorf("failed to query symbol detail: %w", err)
	}
	defer rows.Close()

	// Create output file
	dateStr := time.Now().Format("20060102")
	filename := filepath.Join(e.exportDir, fmt.Sprintf("symbol-%s-%s.csv", symbol, dateStr))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Date", "Open", "High", "Low", "Close", "AdjClose", "Volume",
		"R1M", "R3M", "R6M", "R12M", "Vol3M", "Vol6M", "ADV", "Score", "Rank",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for rows.Next() {
		var date string
		var open, high, low, close, adjClose *float64
		var volume *int64
		var r1m, r3m, r6m, r12m, vol3m, vol6m, adv, score *float64
		var rank *int

		err := rows.Scan(
			&date, &open, &high, &low, &close, &adjClose, &volume,
			&r1m, &r3m, &r6m, &r12m, &vol3m, &vol6m, &adv, &score, &rank,
		)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		row := []string{
			date,
			formatFloat(open, 2),
			formatFloat(high, 2),
			formatFloat(low, 2),
			formatFloat(close, 2),
			formatFloat(adjClose, 2),
			formatInt64(volume),
			formatPercent(r1m),
			formatPercent(r3m),
			formatPercent(r6m),
			formatPercent(r12m),
			formatPercent(vol3m),
			formatPercent(vol6m),
			formatFloat(adv, 0),
			formatFloat(score, 3),
			formatInt(rank),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	return filename, nil
}

// Helper functions for formatting values

func formatFloat(val *float64, precision int) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%.*f", precision, *val)
}

func formatPercent(val *float64) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%.2f%%", *val*100)
}

func formatInt(val *int) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%d", *val)
}

func formatIntPtr(val *int) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%d", *val)
}

func formatInt64(val *int64) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%d", *val)
}

func formatString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
