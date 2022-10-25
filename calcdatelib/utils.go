package calcdatelib

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// calculateUnitOfTime returns defaultValue if partDate is empty
// If partDate have the sign - or plus, it will return defaultValue +/- partDate
func calculateUnitOfTime(partDate string, defaultValue int) (value int, err error) {
	r := regexp.MustCompile(`(\+|-){1}[0-9]+`)
	value = defaultValue

	if r.MatchString(partDate) {
		// number has + or -, so it's a relative time
		change, err := strconv.Atoi(partDate)
		if err != nil {
			return value, err
		}
		return defaultValue + change, nil
	} else {
		if len(partDate) != 0 {
			value, err = strconv.Atoi(partDate)
			return
		} else {
			return defaultValue, nil
		}
	}
}

// doubleReplace replaces keyword by a string representing a scanfformat (like %2d)
// in the string fmtstr. Finally the scanfformat is expandined to data
func doubleReplace(fmtstr string, keyword string, scanfformat string, data int) string {
	finalValue := fmt.Sprintf(scanfformat, data)
	res := fmtstr
	if strings.ContainsAny(fmtstr, keyword) {
		res = strings.ReplaceAll(fmtstr, keyword, finalValue)
		// res = fmt.Sprintf(res, data)
	}
	return res
}

func convertStdFormatToGolang(str string) string {
	res := strings.ReplaceAll(str, "%YYYY", "2006")
	res = strings.ReplaceAll(res, "%MM", "01")
	res = strings.ReplaceAll(res, "%DD", "02")
	res = strings.ReplaceAll(res, "%hh", "15")
	res = strings.ReplaceAll(res, "%mm", "04")
	res = strings.ReplaceAll(res, "%ss", "05")
	return res
}

// splitDate split a string which has format like // ::
func splitDate(ddate string) []string {
	var strsplit []string

	for _, v := range strings.Split(ddate, " ") {
		for _, w := range strings.Split(v, ":") {
			strsplit = append(strsplit, strings.Split(w, "/")...)
		}
	}
	return strsplit
}

// indexArray search a value in an array of string and returns its index
func indexArray(array []string, searchValue string) int {
	for index, value := range array {
		if value == searchValue {
			return index
		}
	}
	return -1
}

// DayInMonth returns the number of days in the month/year given in parameters
func DayInMonth(year int, month int) int {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
}

// renderTemplate will render tmpl according beginTime and endTime
// Possibility to use a function called MinusOneSecond
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
		return "", err
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, d)
	s := doc.String()
	if err != nil {
		return "", err
	}
	return s, err
}

func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

func StartOfDay(t time.Time, location *time.Location) time.Time {
	year, month, day := t.In(location).Date()
	dayStartTime := time.Date(year, month, day, 0, 0, 0, 0, location)
	return dayStartTime
}

func EndOfDay(t time.Time, location *time.Location) time.Time {
	year, month, day := t.In(location).Date()
	dayEndTime := time.Date(year, month, day, 23, 59, 59, 0, location)
	return dayEndTime
}

func IsSameDay(first time.Time, second time.Time) bool {
	return first.YearDay() == second.YearDay() && first.Year() == second.Year()
}

func DiffInDays(start time.Time, end time.Time) int {
	return int(end.Sub(start).Hours() / 24)
}

func IsLeapYear(year int) bool {
	return year%4 == 0 && year%100 != 0 || year%400 == 0
}

func RenderIntervalLines(beginDate Date, endDate Date, interval time.Duration, tmpl string) (res []string, err error) {
	tmpInterval := beginDate
	for tmpInterval.Before(&endDate) {
		tmpEndInterval := tmpInterval
		tmpEndInterval.Add(interval - time.Second)
		l, err := RenderTemplate(tmpl, tmpInterval.Time(), tmpEndInterval.Time())
		if err != nil {
			return res, err
		}
		results := strings.Split(l, "\n")
		res = append(res, results...)
		tmpInterval.Add(interval)
	}
	return
}
