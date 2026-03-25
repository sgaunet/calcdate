package calcdate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	testCases := []struct {
		input    string
		expected []TokenType
	}{
		{
			input:    "today +1d",
			expected: []TokenType{TokenKeyword, TokenUnit, TokenEOF},
		},
		{
			input:    "now | +2h | round hour",
			expected: []TokenType{TokenKeyword, TokenPipe, TokenUnit, TokenPipe, TokenKeyword, TokenKeyword, TokenEOF},
		},
		{
			input:    "$begin +8h",
			expected: []TokenType{TokenVariable, TokenUnit, TokenEOF},
		},
		{
			input:    "2024-01-15...2024-01-31",
			expected: []TokenType{TokenDate, TokenRange, TokenDate, TokenEOF},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			tokenizer := NewTokenizer(tc.input)
			tokens, err := tokenizer.Tokenize()
			require.NoError(t, err)

			types := make([]TokenType, len(tokens))
			for i, token := range tokens {
				types[i] = token.Type
			}

			assert.Equal(t, tc.expected, types)
		})
	}
}

func TestExprParser(t *testing.T) {
	testCases := []struct {
		input       string
		description string
		shouldError bool
	}{
		{
			input:       "today",
			description: "simple date keyword",
			shouldError: false,
		},
		{
			input:       "today +1d",
			description: "date with arithmetic",
			shouldError: false,
		},
		{
			input:       "now | +2h | round hour",
			description: "pipeline operations",
			shouldError: false,
		},
		{
			input:       "today...tomorrow",
			description: "simple range",
			shouldError: false,
		},
		{
			input:       "2024-01-15 +1M",
			description: "ISO date with arithmetic",
			shouldError: false,
		},
		{
			input:       "today + 1d | endOfMonth",
			description: "mixed syntax: arithmetic then pipe",
			shouldError: false,
		},
		{
			input:       "today +1d | -2h | endOfWeek",
			description: "mixed syntax: multiple operations",
			shouldError: false,
		},
		{
			input:       "2024-01-15 +1M | startOfMonth | +7d",
			description: "ISO date with mixed operations",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			parser := NewExprParser(tc.input)
			node, err := parser.Parse(tc.input)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
			}
		})
	}
}

func TestEvaluateExpression(t *testing.T) {

	testCases := []struct {
		input       string
		description string
		check       func(t *testing.T, result time.Time)
	}{
		{
			input:       "2024-01-15",
			description: "ISO date parsing",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 15, result.Day())
			},
		},
		{
			input:       "2024-01-15|+1d",
			description: "ISO date plus one day",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 16, result.Day())
			},
		},
		{
			input:       "2024-01-15 +1d | endOfMonth",
			description: "mixed syntax with endOfMonth",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 31, result.Day())
				assert.Equal(t, 23, result.Hour())
				assert.Equal(t, 59, result.Minute())
				assert.Equal(t, 59, result.Second())
			},
		},
		{
			input:       "2024-01-15 +1M | startOfMonth",
			description: "mixed syntax with startOfMonth",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.February, result.Month())
				assert.Equal(t, 1, result.Day())
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := EvaluateExpression(tc.input, time.UTC)
			require.NoError(t, err)
			tc.check(t, result)
		})
	}
}

func TestTransformParsing(t *testing.T) {
	parser := NewExprParser("")
	transform, err := parser.ParseTransform("$begin +8h, $end +20h")
	
	require.NoError(t, err)
	assert.NotNil(t, transform)
	assert.NotNil(t, transform.BeginExpr)
	assert.NotNil(t, transform.EndExpr)
}

func TestExprParserErrorSuggestions(t *testing.T) {
	tests := []struct {
		input           string
		description     string
		expectedContain string
	}{
		{
			input:           "today | startofmoonth",
			description:     "typo in startofmonth",
			expectedContain: "startofmonth",
		},
		{
			input:           "today | endofwek",
			description:     "typo in endofweek",
			expectedContain: "endofweek",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			parser := NewExprParser(tc.input)
			_, err := parser.Parse(tc.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrUnknownOperation)
			assert.Contains(t, err.Error(), tc.expectedContain)
		})
	}
}

