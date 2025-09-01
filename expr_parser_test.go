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