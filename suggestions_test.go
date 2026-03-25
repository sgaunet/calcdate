package calcdate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"startofmonth", "startofmoonth", 1},
		{"endofweek", "endofwek", 1},
		{"today", "tooday", 1},
		{"tomorrow", "tomorow", 1},
	}

	for _, tc := range tests {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			assert.Equal(t, tc.expected, levenshteinDistance(tc.a, tc.b))
		})
	}
}

func TestFindSuggestions(t *testing.T) {
	t.Run("finds close matches", func(t *testing.T) {
		candidates := []string{"startofmonth", "endofmonth", "startofweek"}
		result := findSuggestions("startofmoonth", candidates)
		require.Len(t, result, 1)
		assert.Equal(t, "startofmonth", result[0])
	})

	t.Run("respects threshold", func(t *testing.T) {
		candidates := []string{"startofmonth", "endofmonth"}
		result := findSuggestions("xyz", candidates)
		assert.Empty(t, result)
	})

	t.Run("returns at most 3 suggestions", func(t *testing.T) {
		candidates := validOperationNames()
		result := findSuggestions("startof", candidates)
		assert.LessOrEqual(t, len(result), maxSuggestions)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		candidates := []string{"startofmonth", "endofmonth"}
		result := findSuggestions("completelyunrelated", candidates)
		assert.Empty(t, result)
	})

	t.Run("case insensitive", func(t *testing.T) {
		candidates := []string{"startOfMonth", "endOfMonth"}
		result := findSuggestions("StartOfMoonth", candidates)
		require.Len(t, result, 1)
		assert.Equal(t, "startOfMonth", result[0])
	})
}

func TestExpressionErrorUnwrap(t *testing.T) {
	err := &ExpressionError{
		Wrapped: ErrUnknownOperation,
		Message: "unknown operation: foo",
	}

	assert.True(t, errors.Is(err, ErrUnknownOperation))
	assert.False(t, errors.Is(err, ErrUnknownUnit))
}

func TestExpressionErrorMessage(t *testing.T) {
	t.Run("with suggestions", func(t *testing.T) {
		err := &ExpressionError{
			Wrapped:     ErrUnknownOperation,
			Message:     "unknown operation: startofmoonth",
			Suggestions: []string{"startofmonth"},
			Position:    -1,
		}

		msg := err.Error()
		assert.Contains(t, msg, "Did you mean: startofmonth?")
	})

	t.Run("with valid options", func(t *testing.T) {
		err := &ExpressionError{
			Wrapped:      ErrUnknownUnit,
			Message:      "unknown unit: x",
			ValidOptions: []string{"s (seconds)", "m (minutes)"},
			Position:     -1,
		}

		msg := err.Error()
		assert.Contains(t, msg, "Valid options: s (seconds), m (minutes)")
	})

	t.Run("with position highlight", func(t *testing.T) {
		err := &ExpressionError{
			Wrapped:    ErrUnknownOperation,
			Message:    "unknown operation: startofmoonth",
			Position:   8,
			Expression: "today | startofmoonth",
		}

		msg := err.Error()
		assert.Contains(t, msg, "today | startofmoonth")
		assert.Contains(t, msg, "^")
	})

	t.Run("no extras", func(t *testing.T) {
		err := &ExpressionError{
			Wrapped:  ErrUnknownOperation,
			Message:  "unknown operation: foo",
			Position: -1,
		}

		assert.Equal(t, "unknown operation: foo", err.Error())
	})
}

func TestFormatPositionHighlight(t *testing.T) {
	t.Run("highlights correct position", func(t *testing.T) {
		result := formatPositionHighlight("today | startofmoonth", 8)
		assert.Contains(t, result, "today | startofmoonth")
		assert.Contains(t, result, "        ^")
	})

	t.Run("position at start", func(t *testing.T) {
		result := formatPositionHighlight("foo", 0)
		assert.Contains(t, result, "foo")
		assert.Contains(t, result, "^")
	})

	t.Run("negative position returns empty", func(t *testing.T) {
		result := formatPositionHighlight("foo", -1)
		assert.Empty(t, result)
	})

	t.Run("position beyond string returns empty", func(t *testing.T) {
		result := formatPositionHighlight("foo", 10)
		assert.Empty(t, result)
	})
}

func TestNewUnknownOperationError(t *testing.T) {
	err := newUnknownOperationError("startofmoonth", -1, "")

	assert.True(t, errors.Is(err, ErrUnknownOperation))
	assert.Contains(t, err.Error(), "startofmoonth")
	assert.Contains(t, err.Error(), "Did you mean: startofmonth?")
}

func TestNewUnknownUnitError(t *testing.T) {
	err := newUnknownUnitError("x")

	assert.True(t, errors.Is(err, ErrUnknownUnit))
	assert.Contains(t, err.Error(), "unknown unit: x")
	assert.Contains(t, err.Error(), "Valid options:")
	assert.Contains(t, err.Error(), "d (days)")
}

func TestNewInvalidDateValueError(t *testing.T) {
	err := newInvalidDateValueError("tomorow")

	assert.True(t, errors.Is(err, ErrInvalidDateValue))
	assert.Contains(t, err.Error(), "tomorow")
	assert.Contains(t, err.Error(), "Did you mean: tomorrow?")
}
