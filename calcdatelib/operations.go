package calcdatelib

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDateValue parses a date value string into a time.Time.
func ParseDateValue(value string, ctx *EvalContext) (time.Time, error) {
	now := ctx.Now
	if now.IsZero() {
		now = time.Now()
	}
	
	loc := ctx.Timezone
	if loc == nil {
		loc = time.Local //nolint:gosmopolitan // intentional default to local timezone
	}
	
	// Adjust now to the specified timezone
	now = now.In(loc)
	
	// Try built-in keywords first
	if t, ok := parseBuiltinKeywords(value, now, loc); ok {
		return t, nil
	}
	
	// Check for relative dates like "+1d", "-2w"
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		return applyRelativeDate(now, value, loc)
	}
	
	// Try to parse as ISO date
	if t, err := parseISODate(value, loc); err == nil {
		return t, nil
	}
	
	// Try to parse as time (HH:MM:SS)
	if t, err := parseTimeOnly(value, now, loc); err == nil {
		return t, nil
	}
	
	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidDateValue, value)
}

func parseBuiltinKeywords(value string, now time.Time, loc *time.Location) (time.Time, bool) {
	value = strings.ToLower(value)
	
	switch value {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), true
	case "now":
		return now, true
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -1), true
	case "tomorrow":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1), true
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return nextWeekday(now, value, loc), true
	default:
		return time.Time{}, false
	}
}

func nextWeekday(from time.Time, weekdayName string, loc *time.Location) time.Time {
	targetWeekday := parseWeekday(weekdayName)
	
	// Start from tomorrow
	date := from.AddDate(0, 0, 1)
	
	// Find the next occurrence of the target weekday
	for date.Weekday() != targetWeekday {
		date = date.AddDate(0, 0, 1)
	}
	
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
}

func parseWeekday(name string) time.Weekday {
	switch strings.ToLower(name) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Sunday
	}
}

func applyRelativeDate(base time.Time, value string, _ *time.Location) (time.Time, error) {
	if len(value) < MinIntervalLength {
		return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidDateValue, value)
	}
	
	sign, startIdx := parseSign(value)
	num, unit, err := parseRelativeComponents(value, startIdx)
	if err != nil {
		return time.Time{}, err
	}
	
	return applyRelativeDelta(base, sign*num, unit)
}

func parseSign(value string) (int, int) {
	switch value[0] {
	case '+':
		return 1, 1
	case '-':
		return -1, 1
	default:
		return 1, 0
	}
}

func parseRelativeComponents(value string, startIdx int) (int, string, error) {
	// Find where the number ends and unit begins
	unitIdx := startIdx
	for unitIdx < len(value) && (value[unitIdx] >= '0' && value[unitIdx] <= '9') {
		unitIdx++
	}
	
	if unitIdx == startIdx || unitIdx >= len(value) {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidDateValue, value)
	}
	
	// Parse number
	num, err := strconv.Atoi(value[startIdx:unitIdx])
	if err != nil {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidNumberFormat, value)
	}
	
	return num, value[unitIdx:], nil
}

func applyRelativeDelta(base time.Time, num int, unit string) (time.Time, error) {
	switch unit {
	case "s":
		return base.Add(time.Duration(num) * time.Second), nil
	case "m":
		return base.Add(time.Duration(num) * time.Minute), nil
	case "h":
		return base.Add(time.Duration(num) * time.Hour), nil
	case "d":
		return base.AddDate(0, 0, num), nil
	case "w":
		return base.AddDate(0, 0, num*DaysInWeek), nil
	case "M":
		return base.AddDate(0, num, 0), nil
	case "Y":
		return base.AddDate(num, 0, 0), nil
	case "q":
		return base.AddDate(0, num*MonthsInQuarter, 0), nil
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrUnknownUnit, unit)
	}
}

func parseISODate(value string, loc *time.Location) (time.Time, error) {
	// Try various ISO date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, value, loc); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidISODateFormat, value)
}

func parseTimeOnly(value string, baseDate time.Time, loc *time.Location) (time.Time, error) {
	// Parse time formats HH:MM:SS or HH:MM
	formats := []string{
		"15:04:05",
		"15:04",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			// Combine with base date
			return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
				t.Hour(), t.Minute(), t.Second(), 0, loc), nil
		}
	}
	
	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidTime, value)
}