func TestOperations(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	testCases := []struct {
		op          string
		value       string
		description string
		check       func(t *testing.T, result time.Time)
	}{
		{
			op:          "+",
			value:       "1d",
			description: "add one day",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 16, result.Day())
			},
		},
		{
			op:          "startofmonth",
			value:       "",
			description: "start of month",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 1, result.Day())
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
			},
		},
		{
			op:          "endofmonth",
			value:       "",
			description: "end of month",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 31, result.Day()) // January has 31 days
				assert.Equal(t, 23, result.Hour())
				assert.Equal(t, 59, result.Minute())
				assert.Equal(t, 59, result.Second())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := ApplyOperation(baseTime, tc.op, tc.value, time.UTC)
			require.NoError(t, err)
			tc.check(t, result)
		})
	}
}

func TestApplyOperation_AllBoundaries(t *testing.T) {
	const maxNanos = 999999999
	// Friday March 15, 2024 14:30:45.123456789 UTC
	base := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)

	tests := []struct {
		op       string
		expected time.Time
	}{
		{"startofday", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{"endofday", time.Date(2024, 3, 15, 23, 59, 59, maxNanos, time.UTC)},
		{"start", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{"end", time.Date(2024, 3, 15, 23, 59, 59, maxNanos, time.UTC)},
		{"startofweek", time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC)},         // Friday→Monday
		{"endofweek", time.Date(2024, 3, 17, 23, 59, 59, maxNanos, time.UTC)},  // Friday→Sunday
		{"startofmonth", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"endofmonth", time.Date(2024, 3, 31, 23, 59, 59, maxNanos, time.UTC)},
		{"startofyear", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"endofyear", time.Date(2024, 12, 31, 23, 59, 59, maxNanos, time.UTC)},
		{"startofquarter", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},         // March→Q1
		{"endofquarter", time.Date(2024, 3, 31, 23, 59, 59, maxNanos, time.UTC)}, // March→Q1 end
		{"startofhour", time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC)},
		{"endofhour", time.Date(2024, 3, 15, 14, 59, 59, maxNanos, time.UTC)},
		{"startofminute", time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)},
		{"endofminute", time.Date(2024, 3, 15, 14, 30, 59, maxNanos, time.UTC)},
		{"startofsecond", time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)},
		{"endofsecond", time.Date(2024, 3, 15, 14, 30, 45, maxNanos, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.op, func(t *testing.T) {
			result, err := ApplyOperation(base, tc.op, "", time.UTC)
			require.NoError(t, err)
			assert.True(t, tc.expected.Equal(result), "op=%s expected=%v got=%v", tc.op, tc.expected, result)
		})
	}
}

