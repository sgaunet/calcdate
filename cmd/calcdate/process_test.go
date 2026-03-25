package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	calcdate "github.com/sgaunet/calcdate/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests for pure functions ---

func TestFormatOutput(t *testing.T) {
	dt := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		format   string
		expected string
	}{
		{"iso", "2024-03-15T14:30:45Z"},
		{"sql", "2024-03-15 14:30:45"},
		{"ts", "1710513045"},
		{"human", "Friday, March 15, 2024"},
		{"compact", "20240315"},
		{"", "2024-03-15 14:30:45"},                    // default = sql
		{"%Y-%m-%d", "2024-03-15"},                     // unix format
		{"%Y/%m/%d %H:%M:%S", "2024/03/15 14:30:45"},  // unix custom
		{"2006/01/02", "2024/03/15"},                    // go format passthrough
	}

	for _, tc := range tests {
		t.Run("format_"+tc.format, func(t *testing.T) {
			result := formatOutput(dt, tc.format, time.UTC)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatOutput_Timezone(t *testing.T) {
	dt := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	result := formatOutput(dt, "sql", ny)
	assert.Equal(t, "2024-03-15 10:30:45", result) // UTC-4 in March (EDT)
}

func TestFormatOutput_NilTimezone(t *testing.T) {
	dt := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	result := formatOutput(dt, "sql", nil)
	assert.Equal(t, "2024-03-15 14:30:45", result)
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		date     time.Time
		weekend  bool
	}{
		{time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC), false}, // Monday
		{time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC), false}, // Tuesday
		{time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC), false}, // Wednesday
		{time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC), false}, // Thursday
		{time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), false}, // Friday
		{time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC), true},  // Saturday
		{time.Date(2024, 3, 17, 0, 0, 0, 0, time.UTC), true},  // Sunday
	}

	for _, tc := range tests {
		t.Run(tc.date.Weekday().String(), func(t *testing.T) {
			assert.Equal(t, tc.weekend, isWeekend(tc.date))
		})
	}
}

func TestIsSpecialInterval(t *testing.T) {
	tests := []struct {
		interval string
		special  bool
	}{
		{"1M", true},
		{"2Y", true},
		{"1q", true},
		{"1d", false},
		{"3h", false},
		{"30m", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.interval, func(t *testing.T) {
			assert.Equal(t, tc.special, isSpecialInterval(tc.interval))
		})
	}
}

func TestParseTransformIfProvided(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		node, err := parseTransformIfProvided("")
		require.NoError(t, err)
		assert.Nil(t, node)
	})

	t.Run("valid transform", func(t *testing.T) {
		node, err := parseTransformIfProvided("$begin +8h, $end +20h")
		require.NoError(t, err)
		require.NotNil(t, node)
		assert.NotNil(t, node.BeginExpr)
		assert.NotNil(t, node.EndExpr)
	})

	t.Run("invalid transform", func(t *testing.T) {
		_, err := parseTransformIfProvided("invalid transform |||")
		assert.Error(t, err)
	})
}

func TestPrintFilteredResults(t *testing.T) {
	results := []calcdate.IterationResult{
		{BeginTime: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), EndTime: time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC)}, // Friday
		{BeginTime: time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC), EndTime: time.Date(2024, 3, 17, 0, 0, 0, 0, time.UTC)}, // Saturday
		{BeginTime: time.Date(2024, 3, 17, 0, 0, 0, 0, time.UTC), EndTime: time.Date(2024, 3, 18, 0, 0, 0, 0, time.UTC)}, // Sunday
		{BeginTime: time.Date(2024, 3, 18, 0, 0, 0, 0, time.UTC), EndTime: time.Date(2024, 3, 19, 0, 0, 0, 0, time.UTC)}, // Monday
	}

	t.Run("without skip weekends", func(t *testing.T) {
		output := captureStdout(func() {
			printFilteredResults(results, "compact", time.UTC, false)
		})
		lines := nonEmptyLines(output)
		assert.Len(t, lines, 4)
	})

	t.Run("with skip weekends", func(t *testing.T) {
		output := captureStdout(func() {
			printFilteredResults(results, "compact", time.UTC, true)
		})
		lines := nonEmptyLines(output)
		assert.Len(t, lines, 2) // only Friday and Monday
		assert.Contains(t, lines[0], "20240315")
		assert.Contains(t, lines[1], "20240318")
	})
}

func TestPrintIterationResult(t *testing.T) {
	result := calcdate.IterationResult{
		BeginTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	output := captureStdout(func() {
		printIterationResult(result, "compact", time.UTC)
	})

	assert.Equal(t, "20240101 - 20240102\n", output)
}

// --- helpers ---

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}

func nonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
