package calcdatelib

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDateValue parses a date value string into a time.Time
func ParseDateValue(value string, ctx *EvalContext) (time.Time, error) {
	now := ctx.Now
	if now.IsZero() {
		now = time.Now()
	}
	
	loc := ctx.Timezone
	if loc == nil {
		loc = time.Local
	}
	
	// Adjust now to the specified timezone
	now = now.In(loc)
	
	switch strings.ToLower(value) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	case "now":
		return now, nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -1), nil
	case "tomorrow":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1), nil
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return nextWeekday(now, value, loc), nil
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
	
	return time.Time{}, fmt.Errorf("cannot parse date value: %s", value)
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

func applyRelativeDate(base time.Time, value string, loc *time.Location) (time.Time, error) {
	// Parse relative date like "+1d", "-2w", "3M"
	if len(value) < 2 {
		return time.Time{}, fmt.Errorf("invalid relative date: %s", value)
	}
	
	// Extract sign
	sign := 1
	startIdx := 0
	if value[0] == '+' {
		startIdx = 1
	} else if value[0] == '-' {
		sign = -1
		startIdx = 1
	}
	
	// Find where the number ends and unit begins
	unitIdx := startIdx
	for unitIdx < len(value) && (value[unitIdx] >= '0' && value[unitIdx] <= '9') {
		unitIdx++
	}
	
	if unitIdx == startIdx || unitIdx >= len(value) {
		return time.Time{}, fmt.Errorf("invalid relative date format: %s", value)
	}
	
	// Parse number
	num, err := strconv.Atoi(value[startIdx:unitIdx])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in relative date: %s", value)
	}
	num *= sign
	
	// Get unit
	unit := value[unitIdx:]
	
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
		return base.AddDate(0, 0, num*7), nil
	case "M":
		return base.AddDate(0, num, 0), nil
	case "Y":
		return base.AddDate(num, 0, 0), nil
	case "q":
		return base.AddDate(0, num*3, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
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
	
	return time.Time{}, fmt.Errorf("not an ISO date: %s", value)
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
	
	return time.Time{}, fmt.Errorf("not a valid time: %s", value)
}

// ApplyOperation applies an operation to a date
func ApplyOperation(date time.Time, op, value string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	
	switch op {
	case "+":
		return applyRelativeDate(date, "+"+value, loc)
	case "-":
		return applyRelativeDate(date, "-"+value, loc)
		
	case "start", "startofday":
		return startOfDay(date, loc), nil
	case "end", "endofday":
		return endOfDay(date, loc), nil
		
	case "startofweek":
		return startOfWeek(date, loc), nil
	case "endofweek":
		return endOfWeek(date, loc), nil
		
	case "startofmonth":
		return startOfMonth(date, loc), nil
	case "endofmonth":
		return endOfMonth(date, loc), nil
		
	case "startofyear":
		return startOfYear(date, loc), nil
	case "endofyear":
		return endOfYear(date, loc), nil
		
	case "startofquarter":
		return startOfQuarter(date, loc), nil
	case "endofquarter":
		return endOfQuarter(date, loc), nil
		
	case "day":
		// Set to specific day of month
		if value == "" {
			return date, nil
		}
		day, err := strconv.Atoi(value)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day number: %s", value)
		}
		return time.Date(date.Year(), date.Month(), day, date.Hour(), date.Minute(), date.Second(), 0, loc), nil
		
	case "time":
		// Set to specific time
		if value == "" {
			return date, nil
		}
		t, err := parseTimeOnly(value, date, loc)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
		
	case "round":
		return roundDate(date, value, loc)
		
	case "trunc":
		return truncateDate(date, value, loc)
		
	default:
		return time.Time{}, fmt.Errorf("unknown operation: %s", op)
	}
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func endOfDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, loc)
}

func startOfWeek(t time.Time, loc *time.Location) time.Time {
	// Adjust to Monday
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return startOfDay(monday, loc)
}

func endOfWeek(t time.Time, loc *time.Location) time.Time {
	// Adjust to Sunday
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	sunday := t.AddDate(0, 0, 7-weekday)
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
	return time.Date(t.Year(), time.December, 31, 23, 59, 59, 999999999, loc)
}

func startOfQuarter(t time.Time, loc *time.Location) time.Time {
	month := t.Month()
	var quarterStartMonth time.Month
	
	switch {
	case month <= 3:
		quarterStartMonth = time.January
	case month <= 6:
		quarterStartMonth = time.April
	case month <= 9:
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
	case month <= 3:
		quarterEndMonth = time.March
	case month <= 6:
		quarterEndMonth = time.June
	case month <= 9:
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
		if hour >= 12 {
			return startOfDay(t.AddDate(0, 0, 1), loc), nil
		}
		return startOfDay(t, loc), nil
		
	case "hour":
		// Round to nearest hour
		minute := t.Minute()
		if minute >= 30 {
			return t.Add(time.Hour).Truncate(time.Hour), nil
		}
		return t.Truncate(time.Hour), nil
		
	case "minute":
		// Round to nearest minute
		second := t.Second()
		if second >= 30 {
			return t.Add(time.Minute).Truncate(time.Minute), nil
		}
		return t.Truncate(time.Minute), nil
		
	default:
		return time.Time{}, fmt.Errorf("unsupported round unit: %s", unit)
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
		return time.Time{}, fmt.Errorf("unsupported truncate unit: %s", unit)
	}
}