func TestApplyOperation_BoundaryEdgeCases(t *testing.T) {
	const maxNanos = 999999999

	t.Run("startofweek on sunday", func(t *testing.T) {
		sunday := time.Date(2024, 3, 17, 10, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(sunday, "startofweek", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC), result)
	})

	t.Run("startofweek on monday", func(t *testing.T) {
		monday := time.Date(2024, 3, 11, 15, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(monday, "startofweek", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC), result)
	})

	t.Run("startofquarter per quarter", func(t *testing.T) {
		quarters := []struct {
			month    time.Month
			expected time.Month
		}{
			{time.January, time.January}, {time.February, time.January}, {time.March, time.January},
			{time.April, time.April}, {time.May, time.April}, {time.June, time.April},
			{time.July, time.July}, {time.August, time.July}, {time.September, time.July},
			{time.October, time.October}, {time.November, time.October}, {time.December, time.October},
		}
		for _, q := range quarters {
			dt := time.Date(2024, q.month, 15, 0, 0, 0, 0, time.UTC)
			result, err := ApplyOperation(dt, "startofquarter", "", time.UTC)
			require.NoError(t, err)
			assert.Equal(t, q.expected, result.Month(), "month=%v", q.month)
			assert.Equal(t, 1, result.Day())
		}
	})

	t.Run("endofquarter Q2 june has 30 days", func(t *testing.T) {
		dt := time.Date(2024, time.May, 10, 0, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(dt, "endofquarter", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 30, result.Day())
	})

	t.Run("endofquarter Q4 december has 31 days", func(t *testing.T) {
		dt := time.Date(2024, time.November, 10, 0, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(dt, "endofquarter", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.December, result.Month())
		assert.Equal(t, 31, result.Day())
	})

	t.Run("endofmonth february leap year", func(t *testing.T) {
		dt := time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(dt, "endofmonth", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 29, result.Day())
		assert.Equal(t, 23, result.Hour())
		assert.Equal(t, 59, result.Minute())
	})

	t.Run("endofmonth february non-leap year", func(t *testing.T) {
		dt := time.Date(2023, time.February, 15, 0, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(dt, "endofmonth", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 28, result.Day())
	})

	t.Run("endofweek from saturday", func(t *testing.T) {
		saturday := time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC)
		result, err := ApplyOperation(saturday, "endofweek", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 3, 17, 23, 59, 59, maxNanos, time.UTC), result)
	})
}

func TestApplyOperation_RoundTrunc(t *testing.T) {
	t.Run("round", func(t *testing.T) {
		tests := []struct {
			desc     string
			base     time.Time
			value    string
			expected time.Time
		}{
			{"day rounds up at noon", time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), "day",
				time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
			{"day rounds down before noon", time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), "day",
				time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{"hour rounds up at 30min", time.Date(2024, 1, 15, 14, 35, 0, 0, time.UTC), "hour",
				time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)},
			{"hour rounds down before 30min", time.Date(2024, 1, 15, 14, 15, 0, 0, time.UTC), "hour",
				time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)},
			{"minute rounds up at 30sec", time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC), "minute",
				time.Date(2024, 1, 15, 14, 31, 0, 0, time.UTC)},
			{"minute rounds down before 30sec", time.Date(2024, 1, 15, 14, 30, 10, 0, time.UTC), "minute",
				time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)},
			{"empty defaults to day", time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), "",
				time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		}
		for _, tc := range tests {
			t.Run(tc.desc, func(t *testing.T) {
				result, err := ApplyOperation(tc.base, "round", tc.value, time.UTC)
				require.NoError(t, err)
				assert.True(t, tc.expected.Equal(result), "expected=%v got=%v", tc.expected, result)
			})
		}
	})

	t.Run("trunc", func(t *testing.T) {
		base := time.Date(2024, 1, 15, 14, 35, 45, 123, time.UTC)
		tests := []struct {
			value    string
			expected time.Time
		}{
			{"day", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{"hour", time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)},
			{"minute", time.Date(2024, 1, 15, 14, 35, 0, 0, time.UTC)},
			{"", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}
		for _, tc := range tests {
			t.Run("trunc_"+tc.value, func(t *testing.T) {
				result, err := ApplyOperation(base, "trunc", tc.value, time.UTC)
				require.NoError(t, err)
				assert.True(t, tc.expected.Equal(result), "expected=%v got=%v", tc.expected, result)
			})
		}
	})

	t.Run("round invalid unit", func(t *testing.T) {
		_, err := ApplyOperation(time.Now(), "round", "invalid", time.UTC)
		assert.ErrorIs(t, err, ErrInvalidUnit)
	})

	t.Run("trunc invalid unit", func(t *testing.T) {
		_, err := ApplyOperation(time.Now(), "trunc", "invalid", time.UTC)
		assert.ErrorIs(t, err, ErrInvalidUnit)
	})
}

