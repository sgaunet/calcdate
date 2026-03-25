package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/calcdate/v2"
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

// --- Integration tests via the compiled binary ---

func TestMain(m *testing.M) {
	// Build the binary once for integration tests
	build := exec.Command("go", "build", "-o", "calcdate_test_bin", ".")
	build.Dir = "."
	if err := build.Run(); err != nil {
		panic("failed to build test binary: " + err.Error())
	}
	code := m.Run()
	os.Remove("calcdate_test_bin")
	os.Exit(code)
}

func runBinary(args ...string) (string, string, error) {
	cmd := exec.Command("./calcdate_test_bin", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestCLI_SingleExpression(t *testing.T) {
	tests := []struct {
		desc   string
		args   []string
		expect string
	}{
		{"simple date", []string{"-x", "2024-01-15"}, "2024-01-15 00:00:00"},
		{"date plus day", []string{"-x", "2024-01-15 +1d"}, "2024-01-16 00:00:00"},
		{"pipeline", []string{"-x", "2024-03-15 | startofmonth"}, "2024-03-01 00:00:00"},
		{"with iso format", []string{"-x", "2024-01-15", "-f", "iso", "-tz", "UTC"}, "2024-01-15T00:00:00Z"},
		{"with compact format", []string{"-x", "2024-01-15", "-f", "compact"}, "20240115"},
		{"with human format", []string{"-x", "2024-01-15", "-f", "human"}, "Monday, January 15, 2024"},
		{"with ts format", []string{"-x", "2024-01-15", "-f", "ts", "-tz", "UTC"}, "1705276800"},
		{"with unix format", []string{"-x", "2024-01-15", "-f", "%Y/%m/%d"}, "2024/01/15"},
		{"explicit tz", []string{"-x", "2024-01-15", "-tz", "UTC"}, "2024-01-15 00:00:00"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			stdout, stderr, err := runBinary(tc.args...)
			require.NoError(t, err, "stderr: %s", stderr)
			assert.Equal(t, tc.expect+"\n", stdout)
		})
	}
}

func TestCLI_RangeExpression(t *testing.T) {
	t.Run("range with interval", func(t *testing.T) {
		stdout, _, err := runBinary("-x", "2024-01-01...2024-01-04", "-each", "1d", "-f", "compact")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		assert.Len(t, lines, 3)
		assert.Equal(t, "20240101 - 20240102", lines[0])
		assert.Equal(t, "20240102 - 20240103", lines[1])
		assert.Equal(t, "20240103 - 20240104", lines[2])
	})

	t.Run("range without interval", func(t *testing.T) {
		stdout, _, err := runBinary("-x", "2024-01-01...2024-01-31", "-f", "compact")
		require.NoError(t, err)
		assert.Equal(t, "20240101 - 20240131\n", stdout)
	})

	t.Run("range with monthly interval", func(t *testing.T) {
		stdout, _, err := runBinary("-x", "2024-01-01...2024-04-01", "-each", "1M", "-f", "compact")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		assert.Len(t, lines, 3)
	})

	t.Run("range with transform", func(t *testing.T) {
		stdout, _, err := runBinary("-x", "2024-01-01...2024-01-03", "-each", "1d",
			"-t", "$begin +8h, $end +20h", "-f", "sql")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		require.Len(t, lines, 2)
		assert.Contains(t, lines[0], "08:00:00")
		assert.Contains(t, lines[0], "20:00:00")
	})

	t.Run("range skip weekends", func(t *testing.T) {
		// 2024-03-11 (Mon) to 2024-03-18 (Mon), 1d intervals
		stdout, _, err := runBinary("-x", "2024-03-11...2024-03-18", "-each", "1d",
			"-f", "compact", "--skip-weekends")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		// 7 days total, 2 weekend days skipped → 5 lines
		assert.Len(t, lines, 5)
		// Each line starts with "BEGINDATE - ENDDATE"
		// No line should START with Saturday or Sunday
		for _, line := range lines {
			beginDate := line[:8] // first 8 chars = compact begin date
			assert.NotEqual(t, "20240316", beginDate) // Saturday
			assert.NotEqual(t, "20240317", beginDate) // Sunday
		}
	})
}

func TestCLI_Version(t *testing.T) {
	stdout, _, err := runBinary("-v")
	require.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(stdout))
}

func TestCLI_ListOps(t *testing.T) {
	stdout, _, err := runBinary("-list-ops")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available Operations:")
	assert.Contains(t, stdout, "startofmonth")
}

func TestCLI_ErrorCases(t *testing.T) {
	t.Run("invalid expression", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "|||")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Failed to parse expression")
	})

	t.Run("invalid timezone", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "today", "-tz", "Invalid/TZ")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Invalid timezone")
	})

	t.Run("invalid interval", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "2024-01-01...2024-01-31", "-each", "abc")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Failed")
	})

	t.Run("invalid transform", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "2024-01-01...2024-01-31", "-each", "1d", "-t", "|||")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Failed to parse transform")
	})
}

func TestCLI_StdinExpression(t *testing.T) {
	cmd := exec.Command("./calcdate_test_bin", "-f", "compact")
	cmd.Stdin = strings.NewReader("2024-06-15\n")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Equal(t, "20240615\n", stdout.String())
}

func TestCLI_StdinEmpty(t *testing.T) {
	cmd := exec.Command("./calcdate_test_bin")
	cmd.Stdin = strings.NewReader("")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, stderr.String(), "stdin")
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
