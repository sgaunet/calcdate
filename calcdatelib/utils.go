package calcdatelib

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// ErrFormatMismatch is returned when the date format doesn't match the expected format.
var ErrFormatMismatch = errors.New("format is not the same")

// doubleReplace replaces keyword by a string representing a scanfformat (like %2d)
// in the string fmtstr. Finally the scanfformat is expanded to data.
func doubleReplace(fmtstr string, keyword string, scanfformat string, data int) string {
	finalValue := fmt.Sprintf(scanfformat, data)
	res := fmtstr
	if strings.ContainsAny(fmtstr, keyword) {
		res = strings.ReplaceAll(fmtstr, keyword, finalValue)
		// res = fmt.Sprintf(res, data)
	}
	return res
}

// convertStdFormatToGolang converts standard date format to Go time format.
func convertStdFormatToGolang(str string) string {
	res := strings.ReplaceAll(str, "%YYYY", "2006")
	res = strings.ReplaceAll(res, "%MM", "01")
	res = strings.ReplaceAll(res, "%DD", "02")
	res = strings.ReplaceAll(res, "%hh", "15")
	res = strings.ReplaceAll(res, "%mm", "04")
	res = strings.ReplaceAll(res, "%ss", "05")
	return res
}

// ConvertUnixFormatToGolang converts Unix date format specifiers to Go time format.
func ConvertUnixFormatToGolang(str string) string {
	res := str
	
	// Handle literal percent first (before any other % replacements)
	res = strings.ReplaceAll(res, "%%", "\x00PERCENT\x00") // Use temporary placeholder
	
	// Core date/time formats
	res = strings.ReplaceAll(res, "%Y", "2006")  // 4-digit year
	res = strings.ReplaceAll(res, "%y", "06")    // 2-digit year
	res = strings.ReplaceAll(res, "%m", "01")    // month (01-12)
	res = strings.ReplaceAll(res, "%d", "02")    // day (01-31)
	res = strings.ReplaceAll(res, "%H", "15")    // hour 24-format (00-23)
	res = strings.ReplaceAll(res, "%I", "03")    // hour 12-format (01-12)
	res = strings.ReplaceAll(res, "%M", "04")    // minute (00-59)
	res = strings.ReplaceAll(res, "%S", "05")    // second (00-59)
	res = strings.ReplaceAll(res, "%p", "PM")    // AM/PM
	
	// Weekday formats
	res = strings.ReplaceAll(res, "%a", "Mon")      // short weekday name
	res = strings.ReplaceAll(res, "%A", "Monday")   // full weekday name
	
	// Month formats
	res = strings.ReplaceAll(res, "%b", "Jan")      // short month name
	res = strings.ReplaceAll(res, "%B", "January")  // full month name
	
	// Timezone formats
	res = strings.ReplaceAll(res, "%z", "-0700")    // numeric timezone offset
	res = strings.ReplaceAll(res, "%Z", "MST")      // timezone name
	
	// Additional common formats
	res = strings.ReplaceAll(res, "%j", "002")      // day of year (001-366)
	// Note: %w and %u (weekday numbers) don't have direct Go equivalents
	
	// Restore literal percent at the end
	res = strings.ReplaceAll(res, "\x00PERCENT\x00", "%")
	
	return res
}

// createRegexpFromIfmt creates a regular expression pattern from input format.
func createRegexpFromIfmt(ifmt string) string {
	r := strings.ReplaceAll(ifmt, "%YYYY", "(?P<Year>([\\+-]?\\d+)?)")
	r = strings.ReplaceAll(r, "%MM", `(?P<Month>([\+-]?\d+)?)`)
	r = strings.ReplaceAll(r, "%DD", `(?P<Day>([\+-]?\d+)?)`)
	r = strings.ReplaceAll(r, "%hh", `(?P<Hour>([\+-]?\d+)?)`)
	r = strings.ReplaceAll(r, "%mm", `(?P<Minute>([\+-]?\d+)?)`)
	r = strings.ReplaceAll(r, "%ss", `(?P<Second>([\+-]?\d+)?)`)
	r = fmt.Sprintf("^%s$", r)
	return r
}

// findIndexOf searches a value in an array of string and returns its index.
func (d *Date) findIndexOf(searchGroup string) (int, error) {
	var idxFound int

	for index, value := range d.subExpNames {
		if value == searchGroup {
			idxFound = index
		}
	}
	if len(d.submatch) < idxFound {
		return -1, ErrFormatMismatch
	}
	if len(d.submatch) != 0 {
		if len(d.submatch[idxFound]) != 0 {
			return idxFound, nil
		}
	}
	return -1, nil
}

// DayInMonth returns the number of days in the month/year given in parameters
// DayInMonth returns the number of days in the specified month and year.
func DayInMonth(year int, month int) int {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
}

// RenderTemplate renders tmpl according beginTime and endTime.
// Possibility to use a function called MinusOneSecond.
func RenderTemplate(tmpl string, beginTime time.Time, endTime time.Time) (string, error) {
	type dataDate struct {
		BeginTime time.Time
		EndTime   time.Time
	}

	d := dataDate{
		BeginTime: beginTime,
		EndTime:   endTime,
	}
	temapltestr := convertStdFormatToGolang(tmpl)

	t, err := template.New("calcline").Funcs(template.FuncMap{
		"MinusOneSecond": func(t time.Time) time.Time {
			return t.Add(-1 * time.Second)
		},
	}).Parse(temapltestr)
	if err != nil {
		return "", fmt.Errorf("template.Parse: %w", err)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, d)
	if err != nil {
		return "", fmt.Errorf("template.Execute: %w", err)
	}
	return doc.String(), nil
}

// AddDays adds the specified number of days to a time.Time.
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// StartOfDay returns the start of the day (00:00:00) for the given time and location.
func StartOfDay(t time.Time, location *time.Location) time.Time {
	year, month, day := t.In(location).Date()
	dayStartTime := time.Date(year, month, day, 0, 0, 0, 0, location)
	return dayStartTime
}

// EndOfDay returns the end of the day (23:59:59) for the given time and location.
func EndOfDay(t time.Time, location *time.Location) time.Time {
	year, month, day := t.In(location).Date()
	dayEndTime := time.Date(year, month, day, 23, 59, 59, 0, location)
	return dayEndTime
}

// IsSameDay returns true if both times are on the same day.
func IsSameDay(first time.Time, second time.Time) bool {
	return first.YearDay() == second.YearDay() && first.Year() == second.Year()
}

// DiffInDays returns the difference in days between two times.
func DiffInDays(start time.Time, end time.Time) int {
	return int(end.Sub(start).Hours() / HoursPerDay)
}

// IsLeapYear returns true if the given year is a leap year.
func IsLeapYear(year int) bool {
	return year%4 == 0 && year%100 != 0 || year%400 == 0
}

// RenderIntervalLines renders multiple interval lines between two dates using a template.
func RenderIntervalLines(beginDate Date, endDate Date, interval time.Duration, tmpl string) ([]string, error) {
	var res []string
	tmpInterval := beginDate
	for tmpInterval.Before(&endDate) {
		tmpEndInterval := tmpInterval
		tmpEndInterval.Add(interval)
		l, err := RenderTemplate(tmpl, tmpInterval.Time(), tmpEndInterval.Time())
		if err != nil {
			return res, err
		}
		results := strings.Split(l, "\n")
		res = append(res, results...)

		tmpInterval.Add(interval)
	}
	return res, nil
}
