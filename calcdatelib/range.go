package calcdatelib

import (
	"fmt"
	"time"
)

// RangeIterator represents an iterator over a date range
type RangeIterator struct {
	Start     time.Time
	End       time.Time
	Interval  time.Duration
	Transform *TransformNode
	Timezone  *time.Location
	Index     int
}

// IterationResult represents a single iteration result
type IterationResult struct {
	BeginTime time.Time
	EndTime   time.Time
	Index     int
}

// NewRangeIterator creates a new range iterator
func NewRangeIterator(start, end time.Time, interval time.Duration, transform *TransformNode, tz *time.Location) *RangeIterator {
	return &RangeIterator{
		Start:     start,
		End:       end,
		Interval:  interval,
		Transform: transform,
		Timezone:  tz,
		Index:     0,
	}
}

// Iterate generates all iterations for the range
func (r *RangeIterator) Iterate() ([]IterationResult, error) {
	results := []IterationResult{}
	
	if r.Interval == 0 {
		// No iteration, just return the range
		beginTime := r.Start
		endTime := r.End
		
		// Apply transform if provided
		if r.Transform != nil {
			ctx := &EvalContext{
				Now:      time.Now(),
				Timezone: r.Timezone,
			}
			var err error
			beginTime, endTime, err = EvaluateTransform(r.Transform, beginTime, endTime, 0, ctx)
			if err != nil {
				return nil, err
			}
		}
		
		results = append(results, IterationResult{
			BeginTime: beginTime,
			EndTime:   endTime,
			Index:     0,
		})
		return results, nil
	}
	
	// Iterate with interval
	currentBegin := r.Start
	index := 0
	
	for currentBegin.Before(r.End) {
		// Calculate end of this iteration
		currentEnd := currentBegin.Add(r.Interval)
		
		// Don't exceed the overall end time
		if currentEnd.After(r.End) {
			currentEnd = r.End
		}
		
		// Skip if this would create a zero-duration or very short range (less than 1 second)
		if currentBegin.Equal(currentEnd) || currentEnd.Sub(currentBegin) < time.Second {
			break
		}
		
		// Apply transform if provided
		iterBegin := currentBegin
		iterEnd := currentEnd
		
		if r.Transform != nil {
			ctx := &EvalContext{
				Now:      time.Now(),
				Timezone: r.Timezone,
			}
			var err error
			iterBegin, iterEnd, err = EvaluateTransform(r.Transform, iterBegin, iterEnd, index, ctx)
			if err != nil {
				return nil, err
			}
		}
		
		results = append(results, IterationResult{
			BeginTime: iterBegin,
			EndTime:   iterEnd,
			Index:     index,
		})
		
		// Move to next iteration
		currentBegin = currentEnd
		index++
		
		// Prevent infinite loop
		if index > 10000 {
			return nil, fmt.Errorf("too many iterations (>10000)")
		}
	}
	
	return results, nil
}

// ParseInterval parses an interval string like "1d", "2h", "30m"
func ParseInterval(interval string) (time.Duration, error) {
	if interval == "" {
		return 0, nil
	}
	
	// Try to parse as Go duration first (handles h, m, s)
	if dur, err := time.ParseDuration(interval); err == nil {
		return dur, nil
	}
	
	// Handle custom units (d, w, M, Y)
	if len(interval) < 2 {
		return 0, fmt.Errorf("invalid interval: %s", interval)
	}
	
	// Extract number and unit
	unitIdx := 0
	for unitIdx < len(interval) && (interval[unitIdx] >= '0' && interval[unitIdx] <= '9') {
		unitIdx++
	}
	
	if unitIdx == 0 || unitIdx >= len(interval) {
		return 0, fmt.Errorf("invalid interval format: %s", interval)
	}
	
	num, err := ParseInt(interval[:unitIdx])
	if err != nil {
		return 0, fmt.Errorf("invalid number in interval: %s", interval)
	}
	
	unit := interval[unitIdx:]
	
	switch unit {
	case "s":
		return time.Duration(num) * time.Second, nil
	case "m":
		return time.Duration(num) * time.Minute, nil
	case "h":
		return time.Duration(num) * time.Hour, nil
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	case "w":
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	default:
		// For M (month) and Y (year), we can't return a fixed duration
		// These need special handling in the iterator
		return 0, fmt.Errorf("interval unit '%s' requires special handling", unit)
	}
}

// ParseInt is a helper to parse integers
func ParseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// IterateWithSpecialInterval handles month and year intervals
func IterateWithSpecialInterval(start, end time.Time, interval string, transform *TransformNode, tz *time.Location) ([]IterationResult, error) {
	results := []IterationResult{}
	
	// Parse interval
	if len(interval) < 2 {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}
	
	unitIdx := 0
	for unitIdx < len(interval) && (interval[unitIdx] >= '0' && interval[unitIdx] <= '9') {
		unitIdx++
	}
	
	num, err := ParseInt(interval[:unitIdx])
	if err != nil {
		return nil, err
	}
	
	unit := interval[unitIdx:]
	
	currentBegin := start
	index := 0
	
	for currentBegin.Before(end) || currentBegin.Equal(end) {
		var currentEnd time.Time
		
		switch unit {
		case "M":
			currentEnd = currentBegin.AddDate(0, num, 0)
		case "Y":
			currentEnd = currentBegin.AddDate(num, 0, 0)
		case "q":
			currentEnd = currentBegin.AddDate(0, num*3, 0)
		default:
			return nil, fmt.Errorf("unsupported interval unit: %s", unit)
		}
		
		// Don't exceed the overall end time
		if currentEnd.After(end) {
			currentEnd = end
		}
		
		// Skip if this would create a zero-duration or very short range (less than 1 second)
		if currentBegin.Equal(currentEnd) || currentEnd.Sub(currentBegin) < time.Second {
			break
		}
		
		// Apply transform if provided
		iterBegin := currentBegin
		iterEnd := currentEnd
		
		if transform != nil {
			ctx := &EvalContext{
				Now:      time.Now(),
				Timezone: tz,
			}
			iterBegin, iterEnd, err = EvaluateTransform(transform, iterBegin, iterEnd, index, ctx)
			if err != nil {
				return nil, err
			}
		}
		
		results = append(results, IterationResult{
			BeginTime: iterBegin,
			EndTime:   iterEnd,
			Index:     index,
		})
		
		// Move to next iteration
		currentBegin = currentEnd
		index++
		
		// Prevent infinite loop
		if index > 10000 {
			return nil, fmt.Errorf("too many iterations (>10000)")
		}
	}
	
	return results, nil
}