package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCalendar(t *testing.T) {
	cal := NewCalendar()
	assert.NotNil(t, cal)
	assert.NotNil(t, cal.holidays)
	assert.Greater(t, len(cal.holidays), 50) // Should have many holidays loaded
}

func TestIsBusinessDay_Weekdays(t *testing.T) {
	cal := NewCalendar()

	// Tuesday, Jan 2, 2024 (not a holiday)
	tuesday := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	assert.True(t, cal.IsBusinessDay(tuesday))

	// Wednesday, Jan 3, 2024
	wednesday := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	assert.True(t, cal.IsBusinessDay(wednesday))

	// Thursday, Jan 4, 2024
	thursday := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
	assert.True(t, cal.IsBusinessDay(thursday))

	// Friday, Jan 5, 2024
	friday := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	assert.True(t, cal.IsBusinessDay(friday))

	// Monday, Jan 8, 2024
	monday := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	assert.True(t, cal.IsBusinessDay(monday))
}

func TestIsBusinessDay_Weekends(t *testing.T) {
	cal := NewCalendar()

	// Saturday, Jan 6, 2024
	saturday := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	assert.False(t, cal.IsBusinessDay(saturday))

	// Sunday, Jan 7, 2024
	sunday := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	assert.False(t, cal.IsBusinessDay(sunday))
}

func TestIsBusinessDay_Holidays(t *testing.T) {
	cal := NewCalendar()

	// Golden vectors: known NYSE holidays
	holidays := []struct {
		date time.Time
		name string
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "New Year's Day 2024"},
		{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), "MLK Day 2024"},
		{time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC), "Presidents' Day 2024"},
		{time.Date(2024, 3, 29, 0, 0, 0, 0, time.UTC), "Good Friday 2024"},
		{time.Date(2024, 5, 27, 0, 0, 0, 0, time.UTC), "Memorial Day 2024"},
		{time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC), "Juneteenth 2024"},
		{time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC), "Independence Day 2024"},
		{time.Date(2024, 9, 2, 0, 0, 0, 0, time.UTC), "Labor Day 2024"},
		{time.Date(2024, 11, 28, 0, 0, 0, 0, time.UTC), "Thanksgiving 2024"},
		{time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC), "Christmas 2024"},
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "New Year's Day 2025"},
		{time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC), "Independence Day 2025"},
	}

	for _, tc := range holidays {
		assert.False(t, cal.IsBusinessDay(tc.date), "Expected %s to be a holiday", tc.name)
		assert.True(t, cal.IsHoliday(tc.date), "Expected %s to be detected as holiday", tc.name)
	}
}

func TestNextBusinessDay(t *testing.T) {
	cal := NewCalendar()

	tests := []struct {
		input    time.Time
		expected time.Time
		name     string
	}{
		{
			// Friday -> Monday (skip weekend)
			input:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			name:     "Friday to Monday",
		},
		{
			// Thursday before New Year's (Friday is holiday) -> Monday
			input:    time.Date(2023, 12, 28, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 12, 29, 0, 0, 0, 0, time.UTC),
			name:     "Thursday to Friday",
		},
		{
			// Day before Memorial Day -> Day after Memorial Day (skips holiday + weekend)
			input:    time.Date(2024, 5, 24, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 5, 28, 0, 0, 0, 0, time.UTC),
			name:     "Before Memorial Day weekend",
		},
	}

	for _, tc := range tests {
		result := cal.NextBusinessDay(tc.input)
		assert.Equal(t, tc.expected, result, tc.name)
	}
}

func TestPreviousBusinessDay(t *testing.T) {
	cal := NewCalendar()

	tests := []struct {
		input    time.Time
		expected time.Time
		name     string
	}{
		{
			// Monday -> Friday (skip weekend)
			input:    time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			name:     "Monday to Friday",
		},
		{
			// Tuesday after New Year's -> Thursday before
			input:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 12, 29, 0, 0, 0, 0, time.UTC),
			name:     "After New Year's",
		},
	}

	for _, tc := range tests {
		result := cal.PreviousBusinessDay(tc.input)
		assert.Equal(t, tc.expected, result, tc.name)
	}
}

