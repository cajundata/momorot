package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/cajundata/momorot/internal/analytics"
	"github.com/cajundata/momorot/internal/config"
	"github.com/cajundata/momorot/internal/db"
	"github.com/cajundata/momorot/internal/export"
	"github.com/cajundata/momorot/internal/fetch"
	"github.com/cajundata/momorot/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Version information (set via ldflags during build)
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	// Define subcommands
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	refreshCmd := flag.NewFlagSet("refresh", flag.ExitOnError)
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	pingCmd := flag.NewFlagSet("ping", flag.ExitOnError)

	// Common flags
	configPath := ""
	for _, fs := range []*flag.FlagSet{runCmd, refreshCmd, exportCmd, pingCmd} {
		fs.StringVar(&configPath, "config", "configs/config.yaml", "Path to configuration file")
	}

	// Export command flags
	exportType := exportCmd.String("type", "leaders", "Export type: leaders, rankings, runs, symbol")
	exportSymbol := exportCmd.String("symbol", "", "Symbol for symbol export")
	exportTopN := exportCmd.Int("top", 5, "Top N for leaders export")
	exportDate := exportCmd.String("date", "", "Date for export (YYYY-MM-DD), defaults to today")

	// Show usage if no subcommand provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommand
	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		runTUI(configPath)

	case "refresh":
		refreshCmd.Parse(os.Args[2:])
		runRefresh(configPath)

	case "export":
		exportCmd.Parse(os.Args[2:])
		runExport(configPath, *exportType, *exportSymbol, *exportTopN, *exportDate)

	case "ping":
		pingCmd.Parse(os.Args[2:])
		runPing(configPath)

	case "version", "--version", "-v":
		printVersion()

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints command usage information
func printUsage() {
	fmt.Printf(`Momentum Screener TUI - %s

USAGE:
    momo <command> [options]

COMMANDS:
    run         Launch the TUI application
    refresh     Refresh data and compute rankings
    export      Export data to CSV files
    ping        Health check (verify config and DB)
    version     Show version information
    help        Show this help message

RUN OPTIONS:
    -config string
        Path to configuration file (default: configs/config.yaml)

REFRESH OPTIONS:
    -config string
        Path to configuration file (default: configs/config.yaml)

EXPORT OPTIONS:
    -config string
        Path to configuration file (default: configs/config.yaml)
    -type string
        Export type: leaders, rankings, runs, symbol (default: leaders)
    -symbol string
        Symbol for symbol export (required for -type=symbol)
    -top int
        Top N for leaders export (default: 5)
    -date string
        Date for export (YYYY-MM-DD), defaults to today

PING OPTIONS:
    -config string
        Path to configuration file (default: configs/config.yaml)

EXAMPLES:
    # Launch TUI
    momo run

    # Launch TUI with custom config
    momo run -config /path/to/config.yaml

    # Refresh data
    momo refresh

    # Export top 5 leaders
    momo export -type leaders -top 5

    # Export full rankings
    momo export -type rankings

    # Export symbol detail
    momo export -type symbol -symbol SPY

    # Export runs history
    momo export -type runs

    # Health check
    momo ping

VERSION:
    %s (commit: %s, built: %s)

`, version, version, commit, buildDate)
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("momo version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("built: %s\n", buildDate)
}

// runTUI launches the Terminal UI application
func runTUI(configPath string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create Bubble Tea program
	model := ui.New(database, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run program in goroutine
	errChan := make(chan error, 1)
	go func() {
		if _, err := p.Run(); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for either completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("TUI error: %v", err)
		}
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v\n", sig)
		p.Quit()
		<-errChan // Wait for program to finish
	}

	fmt.Println("Goodbye!")
}

// runRefresh performs a data refresh operation
func runRefresh(configPath string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	fmt.Println("Starting data refresh...")
	startTime := time.Now()

	// Create run record
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("CLI refresh")
	if err != nil {
		log.Fatalf("Failed to create run: %v", err)
	}

	// Initialize Alpha Vantage client
	avClient := fetch.NewAlphaVantageClient(
		cfg.AlphaVantage.APIKey,
		cfg.AlphaVantage.BaseURL,
		cfg.AlphaVantage.DailyRequestLimit,
		time.Duration(cfg.Fetcher.Timeout)*time.Second,
		cfg.Fetcher.MaxRetries,
	)

	// Create scheduler for concurrent fetching
	scheduler := fetch.NewScheduler(avClient, cfg.Fetcher.MaxWorkers)

	// Get active symbols
	symbolRepo := db.NewSymbolRepository(database)
	activeSymbols, err := symbolRepo.ListActive()
	if err != nil {
		log.Fatalf("Failed to get active symbols: %v", err)
	}

	// Extract symbol names
	symbolNames := make([]string, len(activeSymbols))
	for i, sym := range activeSymbols {
		symbolNames[i] = sym.Symbol
	}

	// Fetch data with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results, err := scheduler.FetchSymbols(ctx, symbolNames, "compact")
	if err != nil {
		log.Fatalf("Failed to fetch symbols: %v", err)
	}

	// Process results
	successCount := 0
	failureCount := 0
	priceRepo := db.NewPriceRepository(database)
	fetchLogRepo := db.NewFetchLogRepository(database)

	for _, result := range results {
		if result.Success {
			fmt.Printf("Fetching %s... ✓\n", result.Symbol)
			successCount++

			// Fetch and store prices
			dailyData, err := avClient.FetchDailyAdjusted(result.Symbol, "compact")
			if err != nil {
				fmt.Printf("  ⚠ Failed to fetch data: %v\n", err)
				failureCount++
				successCount--

				// Log failure
				errMsg := err.Error()
				fetchLogRepo.Log(&db.FetchLog{
					RunID:   runID,
					Symbol:  result.Symbol,
					Rows:    0,
					OK:      false,
					Message: &errMsg,
				})
				continue
			}

			// Store prices in database
			stored := 0
			for dateStr, bar := range dailyData.TimeSeries {
				// Parse string values to numeric types
				open, _ := strconv.ParseFloat(bar.Open, 64)
				high, _ := strconv.ParseFloat(bar.High, 64)
				low, _ := strconv.ParseFloat(bar.Low, 64)
				close, _ := strconv.ParseFloat(bar.Close, 64)
				adjClose, _ := strconv.ParseFloat(bar.AdjustedClose, 64)
				volume, _ := strconv.ParseInt(bar.Volume, 10, 64)

				price := &db.Price{
					Symbol:   result.Symbol,
					Date:     dateStr,
					Open:     open,
					High:     high,
					Low:      low,
					Close:    close,
					AdjClose: &adjClose,
					Volume:   &volume,
				}

				if err := priceRepo.Create(price); err != nil {
					// Ignore duplicate errors (already exists)
					if err.Error() != "UNIQUE constraint failed: prices.symbol, prices.date" {
						fmt.Printf("  ⚠ Failed to store price: %v\n", err)
					}
				} else {
					stored++
				}
			}

			// Log success
			fetchLogRepo.Log(&db.FetchLog{
				RunID:  runID,
				Symbol: result.Symbol,
				Rows:   stored,
				OK:     true,
			})

			fmt.Printf("  Stored %d prices\n", stored)
		} else {
			fmt.Printf("Fetching %s... ✗ %v\n", result.Symbol, result.Error)
			failureCount++

			// Log failure
			errMsg := result.Error.Error()
			fetchLogRepo.Log(&db.FetchLog{
				RunID:   runID,
				Symbol:  result.Symbol,
				Rows:    0,
				OK:      false,
				Message: &errMsg,
			})
		}
	}

	// Compute analytics
	fmt.Println("\nComputing analytics...")
	orchestrator := analytics.NewOrchestrator(
		database,
		map[string]int{
			"r1m":  cfg.Lookbacks.R1M,
			"r3m":  cfg.Lookbacks.R3M,
			"r6m":  cfg.Lookbacks.R6M,
			"r12m": cfg.Lookbacks.R12M,
		},
		map[string]int{
			"short": cfg.VolWindows.Short,
			"long":  cfg.VolWindows.Long,
		},
		analytics.ScoringConfig{
			PenaltyLambda:      cfg.Scoring.PenaltyLambda,
			MinADV:             cfg.Scoring.MinADVUSD,
			BreadthMinPositive: cfg.Scoring.BreadthMinPositive,
			BreadthTotal:       cfg.Scoring.BreadthTotalLookbacks,
		},
	)

	if _, err := orchestrator.ComputeAllIndicators(time.Now()); err != nil {
		log.Fatalf("Failed to compute analytics: %v", err)
	}

	fmt.Println("  ✓ Analytics computed")

	// Update run status
	status := "OK"
	if failureCount > 0 {
		status = "ERROR"
	}
	if err := runRepo.Finish(runID, status, successCount, failureCount); err != nil {
		log.Printf("Warning: Failed to update run status: %v", err)
	}

	// Auto-export if configured
	if cfg.App.AutoExport {
		fmt.Println("\nExporting data...")
		exporter := export.New(database, cfg.Data.ExportDir)

		if filename, err := exporter.ExportLeaders(cfg.App.TopN, ""); err == nil {
			fmt.Printf("  ✓ Leaders: %s\n", filename)
		} else {
			fmt.Printf("  ✗ Leaders export failed: %v\n", err)
		}

		if filename, err := exporter.ExportFullRankings(""); err == nil {
			fmt.Printf("  ✓ Rankings: %s\n", filename)
		} else {
			fmt.Printf("  ✗ Rankings export failed: %v\n", err)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\nRefresh complete in %v\n", duration)
	fmt.Printf("  Success: %d symbols\n", successCount)
	fmt.Printf("  Failed: %d symbols\n", failureCount)
}

// runExport exports data to CSV files
func runExport(configPath, exportType, symbol string, topN int, date string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create exporter
	exporter := export.New(database, cfg.Data.ExportDir)

	var filename string
	switch exportType {
	case "leaders":
		filename, err = exporter.ExportLeaders(topN, date)
		if err != nil {
			log.Fatalf("Export failed: %v", err)
		}
		fmt.Printf("✓ Exported top %d leaders to: %s\n", topN, filename)

	case "rankings":
		filename, err = exporter.ExportFullRankings(date)
		if err != nil {
			log.Fatalf("Export failed: %v", err)
		}
		fmt.Printf("✓ Exported full rankings to: %s\n", filename)

	case "runs":
		filename, err = exporter.ExportRuns()
		if err != nil {
			log.Fatalf("Export failed: %v", err)
		}
		fmt.Printf("✓ Exported runs history to: %s\n", filename)

	case "symbol":
		if symbol == "" {
			log.Fatal("Symbol required for symbol export (use -symbol flag)")
		}
		filename, err = exporter.ExportSymbolDetail(symbol)
		if err != nil {
			log.Fatalf("Export failed: %v", err)
		}
		fmt.Printf("✓ Exported %s detail to: %s\n", symbol, filename)

	default:
		log.Fatalf("Unknown export type: %s (valid: leaders, rankings, runs, symbol)", exportType)
	}
}

// runPing performs a health check
func runPing(configPath string) {
	fmt.Println("Performing health check...")

	// Check config
	fmt.Printf("  Config file: %s\n", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("  ✗ Config load failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ Config loaded successfully")

	// Check API key
	if cfg.AlphaVantage.APIKey == "" || cfg.AlphaVantage.APIKey == "YOUR_API_KEY_HERE" {
		fmt.Println("  ⚠ Warning: API key not configured")
	} else {
		fmt.Println("  ✓ API key configured")
	}

	// Check database
	dbPath := filepath.Join(cfg.Data.DataDir, cfg.Data.DBName)
	fmt.Printf("  Database: %s\n", dbPath)

	database, err := initDatabase(cfg)
	if err != nil {
		fmt.Printf("  ✗ Database initialization failed: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("  ✓ Database accessible")

	// Check symbol count
	symbolRepo := db.NewSymbolRepository(database)
	activeSymbols, err := symbolRepo.ListActive()
	if err != nil {
		fmt.Printf("  ✗ Failed to query symbols: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Active symbols: %d\n", len(activeSymbols))

	// Check latest data
	var latestDate string
	err = database.QueryRow("SELECT COALESCE(MAX(date), 'no data') FROM prices").Scan(&latestDate)
	if err != nil {
		fmt.Printf("  ✗ Failed to query latest data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Latest price data: %s\n", latestDate)

	// Check export directory
	if _, err := os.Stat(cfg.Data.ExportDir); os.IsNotExist(err) {
		fmt.Printf("  ⚠ Export directory does not exist: %s\n", cfg.Data.ExportDir)
	} else {
		fmt.Printf("  ✓ Export directory: %s\n", cfg.Data.ExportDir)
	}

	fmt.Println("\n✓ Health check passed")
}

// initDatabase initializes and migrates the database
func initDatabase(cfg *config.Config) (*db.DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.Data.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(cfg.Data.DataDir, cfg.Data.DBName)
	database, err := db.New(db.Config{Path: dbPath})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure all configured symbols exist in database
	symbolRepo := db.NewSymbolRepository(database)
	for _, symbol := range cfg.Universe {
		// Check if symbol exists
		_, err := symbolRepo.Get(symbol)
		if err != nil {
			// Symbol doesn't exist, create it
			if err := symbolRepo.Create(&db.Symbol{
				Symbol:    symbol,
				Name:      symbol, // Use symbol as name initially
				AssetType: "ETF",  // Default to ETF
				Active:    true,
			}); err != nil {
				log.Printf("Warning: Failed to create symbol %s: %v", symbol, err)
			}
		}
	}

	return database, nil
}
