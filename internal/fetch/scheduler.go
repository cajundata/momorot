package fetch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FetchResult represents the result of a fetch operation.
type FetchResult struct {
	Symbol    string
	Success   bool
	Error     error
	RecordsFetched int
	Duration  time.Duration
	Timestamp time.Time
}

// FetchTask represents a task to fetch data for a symbol.
type FetchTask struct {
	Symbol     string
	OutputSize string // "compact" or "full"
}

// Scheduler manages concurrent fetching of data across multiple symbols.
type Scheduler struct {
	avClient    *AlphaVantageClient
	maxWorkers  int
	resultsChan chan FetchResult
}

// NewScheduler creates a new fetch scheduler.
func NewScheduler(avClient *AlphaVantageClient, maxWorkers int) *Scheduler {
	return &Scheduler{
		avClient:    avClient,
		maxWorkers:  maxWorkers,
		resultsChan: make(chan FetchResult, 100),
	}
}

// FetchSymbols fetches data for multiple symbols concurrently.
// It respects the rate limiter and uses a worker pool to control concurrency.
func (s *Scheduler) FetchSymbols(ctx context.Context, symbols []string, outputSize string) ([]FetchResult, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	// Create task queue
	tasks := make(chan FetchTask, len(symbols))
	for _, symbol := range symbols {
		tasks <- FetchTask{
			Symbol:     symbol,
			OutputSize: outputSize,
		}
	}
	close(tasks)

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, tasks)
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(s.resultsChan)
	}()

	// Collect results
	var results []FetchResult
	for result := range s.resultsChan {
		results = append(results, result)
	}

	return results, nil
}

// worker processes fetch tasks from the queue.
func (s *Scheduler) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan FetchTask) {
	defer wg.Done()

	for task := range tasks {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			s.resultsChan <- FetchResult{
				Symbol:    task.Symbol,
				Success:   false,
				Error:     ctx.Err(),
				Timestamp: time.Now(),
			}
			continue
		default:
		}

		// Fetch data
		result := s.fetchSymbol(task)
		s.resultsChan <- result
	}
}

// fetchSymbol fetches data for a single symbol using FREE TIER endpoint.
func (s *Scheduler) fetchSymbol(task FetchTask) FetchResult {
	startTime := time.Now()

	data, err := s.avClient.FetchDaily(task.Symbol, task.OutputSize)

	duration := time.Since(startTime)

	if err != nil {
		return FetchResult{
			Symbol:    task.Symbol,
			Success:   false,
			Error:     err,
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	return FetchResult{
		Symbol:         task.Symbol,
		Success:        true,
		Error:          nil,
		RecordsFetched: len(data.TimeSeries),
		Duration:       duration,
		Timestamp:      time.Now(),
	}
}

// GetRateLimiterStatus returns the current rate limiter status.
func (s *Scheduler) GetRateLimiterStatus() *RateLimiterStatus {
	return s.avClient.GetRateLimiterStatus()
}

// PrioritizeFetchOrder determines which symbols to fetch first based on staleness.
// Returns symbols ordered by priority (most stale first).
func PrioritizeFetchOrder(symbols []string, lastFetchTimes map[string]time.Time) []string {
	type symbolPriority struct {
		symbol      string
		lastFetched time.Time
	}

	// Build priority list
	priorities := make([]symbolPriority, 0, len(symbols))
	for _, symbol := range symbols {
		lastFetched, exists := lastFetchTimes[symbol]
		if !exists {
			// Never fetched - highest priority (use zero time)
			lastFetched = time.Time{}
		}
		priorities = append(priorities, symbolPriority{
			symbol:      symbol,
			lastFetched: lastFetched,
		})
	}

	// Sort by staleness (oldest first)
	// Using a simple bubble sort for small lists
	for i := 0; i < len(priorities)-1; i++ {
		for j := 0; j < len(priorities)-i-1; j++ {
			if priorities[j].lastFetched.After(priorities[j+1].lastFetched) {
				priorities[j], priorities[j+1] = priorities[j+1], priorities[j]
			}
		}
	}

	// Extract ordered symbols
	orderedSymbols := make([]string, len(priorities))
	for i, p := range priorities {
		orderedSymbols[i] = p.symbol
	}

	return orderedSymbols
}

// EstimateFetchTime estimates how long it will take to fetch N symbols.
// Assumes ~2 seconds per request including network latency and API processing.
func EstimateFetchTime(symbolCount int, maxWorkers int) time.Duration {
	if symbolCount == 0 {
		return 0
	}

	// Estimate 2 seconds per symbol
	avgRequestTime := 2 * time.Second

	// Calculate parallel execution time
	batchCount := (symbolCount + maxWorkers - 1) / maxWorkers
	estimatedTime := time.Duration(batchCount) * avgRequestTime

	return estimatedTime
}
