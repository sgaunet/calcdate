package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		{"with iso format", []string{"-x", "2024-01-15", "-f", "iso", "--tz", "UTC"}, "2024-01-15T00:00:00Z"},
		{"with compact format", []string{"-x", "2024-01-15", "-f", "compact"}, "20240115"},
		{"with human format", []string{"-x", "2024-01-15", "-f", "human"}, "Monday, January 15, 2024"},
		{"with ts format", []string{"-x", "2024-01-15", "-f", "ts", "--tz", "UTC"}, "1705276800"},
		{"with unix format", []string{"-x", "2024-01-15", "-f", "%Y/%m/%d"}, "2024/01/15"},
		{"explicit tz", []string{"-x", "2024-01-15", "--tz", "UTC"}, "2024-01-15 00:00:00"},
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
		stdout, _, err := runBinary("-x", "2024-01-01...2024-01-04", "--each", "1d", "-f", "compact")
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
		stdout, _, err := runBinary("-x", "2024-01-01...2024-04-01", "--each", "1M", "-f", "compact")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		assert.Len(t, lines, 3)
	})

	t.Run("range with transform", func(t *testing.T) {
		stdout, _, err := runBinary("-x", "2024-01-01...2024-01-03", "--each", "1d",
			"-t", "$begin +8h, $end +20h", "-f", "sql")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		require.Len(t, lines, 2)
		assert.Contains(t, lines[0], "08:00:00")
		assert.Contains(t, lines[0], "20:00:00")
	})

	t.Run("range skip weekends", func(t *testing.T) {
		// 2024-03-11 (Mon) to 2024-03-18 (Mon), 1d intervals
		stdout, _, err := runBinary("-x", "2024-03-11...2024-03-18", "--each", "1d",
			"-f", "compact", "--skip-weekends")
		require.NoError(t, err)
		lines := nonEmptyLines(stdout)
		// 7 days total, 2 weekend days skipped -> 5 lines
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
	stdout, _, err := runBinary("--list-ops")
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
		_, stderr, err := runBinary("-x", "today", "--tz", "Invalid/TZ")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Invalid timezone")
	})

	t.Run("invalid interval", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "2024-01-01...2024-01-31", "--each", "abc")
		assert.Error(t, err)
		assert.Contains(t, stderr, "Failed")
	})

	t.Run("invalid transform", func(t *testing.T) {
		_, stderr, err := runBinary("-x", "2024-01-01...2024-01-31", "--each", "1d", "-t", "|||")
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

func TestCLI_Completion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			stdout, _, err := runBinary("completion", shell)
			require.NoError(t, err)
			assert.NotEmpty(t, stdout)
		})
	}
}

func TestCLI_CompletionInvalid(t *testing.T) {
	_, _, err := runBinary("completion", "invalid")
	assert.Error(t, err)
}
