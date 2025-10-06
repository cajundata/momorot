// Package config provides configuration management for the Momentum Screener TUI.
// It uses Viper to load settings from YAML files and environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration.
type Config struct {
	AlphaVantage AlphaVantageConfig `mapstructure:"alpha_vantage"`
	Universe     []string           `mapstructure:"universe"`
	Lookbacks    LookbacksConfig    `mapstructure:"lookbacks"`
	VolWindows   VolWindowsConfig   `mapstructure:"vol_windows"`
	Scoring      ScoringConfig      `mapstructure:"scoring"`
	Data         DataConfig         `mapstructure:"data"`
	App          AppConfig          `mapstructure:"app"`
	Fetcher      FetcherConfig      `mapstructure:"fetcher"`
}

// AlphaVantageConfig contains Alpha Vantage API settings.
type AlphaVantageConfig struct {
	APIKey            string `mapstructure:"api_key"`
	DailyRequestLimit int    `mapstructure:"daily_request_limit"`
	BaseURL           string `mapstructure:"base_url"`
}

// LookbacksConfig defines momentum lookback periods in trading days.
type LookbacksConfig struct {
	R1M  int `mapstructure:"r1m"`
	R3M  int `mapstructure:"r3m"`
	R6M  int `mapstructure:"r6m"`
	R12M int `mapstructure:"r12m"`
}

// VolWindowsConfig defines volatility calculation windows in trading days.
type VolWindowsConfig struct {
	Short int `mapstructure:"short"`
	Long  int `mapstructure:"long"`
}

// ScoringConfig contains momentum scoring parameters.
type ScoringConfig struct {
	PenaltyLambda       float64 `mapstructure:"penalty_lambda"`
	MinADVUSD           float64 `mapstructure:"min_adv_usd"`
	BreadthMinPositive  int     `mapstructure:"breadth_min_positive"`
	BreadthTotalLookbacks int   `mapstructure:"breadth_total_lookbacks"`
}

// DataConfig contains data storage settings.
type DataConfig struct {
	DataDir   string `mapstructure:"data_dir"`
	DBName    string `mapstructure:"db_name"`
	ExportDir string `mapstructure:"export_dir"`
}

// AppConfig contains application-level settings.
type AppConfig struct {
	TopN       int    `mapstructure:"top_n"`
	AutoExport bool   `mapstructure:"auto_export"`
	LogLevel   string `mapstructure:"log_level"`
	LogFormat  string `mapstructure:"log_format"`
}

// FetcherConfig contains data fetching settings.
type FetcherConfig struct {
	MaxWorkers          int  `mapstructure:"max_workers"`
	Timeout             int  `mapstructure:"timeout"`
	MaxRetries          int  `mapstructure:"max_retries"`
	ExponentialBackoff  bool `mapstructure:"exponential_backoff"`
	InitialRetryDelay   int  `mapstructure:"initial_retry_delay"`
	CacheEnabled        bool `mapstructure:"cache_enabled"`
	OnlyFetchDeltas     bool `mapstructure:"only_fetch_deltas"`
}