// ApplyOperation applies an operation to a date.
func ApplyOperation(date time.Time, op, value string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.Local //nolint:gosmopolitan // intentional default to local timezone
	}
	
	// Handle arithmetic operations
	if result, ok := applyArithmeticOperation(date, op, value, loc); ok {
		return result.t, result.err
	}
	
	// Handle boundary operations
	if result, ok := applyBoundaryOperation(date, op, loc); ok {
		return result.t, result.err
	}
	
	// Handle value-setting operations
	if result, ok := applyValueOperation(date, op, value, loc); ok {
		return result.t, result.err
	}
	
	// Handle transformation operations
	if result, ok := applyTransformOperation(date, op, value, loc); ok {
		return result.t, result.err
	}
	
	return time.Time{}, fmt.Errorf("%w: %s", ErrUnknownOperation, op)
}

type operationResult struct {
	t   time.Time
	err error
}

func applyArithmeticOperation(date time.Time, op, value string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "+":
		t, err := applyRelativeDate(date, "+"+value, loc)
		return operationResult{t, err}, true
	case "-":
		t, err := applyRelativeDate(date, "-"+value, loc)
		return operationResult{t, err}, true
	default:
		return operationResult{}, false
	}
}

func applyBoundaryOperation(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	if result, ok := applyDayBoundaryOps(date, op, loc); ok {
		return result, true
	}
	if result, ok := applyWeekBoundaryOps(date, op, loc); ok {
		return result, true
	}
	if result, ok := applyMonthBoundaryOps(date, op, loc); ok {
		return result, true
	}
	if result, ok := applyYearQuarterBoundaryOps(date, op, loc); ok {
		return result, true
	}
	if result, ok := applyTimeBoundaryOps(date, op, loc); ok {
		return result, true
	}
	return operationResult{}, false
}

func applyDayBoundaryOps(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "start", "startofday":
		return operationResult{startOfDay(date, loc), nil}, true
	case "end", "endofday":
		return operationResult{endOfDay(date, loc), nil}, true
	default:
		return operationResult{}, false
	}
}

func applyWeekBoundaryOps(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "startofweek":
		return operationResult{startOfWeek(date, loc), nil}, true
	case "endofweek":
		return operationResult{endOfWeek(date, loc), nil}, true
	default:
		return operationResult{}, false
	}
}

func applyMonthBoundaryOps(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "startofmonth":
		return operationResult{startOfMonth(date, loc), nil}, true
	case "endofmonth":
		return operationResult{endOfMonth(date, loc), nil}, true
	default:
		return operationResult{}, false
	}
}

func applyYearQuarterBoundaryOps(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "startofyear":
		return operationResult{startOfYear(date, loc), nil}, true
	case "endofyear":
		return operationResult{endOfYear(date, loc), nil}, true
	case "startofquarter":
		return operationResult{startOfQuarter(date, loc), nil}, true
	case "endofquarter":
		return operationResult{endOfQuarter(date, loc), nil}, true
	default:
		return operationResult{}, false
	}
}

func applyTimeBoundaryOps(date time.Time, op string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "startofhour":
		return operationResult{startOfHour(date, loc), nil}, true
	case "endofhour":
		return operationResult{endOfHour(date, loc), nil}, true
	case "startofminute":
		return operationResult{startOfMinute(date, loc), nil}, true
	case "endofminute":
		return operationResult{endOfMinute(date, loc), nil}, true
	case "startofsecond":
		return operationResult{startOfSecond(date, loc), nil}, true
	case "endofsecond":
		return operationResult{endOfSecond(date, loc), nil}, true
	default:
		return operationResult{}, false
	}
}

func applyValueOperation(date time.Time, op, value string, loc *time.Location) (operationResult, bool) {
	const (
		dayOp  = "day"
		timeOp = "time"
	)
	switch op {
	case dayOp:
		if value == "" {
			return operationResult{date, nil}, true
		}
		day, err := strconv.Atoi(value)
		if err != nil {
			return operationResult{time.Time{}, fmt.Errorf("%w: %s", ErrInvalidDay, value)}, true
		}
		t := time.Date(date.Year(), date.Month(), day, date.Hour(), date.Minute(), date.Second(), 0, loc)
		return operationResult{t, nil}, true
	case timeOp:
		if value == "" {
			return operationResult{date, nil}, true
		}
		t, err := parseTimeOnly(value, date, loc)
		return operationResult{t, err}, true
	default:
		return operationResult{}, false
	}
}

