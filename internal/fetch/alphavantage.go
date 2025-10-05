// Package fetch provides data fetching functionality from various sources.
// It includes Alpha Vantage API client, CSV importers, and scheduling.
package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// AlphaVantageClient handles API interactions with Alpha Vantage.
type AlphaVantageClient struct {
	apiKey     string
	baseURL    string
	httpClient *retryablehttp.Client
	rateLimiter *RateLimiter
}

// NewAlphaVantageClient creates a new Alpha Vantage API client with rate limiting.
func NewAlphaVantageClient(apiKey, baseURL string, dailyLimit int, timeout time.Duration, maxRetries int) *AlphaVantageClient {
	// Create retryable HTTP client
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 10 * time.Second
	client.HTTPClient.Timeout = timeout
	client.Logger = nil // Disable default logging

	// Create rate limiter (25 requests per day for free tier)
	rateLimiter := NewRateLimiter(dailyLimit)

	return &AlphaVantageClient{
		apiKey:      apiKey,
		baseURL:     baseURL,
		httpClient:  client,
		rateLimiter: rateLimiter,
	}
}

// DailyAdjusted represents the TIME_SERIES_DAILY_ADJUSTED response structure.
type DailyAdjusted struct {
	MetaData   DailyMetaData          `json:"Meta Data"`
	TimeSeries map[string]DailyOHLCV  `json:"Time Series (Daily)"`
	ErrorMessage string                `json:"Error Message,omitempty"`
	Note         string                `json:"Note,omitempty"`
}

// DailyMetaData contains metadata from the API response.
type DailyMetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

// DailyOHLCV represents a single day's OHLCV data with adjusted close.
type DailyOHLCV struct {
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

// FetchDailyAdjusted fetches daily adjusted OHLCV data for a symbol.
// outputSize can be "compact" (100 days) or "full" (20+ years).
func (c *AlphaVantageClient) FetchDailyAdjusted(symbol, outputSize string) (*DailyAdjusted, error) {
	// Check rate limit before making request
	if err := c.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build URL
	params := url.Values{}
	params.Add("function", "TIME_SERIES_DAILY_ADJUSTED")
	params.Add("symbol", symbol)
	params.Add("outputsize", outputSize)
	params.Add("apikey", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	// Make request
	req, err := retryablehttp.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	// Record the request
	c.rateLimiter.RecordRequest()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var data DailyAdjusted
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response for %s: %w", symbol, err)
	}

	// Check for API errors
	if data.ErrorMessage != "" {
		return nil, fmt.Errorf("API error for %s: %s", symbol, data.ErrorMessage)
	}

	// Check for rate limit note
	if data.Note != "" {
		return nil, fmt.Errorf("API rate limit note for %s: %s", symbol, data.Note)
	}

	// Verify we got data
	if len(data.TimeSeries) == 0 {
		return nil, fmt.Errorf("no data returned for %s", symbol)
	}

	return &data, nil
}

// ParseOHLCV converts API string values to float64.
func ParseOHLCV(d DailyOHLCV) (open, high, low, close, adjClose, volume, dividend, split float64, err error) {
	open, err = strconv.ParseFloat(d.Open, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse open: %w", err)
	}

	high, err = strconv.ParseFloat(d.High, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse high: %w", err)
	}

	low, err = strconv.ParseFloat(d.Low, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse low: %w", err)
	}

	close, err = strconv.ParseFloat(d.Close, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse close: %w", err)
	}

	adjClose, err = strconv.ParseFloat(d.AdjustedClose, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse adjusted close: %w", err)
	}

	volume, err = strconv.ParseFloat(d.Volume, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse volume: %w", err)
	}

	dividend, err = strconv.ParseFloat(d.DividendAmount, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse dividend: %w", err)
	}

	split, err = strconv.ParseFloat(d.SplitCoefficient, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse split coefficient: %w", err)
	}

	return open, high, low, close, adjClose, volume, dividend, split, nil
}

// GetRateLimiterStatus returns the current status of the rate limiter.
func (c *AlphaVantageClient) GetRateLimiterStatus() *RateLimiterStatus {
	return c.rateLimiter.GetStatus()
}

// ResetRateLimiter manually resets the rate limiter (for testing or admin purposes).
func (c *AlphaVantageClient) ResetRateLimiter() {
	c.rateLimiter.Reset()
}