// Load loads the configuration from the specified file path or default locations.
// It supports environment variable overrides with the MOMOROT_ prefix.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure Viper
	v.SetConfigType("yaml")

	// If a specific config path is provided, use it
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Otherwise, search in default locations
		v.SetConfigName("config")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")

		// Also check for config in home directory
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".momorot"))
		}
	}

	// Enable environment variable overrides
	v.SetEnvPrefix("MOMOROT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Special handling for API key from ALPHAVANTAGE_API_KEY env var
	if apiKey := os.Getenv("ALPHAVANTAGE_API_KEY"); apiKey != "" {
		v.Set("alpha_vantage.api_key", apiKey)
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we can work with env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults sets default values for optional configuration parameters.
func setDefaults(v *viper.Viper) {
	// Alpha Vantage defaults
	v.SetDefault("alpha_vantage.daily_request_limit", 25)
	v.SetDefault("alpha_vantage.base_url", "https://www.alphavantage.co/query")

	// Lookback periods (trading days)
	v.SetDefault("lookbacks.r1m", 21)
	v.SetDefault("lookbacks.r3m", 63)
	v.SetDefault("lookbacks.r6m", 126)
	v.SetDefault("lookbacks.r12m", 252)

	// Volatility windows (trading days)
	v.SetDefault("vol_windows.short", 63)
	v.SetDefault("vol_windows.long", 126)

	// Scoring parameters
	v.SetDefault("scoring.penalty_lambda", 0.35)
	v.SetDefault("scoring.min_adv_usd", 5000000.0) // $5M
	v.SetDefault("scoring.breadth_min_positive", 3)
	v.SetDefault("scoring.breadth_total_lookbacks", 4)

	// Data storage
	v.SetDefault("data.data_dir", "./data")
	v.SetDefault("data.db_name", "momentum.db")
	v.SetDefault("data.export_dir", "./exports")

	// Application settings
	v.SetDefault("app.top_n", 5)
	v.SetDefault("app.auto_export", true)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.log_format", "json")

	// Fetcher settings
	v.SetDefault("fetcher.max_workers", 5)
	v.SetDefault("fetcher.timeout", 30)
	v.SetDefault("fetcher.max_retries", 3)
	v.SetDefault("fetcher.exponential_backoff", true)
	v.SetDefault("fetcher.initial_retry_delay", 1)
	v.SetDefault("fetcher.cache_enabled", true)
	v.SetDefault("fetcher.only_fetch_deltas", true)
}

// validate checks that all required configuration fields are present and valid.
func validate(cfg *Config) error {
	// Validate API key
	if cfg.AlphaVantage.APIKey == "" || cfg.AlphaVantage.APIKey == "YOUR_API_KEY_HERE" {
		return fmt.Errorf("alpha_vantage.api_key is required (set via config file or ALPHAVANTAGE_API_KEY env var)")
	}

	// Validate universe
	if len(cfg.Universe) == 0 {
		return fmt.Errorf("universe must contain at least one symbol")
	}

	// Validate daily request limit
	if cfg.AlphaVantage.DailyRequestLimit < 1 {
		return fmt.Errorf("alpha_vantage.daily_request_limit must be at least 1")
	}

	// Validate lookback periods
	if cfg.Lookbacks.R1M < 1 || cfg.Lookbacks.R3M < 1 || cfg.Lookbacks.R6M < 1 || cfg.Lookbacks.R12M < 1 {
		return fmt.Errorf("all lookback periods must be positive")
	}

	// Validate volatility windows
	if cfg.VolWindows.Short < 1 || cfg.VolWindows.Long < 1 {
		return fmt.Errorf("volatility windows must be positive")
	}

	// Validate scoring parameters
	if cfg.Scoring.PenaltyLambda < 0 || cfg.Scoring.PenaltyLambda > 1 {
		return fmt.Errorf("scoring.penalty_lambda must be between 0 and 1")
	}
	if cfg.Scoring.MinADVUSD < 0 {
		return fmt.Errorf("scoring.min_adv_usd must be non-negative")
	}
	if cfg.Scoring.BreadthMinPositive < 0 || cfg.Scoring.BreadthTotalLookbacks < 1 {
		return fmt.Errorf("breadth filter parameters must be positive")
	}
	if cfg.Scoring.BreadthMinPositive > cfg.Scoring.BreadthTotalLookbacks {
		return fmt.Errorf("breadth_min_positive cannot exceed breadth_total_lookbacks")
	}

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cfg.App.LogLevel] {
		return fmt.Errorf("app.log_level must be one of: debug, info, warn, error")
	}

	// Validate log format
	if cfg.App.LogFormat != "json" && cfg.App.LogFormat != "text" {
		return fmt.Errorf("app.log_format must be either 'json' or 'text'")
	}

	// Validate fetcher settings
	if cfg.Fetcher.MaxWorkers < 1 {
		return fmt.Errorf("fetcher.max_workers must be at least 1")
	}
	if cfg.Fetcher.Timeout < 1 {
		return fmt.Errorf("fetcher.timeout must be at least 1 second")
	}
	if cfg.Fetcher.MaxRetries < 0 {
		return fmt.Errorf("fetcher.max_retries must be non-negative")
	}
	if cfg.Fetcher.InitialRetryDelay < 1 {
		return fmt.Errorf("fetcher.initial_retry_delay must be at least 1 second")
	}

	return nil
}

// DBPath returns the full path to the database file.
func (c *Config) DBPath() string {
	return filepath.Join(c.Data.DataDir, c.Data.DBName)
}

// GetLookbackPeriods returns all lookback periods as a slice.
func (c *Config) GetLookbackPeriods() []int {
	return []int{c.Lookbacks.R1M, c.Lookbacks.R3M, c.Lookbacks.R6M, c.Lookbacks.R12M}
}