func applyTransformOperation(date time.Time, op, value string, loc *time.Location) (operationResult, bool) {
	switch op {
	case "round":
		t, err := roundDate(date, value, loc)
		return operationResult{t, err}, true
	case "trunc":
		t, err := truncateDate(date, value, loc)
		return operationResult{t, err}, true
	default:
		return operationResult{}, false
	}
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func endOfDay(t time.Time, loc *time.Location) time.Time {
	const maxNanos = 999999999
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, maxNanos, loc)
}

func startOfWeek(t time.Time, loc *time.Location) time.Time {
	// Adjust to Monday
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = DaysInWeek
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return startOfDay(monday, loc)
}

func endOfWeek(t time.Time, loc *time.Location) time.Time {
	// Adjust to Sunday
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = DaysInWeek
	}
	sunday := t.AddDate(0, 0, DaysInWeek-weekday)
	return endOfDay(sunday, loc)
}

func startOfMonth(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
}

func endOfMonth(t time.Time, loc *time.Location) time.Time {
	// Go to first day of next month, then subtract one day
	firstOfNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc)
	lastOfMonth := firstOfNext.AddDate(0, 0, -1)
	return endOfDay(lastOfMonth, loc)
}

func startOfYear(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, loc)
}

func endOfYear(t time.Time, loc *time.Location) time.Time {
	const maxNanos = 999999999
	const lastDay = 31
	return time.Date(t.Year(), time.December, lastDay, 23, 59, 59, maxNanos, loc)
}

func startOfQuarter(t time.Time, loc *time.Location) time.Time {
	month := t.Month()
	var quarterStartMonth time.Month
	
	switch {
	case month <= FirstQuarterEnd:
		quarterStartMonth = time.January
	case month <= SecondQuarterEnd:
		quarterStartMonth = time.April
	case month <= ThirdQuarterEnd:
		quarterStartMonth = time.July
	default:
		quarterStartMonth = time.October
	}
	
	return time.Date(t.Year(), quarterStartMonth, 1, 0, 0, 0, 0, loc)
}

func endOfQuarter(t time.Time, loc *time.Location) time.Time {
	month := t.Month()
	var quarterEndMonth time.Month
	
	switch {
	case month <= FirstQuarterEnd:
		quarterEndMonth = time.March
	case month <= SecondQuarterEnd:
		quarterEndMonth = time.June
	case month <= ThirdQuarterEnd:
		quarterEndMonth = time.September
	default:
		quarterEndMonth = time.December
	}
	
	// Get last day of quarter end month
	firstOfNext := time.Date(t.Year(), quarterEndMonth+1, 1, 0, 0, 0, 0, loc)
	lastOfQuarter := firstOfNext.AddDate(0, 0, -1)
	return endOfDay(lastOfQuarter, loc)
}

func roundDate(t time.Time, unit string, loc *time.Location) (time.Time, error) {
	switch unit {
	case "day", "":
		// Round to nearest day
		hour := t.Hour()
		if hour >= NoonHour {
			return startOfDay(t.AddDate(0, 0, 1), loc), nil
		}
		return startOfDay(t, loc), nil
		
	case "hour":
		// Round to nearest hour
		minute := t.Minute()
		if minute >= HalfMinute {
			return t.Add(time.Hour).Truncate(time.Hour), nil
		}
		return t.Truncate(time.Hour), nil
		
	case "minute":
		// Round to nearest minute
		second := t.Second()
		if second >= HalfSecond {
			return t.Add(time.Minute).Truncate(time.Minute), nil
		}
		return t.Truncate(time.Minute), nil
		
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidUnit, unit)
	}
}

func truncateDate(t time.Time, unit string, loc *time.Location) (time.Time, error) {
	switch unit {
	case "day", "":
		return startOfDay(t, loc), nil
	case "hour":
		return t.Truncate(time.Hour), nil
	case "minute":
		return t.Truncate(time.Minute), nil
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidUnit, unit)
	}
}

func startOfHour(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
}

func endOfHour(t time.Time, loc *time.Location) time.Time {
	const maxNanos = 999999999
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, maxNanos, loc)
}

func startOfMinute(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, loc)
}

func endOfMinute(t time.Time, loc *time.Location) time.Time {
	const maxNanos = 999999999
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, maxNanos, loc)
}

func startOfSecond(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
}

func endOfSecond(t time.Time, loc *time.Location) time.Time {
	const maxNanos = 999999999
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), maxNanos, loc)
}