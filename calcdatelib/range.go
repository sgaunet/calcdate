package calcdatelib

import (
	"fmt"
	"time"
)

// RangeIterator represents an iterator over a date range.
type RangeIterator struct {
	Start     time.Time
	End       time.Time
	Interval  time.Duration
	Transform *TransformNode
	Timezone  *time.Location
	Index     int
}

// IterationResult represents a single iteration result.
type IterationResult struct {
	BeginTime time.Time
	EndTime   time.Time
	Index     int
}

// NewRangeIterator creates a new range iterator
//nolint:lll // long function signature is readable
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

// Iterate generates all iterations for the range.
func (r *RangeIterator) Iterate() ([]IterationResult, error) {
	if r.Interval == 0 {
		return r.iterateWithoutInterval()
	}
	return r.iterateWithInterval()
}

func (r *RangeIterator) iterateWithoutInterval() ([]IterationResult, error) {
	beginTime, endTime, err := r.applyTransform(r.Start, r.End, 0)
	if err != nil {
		return nil, err
	}
	
	results := []IterationResult{{
		BeginTime: beginTime,
		EndTime:   endTime,
		Index:     0,
	}}
	return results, nil
}

func (r *RangeIterator) iterateWithInterval() ([]IterationResult, error) {
	results := []IterationResult{}
	currentBegin := r.Start
	index := 0
	
	for currentBegin.Before(r.End) {
		currentEnd := r.calculateIterationEnd(currentBegin)
		
		// Skip if this would create a zero-duration or very short range
		if r.isInvalidRange(currentBegin, currentEnd) {
			break
		}
		
		iterBegin, iterEnd, err := r.applyTransform(currentBegin, currentEnd, index)
		if err != nil {
			return nil, err
		}
		
		results = append(results, IterationResult{
			BeginTime: iterBegin,
			EndTime:   iterEnd,
			Index:     index,
		})
		
		currentBegin = currentEnd
		index++
		
		if index > MaxIterations {
			return nil, ErrTooManyIterations
		}
	}
	
	return results, nil
}

func (r *RangeIterator) calculateIterationEnd(currentBegin time.Time) time.Time {
	currentEnd := currentBegin.Add(r.Interval)
	if currentEnd.After(r.End) {
		currentEnd = r.End
	}
	return currentEnd
}

func (r *RangeIterator) isInvalidRange(begin, end time.Time) bool {
	return begin.Equal(end) || end.Sub(begin) < time.Second
}

func (r *RangeIterator) applyTransform(begin, end time.Time, index int) (time.Time, time.Time, error) {
	if r.Transform == nil {
		return begin, end, nil
	}
	
	ctx := &EvalContext{
		Now:      time.Now(),
		Timezone: r.Timezone,
	}
	return EvaluateTransform(r.Transform, begin, end, index, ctx)
}

// ParseInterval parses an interval string like "1d", "2h", "30m".
func ParseInterval(interval string) (time.Duration, error) {
	if interval == "" {
		return 0, nil
	}
	
	// Try to parse as Go duration first
	if dur, err := time.ParseDuration(interval); err == nil {
		return dur, nil
	}
	
	return parseCustomInterval(interval)
}

func parseCustomInterval(interval string) (time.Duration, error) {
	if len(interval) < MinIntervalLength {
		return 0, fmt.Errorf("%w: %s", ErrInvalidInterval, interval)
	}
	
	num, unit, err := parseIntervalComponents(interval)
	if err != nil {
		return 0, err
	}
	
	return convertUnitToDuration(num, unit)
}

func parseIntervalComponents(interval string) (int, string, error) {
	// Extract number and unit
	unitIdx := 0
	for unitIdx < len(interval) && (interval[unitIdx] >= '0' && interval[unitIdx] <= '9') {
		unitIdx++
	}
	
	if unitIdx == 0 || unitIdx >= len(interval) {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidInterval, interval)
	}
	
	num, err := ParseInt(interval[:unitIdx])
	if err != nil {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidNumberFormat, interval)
	}
	
	return num, interval[unitIdx:], nil
}

func convertUnitToDuration(num int, unit string) (time.Duration, error) {
	switch unit {
	case "s":
		return time.Duration(num) * time.Second, nil
	case "m":
		return time.Duration(num) * time.Minute, nil
	case "h":
		return time.Duration(num) * time.Hour, nil
	case "d":
		return time.Duration(num) * HoursInDay * time.Hour, nil
	case "w":
		return time.Duration(num) * DaysInWeek * HoursInDay * time.Hour, nil
	default:
		return 0, fmt.Errorf("%w: unit '%s' requires special handling", ErrInvalidInterval, unit)
	}
}

// ParseInt is a helper to parse integers.
func ParseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidNumberFormat, s)
	}
	return result, nil
}

// IterateWithSpecialInterval handles month and year intervals
//nolint:lll // long function signature is readable
func IterateWithSpecialInterval(start, end time.Time, interval string, transform *TransformNode, tz *time.Location) ([]IterationResult, error) {
	num, unit, err := parseSpecialInterval(interval)
	if err != nil {
		return nil, err
	}
	
	return performSpecialIteration(start, end, num, unit, transform, tz)
}

func parseSpecialInterval(interval string) (int, string, error) {
	if len(interval) < MinIntervalLength {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidInterval, interval)
	}
	
	return parseIntervalComponents(interval)
}

//nolint:lll // long function signature is readable
func performSpecialIteration(start, end time.Time, num int, unit string, transform *TransformNode, tz *time.Location) ([]IterationResult, error) {
	results := []IterationResult{}
	currentBegin := start
	index := 0
	
	for currentBegin.Before(end) || currentBegin.Equal(end) {
		currentEnd, err := calculateSpecialIntervalEnd(currentBegin, num, unit)
		if err != nil {
			return nil, err
		}
		
		// Don't exceed the overall end time
		if currentEnd.After(end) {
			currentEnd = end
		}
		
		// Skip invalid ranges
		if isInvalidTimeRange(currentBegin, currentEnd) {
			break
		}
		
		iterBegin, iterEnd, err := applySpecialTransform(currentBegin, currentEnd, index, transform, tz)
		if err != nil {
			return nil, err
		}
		
		results = append(results, IterationResult{
			BeginTime: iterBegin,
			EndTime:   iterEnd,
			Index:     index,
		})
		
		currentBegin = currentEnd
		index++
		
		if index > MaxIterations {
			return nil, ErrTooManyIterations
		}
	}
	
	return results, nil
}

func calculateSpecialIntervalEnd(begin time.Time, num int, unit string) (time.Time, error) {
	switch unit {
	case "M":
		return begin.AddDate(0, num, 0), nil
	case "Y":
		return begin.AddDate(num, 0, 0), nil
	case "q":
		return begin.AddDate(0, num*MonthsInQuarter, 0), nil
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidUnit, unit)
	}
}

func isInvalidTimeRange(begin, end time.Time) bool {
	return begin.Equal(end) || end.Sub(begin) < time.Second
}

//nolint:lll // long function signature is readable
func applySpecialTransform(begin, end time.Time, index int, transform *TransformNode, tz *time.Location) (time.Time, time.Time, error) {
	if transform == nil {
		return begin, end, nil
	}
	
	ctx := &EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}
	return EvaluateTransform(transform, begin, end, index, ctx)
}