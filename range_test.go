package calcdate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInterval(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"", 0, false},
		{"1d", 24 * time.Hour, false},
		{"2w", 2 * 7 * 24 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"3h", 3 * time.Hour, false},
		{"10s", 10 * time.Second, false},
		{"abc", 0, true},
		{"1", 0, true},   // no unit
		{"d", 0, true},   // no number
		{"1M", 0, true},  // special interval, not parseable as duration
	}

	for _, tc := range tests {
		t.Run("interval_"+tc.input, func(t *testing.T) {
			dur, err := ParseInterval(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, dur)
		})
	}
}

func TestIterate_WithoutInterval(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	t.Run("simple range produces single result", func(t *testing.T) {
		iter := NewRangeIterator(start, end, 0, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, start.Equal(results[0].BeginTime))
		assert.True(t, end.Equal(results[0].EndTime))
		assert.Equal(t, 0, results[0].Index)
	})

	t.Run("same start and end", func(t *testing.T) {
		iter := NewRangeIterator(start, start, 0, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, start.Equal(results[0].BeginTime))
		assert.True(t, start.Equal(results[0].EndTime))
	})
}

func TestIterate_WithInterval(t *testing.T) {
	t.Run("7 days with 1d interval", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
		iter := NewRangeIterator(start, end, 24*time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		assert.Len(t, results, 7)
		// First chunk
		assert.Equal(t, 1, results[0].BeginTime.Day())
		assert.Equal(t, 2, results[0].EndTime.Day())
		// Last chunk
		assert.Equal(t, 7, results[6].BeginTime.Day())
		assert.Equal(t, 8, results[6].EndTime.Day())
	})

	t.Run("3 hours with 1h interval", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)
		iter := NewRangeIterator(start, end, time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("not evenly divisible capped at end", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC) // 7 days
		iter := NewRangeIterator(start, end, 3*24*time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		require.Len(t, results, 3) // [1-4], [4-7], [7-8(capped)]
		// Last chunk is shorter: begins on 7th, ends on 8th
		assert.Equal(t, 7, results[2].BeginTime.Day())
		assert.Equal(t, 8, results[2].EndTime.Day())
	})

	t.Run("zero duration range produces no results", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		iter := NewRangeIterator(start, start, time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("sub-second range skipped", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.Add(500 * time.Millisecond)
		iter := NewRangeIterator(start, end, time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("indexes are sequential", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
		iter := NewRangeIterator(start, end, 24*time.Hour, nil, time.UTC)
		results, err := iter.Iterate()
		require.NoError(t, err)
		for i, r := range results {
			assert.Equal(t, i, r.Index)
		}
	})
}

func TestIterateWithSpecialInterval(t *testing.T) {
	t.Run("12 months with 1M interval", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		results, err := IterateWithSpecialInterval(start, end, "1M", nil, time.UTC)
		require.NoError(t, err)
		assert.Len(t, results, 12)
		assert.Equal(t, time.January, results[0].BeginTime.Month())
		assert.Equal(t, time.December, results[11].BeginTime.Month())
	})

	t.Run("4 quarters with 1q interval", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		results, err := IterateWithSpecialInterval(start, end, "1q", nil, time.UTC)
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, time.January, results[0].BeginTime.Month())
		assert.Equal(t, time.April, results[1].BeginTime.Month())
		assert.Equal(t, time.July, results[2].BeginTime.Month())
		assert.Equal(t, time.October, results[3].BeginTime.Month())
	})

	t.Run("1Y interval over 6 months", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
		results, err := IterateWithSpecialInterval(start, end, "1Y", nil, time.UTC)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		// End capped at range end
		assert.True(t, end.Equal(results[0].EndTime))
	})

	t.Run("invalid unit", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := IterateWithSpecialInterval(start, end, "1x", nil, time.UTC)
		assert.Error(t, err)
	})

	t.Run("invalid format", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := IterateWithSpecialInterval(start, end, "x", nil, time.UTC)
		assert.Error(t, err)
	})
}

func TestParseInt(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		n, err := ParseInt("42")
		require.NoError(t, err)
		assert.Equal(t, 42, n)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseInt("abc")
		assert.ErrorIs(t, err, ErrInvalidNumberFormat)
	})

	t.Run("empty", func(t *testing.T) {
		_, err := ParseInt("")
		assert.Error(t, err)
	})
}