func TestAddBusinessDays(t *testing.T) {
	cal := NewCalendar()

	tests := []struct {
		start    time.Time
		days     int
		expected time.Time
		name     string
	}{
		{
			// Add 0 days
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			days:     0,
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			name:     "Add 0 days",
		},
		{
			// Add 1 business day (Tue -> Wed)
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			days:     1,
			expected: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			name:     "Add 1 day",
		},
		{
			// Add 5 business days (Mon -> Mon, skipping weekend)
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			days:     5,
			expected: time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
			name:     "Add 5 days (skip weekend)",
		},
		{
			// Add 21 business days (~1 month)
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			days:     21,
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			name:     "Add 21 days (~1 month)",
		},
		{
			// Subtract 1 business day
			start:    time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			days:     -1,
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			name:     "Subtract 1 day",
		},
		{
			// Subtract 5 business days (Mon -> Mon, skipping weekend)
			start:    time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
			days:     -5,
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			name:     "Subtract 5 days",
		},
	}

	for _, tc := range tests {
		result := cal.AddBusinessDays(tc.start, tc.days)
		assert.Equal(t, tc.expected, result, tc.name)
	}
}

func TestCountBusinessDays(t *testing.T) {
	cal := NewCalendar()

	tests := []struct {
		start    time.Time
		end      time.Time
		expected int
		name     string
	}{
		{
			// Same day
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: 0,
			name:     "Same day",
		},
		{
			// Mon -> Tue (1 day)
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			expected: 1,
			name:     "Mon to Tue",
		},
		{
			// Mon -> next Mon (5 business days, skipping weekend)
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
			expected: 5,
			name:     "Full week",
		},
		{
			// Entire month of January 2024 (21 business days)
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 21, // Jan 1 is holiday (Mon), 31 days - 8 weekend days - 2 holidays (Jan 1, Jan 15) = 21
			name:     "January 2024",
		},
		{
			// End before start
			start:    time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: 0,
			name:     "End before start",
		},
	}

	for _, tc := range tests {
		result := cal.CountBusinessDays(tc.start, tc.end)
		assert.Equal(t, tc.expected, result, tc.name)
	}
}

func TestGetLastBusinessDayOfMonth(t *testing.T) {
	cal := NewCalendar()

	tests := []struct {
		year     int
		month    time.Month
		expected time.Time
		name     string
	}{
		{
			// January 2024 - last day is Wednesday
			year:     2024,
			month:    time.January,
			expected: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			name:     "January 2024",
		},
		{
			// February 2024 - last day is Thursday
			year:     2024,
			month:    time.February,
			expected: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			name:     "February 2024 (leap year)",
		},
		{
			// June 2024 - 30th is Sunday, so last business day is Friday 28th
			year:     2024,
			month:    time.June,
			expected: time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC),
			name:     "June 2024 (ends on weekend)",
		},
		{
			// December 2024 - 31st is Tuesday
			year:     2024,
			month:    time.December,
			expected: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			name:     "December 2024",
		},
	}

	for _, tc := range tests {
		result := cal.GetLastBusinessDayOfMonth(tc.year, tc.month)
		assert.Equal(t, tc.expected, result, tc.name)
	}
}

func TestDefaultCalendarFunctions(t *testing.T) {
	// Test that package-level functions work
	monday := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)

	assert.True(t, IsBusinessDay(monday))
	assert.False(t, IsHoliday(monday))

	next := NextBusinessDay(monday)
	assert.Equal(t, time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC), next)

	prev := PreviousBusinessDay(monday)
	assert.Equal(t, time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), prev)

	future := AddBusinessDays(monday, 5)
	assert.Equal(t, time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), future)

	count := CountBusinessDays(monday, future)
	assert.Equal(t, 5, count)
}

func TestIsHoliday_NonHolidays(t *testing.T) {
	cal := NewCalendar()

	// Regular weekdays that are NOT holidays
	nonHolidays := []time.Time{
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),  // Tuesday
		time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), // Friday
		time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), // Monday
		time.Date(2024, 8, 20, 0, 0, 0, 0, time.UTC), // Tuesday
	}

	for _, date := range nonHolidays {
		assert.False(t, cal.IsHoliday(date), "Expected %s to NOT be a holiday", date.Format("2006-01-02"))
	}
}

// Benchmark calendar operations
func BenchmarkIsBusinessDay(b *testing.B) {
	cal := NewCalendar()
	date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cal.IsBusinessDay(date)
	}
}

func BenchmarkAddBusinessDays(b *testing.B) {
	cal := NewCalendar()
	date := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cal.AddBusinessDays(date, 21)
	}
}

func BenchmarkCountBusinessDays(b *testing.B) {
	cal := NewCalendar()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cal.CountBusinessDays(start, end)
	}
}
