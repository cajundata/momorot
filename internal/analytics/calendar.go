// Package analytics provides momentum calculation, scoring, and ranking functionality.
package analytics

import (
	"time"
)

// Calendar handles NYSE trading day calculations and holiday detection.
type Calendar struct {
	holidays map[string]bool
}

// NewCalendar creates a new NYSE trading calendar with holidays for 2024-2030.
func NewCalendar() *Calendar {
	c := &Calendar{
		holidays: make(map[string]bool),
	}
	c.initializeHolidays()
	return c
}

// initializeHolidays populates the calendar with NYSE holidays for 2024-2030.
func (c *Calendar) initializeHolidays() {
	// Format: YYYY-MM-DD
	nyseHolidays := []string{
		// 2024
		"2024-01-01", // New Year's Day
		"2024-01-15", // Martin Luther King Jr. Day
		"2024-02-19", // Presidents' Day
		"2024-03-29", // Good Friday
		"2024-05-27", // Memorial Day
		"2024-06-19", // Juneteenth
		"2024-07-04", // Independence Day
		"2024-09-02", // Labor Day
		"2024-11-28", // Thanksgiving
		"2024-12-25", // Christmas

		// 2025
		"2025-01-01", // New Year's Day
		"2025-01-20", // Martin Luther King Jr. Day
		"2025-02-17", // Presidents' Day
		"2025-04-18", // Good Friday
		"2025-05-26", // Memorial Day
		"2025-06-19", // Juneteenth
		"2025-07-04", // Independence Day
		"2025-09-01", // Labor Day
		"2025-11-27", // Thanksgiving
		"2025-12-25", // Christmas

		// 2026
		"2026-01-01", // New Year's Day
		"2026-01-19", // Martin Luther King Jr. Day
		"2026-02-16", // Presidents' Day
		"2026-04-03", // Good Friday
		"2026-05-25", // Memorial Day
		"2026-06-19", // Juneteenth
		"2026-07-03", // Independence Day (observed, 7/4 is Saturday)
		"2026-09-07", // Labor Day
		"2026-11-26", // Thanksgiving
		"2026-12-25", // Christmas

		// 2027
		"2027-01-01", // New Year's Day
		"2027-01-18", // Martin Luther King Jr. Day
		"2027-02-15", // Presidents' Day
		"2027-03-26", // Good Friday
		"2027-05-31", // Memorial Day
		"2027-06-18", // Juneteenth (observed, 6/19 is Saturday)
		"2027-07-05", // Independence Day (observed, 7/4 is Sunday)
		"2027-09-06", // Labor Day
		"2027-11-25", // Thanksgiving
		"2027-12-24", // Christmas (observed, 12/25 is Saturday)

		// 2028
		"2028-01-17", // Martin Luther King Jr. Day
		"2028-02-21", // Presidents' Day
		"2028-04-14", // Good Friday
		"2028-05-29", // Memorial Day
		"2028-06-19", // Juneteenth
		"2028-07-04", // Independence Day
		"2028-09-04", // Labor Day
		"2028-11-23", // Thanksgiving
		"2028-12-25", // Christmas

		// 2029
		"2029-01-01", // New Year's Day
		"2029-01-15", // Martin Luther King Jr. Day
		"2029-02-19", // Presidents' Day
		"2029-03-30", // Good Friday
		"2029-05-28", // Memorial Day
		"2029-06-19", // Juneteenth
		"2029-07-04", // Independence Day
		"2029-09-03", // Labor Day
		"2029-11-22", // Thanksgiving
		"2029-12-25", // Christmas

		// 2030
		"2030-01-01", // New Year's Day
		"2030-01-21", // Martin Luther King Jr. Day
		"2030-02-18", // Presidents' Day
		"2030-04-19", // Good Friday
		"2030-05-27", // Memorial Day
		"2030-06-19", // Juneteenth
		"2030-07-04", // Independence Day
		"2030-09-02", // Labor Day
		"2030-11-28", // Thanksgiving
		"2030-12-25", // Christmas
	}

	for _, holiday := range nyseHolidays {
		c.holidays[holiday] = true
	}
}

// IsBusinessDay returns true if the date is a trading day (not weekend or holiday).
func (c *Calendar) IsBusinessDay(date time.Time) bool {
	// Check if weekend
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Check if holiday
	dateStr := date.Format("2006-01-02")
	if c.holidays[dateStr] {
		return false
	}

	return true
}

// IsHoliday returns true if the date is a NYSE holiday.
func (c *Calendar) IsHoliday(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	return c.holidays[dateStr]
}

// NextBusinessDay returns the next trading day after the given date.
func (c *Calendar) NextBusinessDay(date time.Time) time.Time {
	next := date.AddDate(0, 0, 1)
	for !c.IsBusinessDay(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// PreviousBusinessDay returns the previous trading day before the given date.
func (c *Calendar) PreviousBusinessDay(date time.Time) time.Time {
	prev := date.AddDate(0, 0, -1)
	for !c.IsBusinessDay(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// AddBusinessDays adds N business days to the given date.
// If n is negative, it subtracts business days.
func (c *Calendar) AddBusinessDays(date time.Time, n int) time.Time {
	if n == 0 {
		return date
	}

	direction := 1
	if n < 0 {
		direction = -1
		n = -n
	}

	result := date
	for i := 0; i < n; i++ {
		result = result.AddDate(0, 0, direction)
		for !c.IsBusinessDay(result) {
			result = result.AddDate(0, 0, direction)
		}
	}

	return result
}

// CountBusinessDays counts the number of business days between two dates (inclusive of start, exclusive of end).
func (c *Calendar) CountBusinessDays(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}

	count := 0
	current := start

	for current.Before(end) {
		if c.IsBusinessDay(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// GetLastBusinessDayOfMonth returns the last trading day of the given month.
func (c *Calendar) GetLastBusinessDayOfMonth(year int, month time.Month) time.Time {
	// Start with the last day of the month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)

	// Walk backwards until we find a business day
	for !c.IsBusinessDay(lastDay) {
		lastDay = lastDay.AddDate(0, 0, -1)
	}

	return lastDay
}

// Default calendar instance
var defaultCalendar = NewCalendar()

// IsBusinessDay checks if a date is a business day using the default calendar.
func IsBusinessDay(date time.Time) bool {
	return defaultCalendar.IsBusinessDay(date)
}

// IsHoliday checks if a date is a holiday using the default calendar.
func IsHoliday(date time.Time) bool {
	return defaultCalendar.IsHoliday(date)
}

// NextBusinessDay returns the next business day using the default calendar.
func NextBusinessDay(date time.Time) time.Time {
	return defaultCalendar.NextBusinessDay(date)
}

// PreviousBusinessDay returns the previous business day using the default calendar.
func PreviousBusinessDay(date time.Time) time.Time {
	return defaultCalendar.PreviousBusinessDay(date)
}

// AddBusinessDays adds N business days using the default calendar.
func AddBusinessDays(date time.Time, n int) time.Time {
	return defaultCalendar.AddBusinessDays(date, n)
}

// CountBusinessDays counts business days using the default calendar.
func CountBusinessDays(start, end time.Time) int {
	return defaultCalendar.CountBusinessDays(start, end)
}