func TestApplyOperation_DayTime(t *testing.T) {
	base := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)

	t.Run("day set to 1", func(t *testing.T) {
		result, err := ApplyOperation(base, "day", "1", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Day())
		assert.Equal(t, time.March, result.Month())
	})

	t.Run("day set to 31", func(t *testing.T) {
		result, err := ApplyOperation(base, "day", "31", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 31, result.Day())
	})

	t.Run("day empty value unchanged", func(t *testing.T) {
		result, err := ApplyOperation(base, "day", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 15, result.Day())
	})

	t.Run("day invalid value", func(t *testing.T) {
		_, err := ApplyOperation(base, "day", "abc", time.UTC)
		assert.ErrorIs(t, err, ErrInvalidDay)
	})

	t.Run("time HH:MM:SS", func(t *testing.T) {
		result, err := ApplyOperation(base, "time", "08:30:00", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 8, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 15, result.Day()) // day preserved
	})

	t.Run("time HH:MM", func(t *testing.T) {
		result, err := ApplyOperation(base, "time", "22:00", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 22, result.Hour())
		assert.Equal(t, 0, result.Minute())
	})

	t.Run("time empty value unchanged", func(t *testing.T) {
		result, err := ApplyOperation(base, "time", "", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 14, result.Hour())
	})

	t.Run("time invalid value", func(t *testing.T) {
		_, err := ApplyOperation(base, "time", "invalid", time.UTC)
		assert.Error(t, err)
	})
}

func TestApplyOperation_UnknownOp(t *testing.T) {
	_, err := ApplyOperation(time.Now(), "foobar", "", time.UTC)
	assert.ErrorIs(t, err, ErrUnknownOperation)
}

func TestParseWeekday(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Weekday
	}{
		{"sunday", time.Sunday},
		{"monday", time.Monday},
		{"tuesday", time.Tuesday},
		{"wednesday", time.Wednesday},
		{"thursday", time.Thursday},
		{"friday", time.Friday},
		{"saturday", time.Saturday},
		{"MONDAY", time.Monday},   // case insensitive
		{"unknown", time.Sunday},  // default
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseWeekday(tc.name))
		})
	}
}

func TestNextWeekday(t *testing.T) {
	// Wednesday March 13, 2024
	wednesday := time.Date(2024, 3, 13, 10, 0, 0, 0, time.UTC)

	t.Run("next friday from wednesday", func(t *testing.T) {
		result := nextWeekday(wednesday, "friday", time.UTC)
		assert.Equal(t, time.Friday, result.Weekday())
		assert.Equal(t, 15, result.Day()) // March 15
		assert.Equal(t, 0, result.Hour()) // midnight
	})

	t.Run("next wednesday from wednesday is 7 days later", func(t *testing.T) {
		result := nextWeekday(wednesday, "wednesday", time.UTC)
		assert.Equal(t, time.Wednesday, result.Weekday())
		assert.Equal(t, 20, result.Day()) // March 20, not March 13
	})

	t.Run("next monday from wednesday", func(t *testing.T) {
		result := nextWeekday(wednesday, "monday", time.UTC)
		assert.Equal(t, time.Monday, result.Weekday())
		assert.Equal(t, 18, result.Day()) // March 18
	})
}

func TestEvaluateExpression_Pipelines(t *testing.T) {
	tests := []struct {
		input string
		desc  string
		check func(t *testing.T, result time.Time)
	}{
		{
			input: "2024-06-15 | endofquarter",
			desc:  "Q2 end is june 30",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, time.June, result.Month())
				assert.Equal(t, 30, result.Day())
				assert.Equal(t, 23, result.Hour())
			},
		},
		{
			input: "2024-02-15 | endofmonth",
			desc:  "feb leap year ends on 29",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 29, result.Day())
			},
		},
		{
			input: "2024-03-15 | trunc hour",
			desc:  "trunc hour zeroes minutes",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
			},
		},
		{
			input: "2024-07-15 | startofquarter",
			desc:  "Q3 start is july 1",
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, time.July, result.Month())
				assert.Equal(t, 1, result.Day())
			},
		},
		{
			input: "2024-03-20 | startofweek | +7d",
			desc:  "chained pipeline",
			check: func(t *testing.T, result time.Time) {
				// March 20 (Wed) → startofweek = March 18 (Mon) → +7d = March 25
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, time.March, result.Month())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := EvaluateExpression(tc.input, time.UTC)
			require.NoError(t, err)
			tc.check(t, result)
		})
	}
}