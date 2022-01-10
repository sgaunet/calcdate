package calcdate

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ddate struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

// SplitDate split a string which has format like // ::
func SplitDate(ddate string) []string {
	var strsplit []string

	for _, v := range strings.Split(ddate, " ") {
		for _, w := range strings.Split(v, ":") {
			for _, x := range strings.Split(w, "/") {
				strsplit = append(strsplit, x)
				//fmt.Println(strsplit2)
			}
		}
	}
	return strsplit
}

// IndexArray search a value in an array of string and returns its index
func IndexArray(array []string, searchValue string) int {
	for index, value := range array {
		if value == searchValue {
			return index
		}
	}
	return -1
}

// DayInMonth return the number of days in the month/year given in parameters
func DayInMonth(year int, month int) int {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
}

// CheckDateFormat check that format and adate share the same format
func CheckDateFormat(adate string, format string) bool {
	//var tab []int
	//var d ddate
	var year, month, day, hour, minute, second int
	var err error
	var r, r2 *regexp.Regexp

	tabDate := SplitDate(adate)
	fmtDate := SplitDate(format)

	if len(tabDate) != len(fmtDate) {
		fmt.Println("Error format and date don't have the same fields.")
		return false
	}

	indexYear := IndexArray(fmtDate, "%YYYY")
	indexMonth := IndexArray(fmtDate, "%MM")
	indexDay := IndexArray(fmtDate, "%DD")
	indexHour := IndexArray(fmtDate, "%hh")
	indexMinute := IndexArray(fmtDate, "%mm")
	indexSecond := IndexArray(fmtDate, "%ss")

	if indexYear != -1 && len(tabDate[indexYear]) != 0 {
		// Pas de mention de YYYY dans le format
		//year, err = strconv.Atoi(tabDate[indexYear])
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexYear]) && !r2.MatchString(tabDate[indexYear]) {
			return false
		}
		if r2.MatchString(tabDate[indexYear]) {
			year, err = strconv.Atoi(tabDate[indexYear])
		}
	}

	if indexMonth != -1 && len(tabDate[indexMonth]) != 0 {
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexMonth]) && !r2.MatchString(tabDate[indexMonth]) {
			return false
		}
		if r2.MatchString(tabDate[indexMonth]) {
			month, err = strconv.Atoi(tabDate[indexMonth])
			if err != nil {
				return false
			}
			if month < 1 || month > 12 {
				return false
			}
		}
	}

	if indexDay != -1 && len(tabDate[indexDay]) != 0 {
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexDay]) && !r2.MatchString(tabDate[indexDay]) {
			return false
		}
		if r2.MatchString(tabDate[indexDay]) {
			day, err = strconv.Atoi(tabDate[indexDay])
			if err != nil {
				return false
			}
			if day > DayInMonth(year, month) || day < 1 {
				return false
			}
		}
	}

	if indexHour != -1 && len(tabDate[indexHour]) != 0 {
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexHour]) && !r2.MatchString(tabDate[indexHour]) {
			return false
		}
		if r2.MatchString(tabDate[indexHour]) {
			hour, err = strconv.Atoi(tabDate[indexHour])
			if err != nil {
				return false
			}
			if hour < 0 || hour > 23 {
				return false
			}
		}
	}
	if indexMinute != -1 && len(tabDate[indexMinute]) != 0 {
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexMinute]) && !r2.MatchString(tabDate[indexMinute]) {
			return false
		}
		if r2.MatchString(tabDate[indexMinute]) {
			minute, err = strconv.Atoi(tabDate[indexMinute])
			if err != nil {
				return false
			}
			if minute < 0 || minute > 59 {
				return false
			}
		}
	}

	if indexSecond != -1 && len(tabDate[indexSecond]) != 0 {
		r, _ = regexp.Compile("(\\+|-){1}[0-9]+")
		r2, _ = regexp.Compile("^[0-9]+$")
		if !r.MatchString(tabDate[indexSecond]) && !r2.MatchString(tabDate[indexSecond]) {
			return false
		}
		if r2.MatchString(tabDate[indexSecond]) {
			second, err = strconv.Atoi(tabDate[indexSecond])
			if err != nil {
				return false
			}
			if second < 0 || second > 59 {
				return false
			}
		}
	}

	if indexYear == -1 && indexMonth == -1 && indexDay == -1 && indexHour == -1 && indexMinute == -1 && indexSecond == -1 {
		return false
	}

	return true
}

func extractDateAndChange(partDate string, defaultValue int) (value int, change int, err error) {
	regexpstr := "(\\+|-){1}[0-9]+"
	r, err := regexp.Compile(regexpstr)
	value = defaultValue

	if r.MatchString(partDate) {
		// +/-...
		change, err = strconv.Atoi(partDate)
	} else {
		if len(partDate) != 0 {
			value, err = strconv.Atoi(partDate)
			if err != nil {
				err = errors.New("Not an int")
			}
		}
	}
	return
}

// CreateDate returns a time.Time object represented by the parameters adate formatted with ifmt. If enddate is true,
// the time returned will ...
func CreateDate(adate string, ifmt string, tz string, begindate bool, enddate bool) (time.Time, error) {
	var err error
	var year, month, day, hour, minute, second int
	var yearChg, monthChg, dayChg, hourChg, minuteChg, secondChg int
	var location, locationOrigin *time.Location

	locationOrigin, err = time.LoadLocation("Local")
	if tz != "" {
		location, err = time.LoadLocation(tz)
	} else {
		location = locationOrigin
	}

	if err != nil {
		return time.Now(), err
	}

	if !CheckDateFormat(adate, ifmt) {
		return time.Now(), errors.New("Date and Format incompatible")
	}

	tabDate := SplitDate(adate)
	fmtDate := SplitDate(ifmt)

	indexYear := IndexArray(fmtDate, "%YYYY")
	indexMonth := IndexArray(fmtDate, "%MM")
	indexDay := IndexArray(fmtDate, "%DD")
	indexHour := IndexArray(fmtDate, "%hh")
	indexMinute := IndexArray(fmtDate, "%mm")
	indexSecond := IndexArray(fmtDate, "%ss")

	year = time.Now().Year()
	if indexYear != -1 {
		year, yearChg, err = extractDateAndChange(tabDate[indexYear], time.Now().Year())
		if err != nil {
			panic(adate)
		}
	}

	month = int(time.Now().Month())
	if indexMonth != -1 {
		month, monthChg, err = extractDateAndChange(tabDate[indexMonth], int(time.Now().Month()))
		if err != nil {
			panic(adate)
		}
	}

	day = time.Now().Day()
	if indexDay != -1 {
		day, dayChg, err = extractDateAndChange(tabDate[indexDay], time.Now().Day())
		if err != nil {
			panic(adate)
		}
	}

	hour = time.Now().Hour()
	if indexHour != -1 {
		hour, hourChg, err = extractDateAndChange(tabDate[indexHour], time.Now().Hour())
		if err != nil {
			panic(adate)
		}
	}

	minute = time.Now().Minute()
	if indexMinute != -1 {
		minute, minuteChg, err = extractDateAndChange(tabDate[indexMinute], time.Now().Minute())
		if err != nil {
			panic(adate)
		}
	}

	second = time.Now().Second()
	if indexSecond != -1 {
		second, secondChg, err = extractDateAndChange(tabDate[indexSecond], time.Now().Second())
		if err != nil {
			panic(adate)
		}
	}

	// Need to calculate the operations on the date
	if enddate || begindate {
		// treatFields: stop to calculate the begin and end of date fields when a value have been precised
		treatFields := false
		// setenddate allows to know if the day need to be calculated (if enddate)
		setenddate := false
		// If second is not specified
		if indexSecond != -1 {
			if len(tabDate[indexSecond]) == 0 {
				if begindate {
					second = 0
				}
				if enddate {
					second = 59
				}
			} else {
				treatFields = true
			}
		}

		if indexMinute != -1 {
			if !treatFields && len(tabDate[indexMinute]) == 0 {
				if begindate {
					minute = 0
				}
				if enddate {
					minute = 59
				}
			} else {
				treatFields = true
			}
		}

		if indexHour != -1 {
			if !treatFields && len(tabDate[indexHour]) == 0 {
				if begindate {
					hour = 0
				}
				if enddate {
					hour = 23
				}
			} else {
				treatFields = true
			}
		}

		if indexDay != -1 {
			if !treatFields && len(tabDate[indexDay]) == 0 {
				if begindate {
					day = 1
				}
				if enddate {
					setenddate = true
				}
			} else {
				treatFields = true
			}
		}

		if indexMonth != -1 {
			if !treatFields && len(tabDate[indexMonth]) == 0 {
				if begindate {
					month = 1
				}
				if enddate {
					month = 12
				}
			} else {
				treatFields = true
			}
		}

		if setenddate {
			day = DayInMonth(year, month)
		}
	}
	// fmt.Println("year=", year, "month", month, "day", day, "hour", hour, "minute", minute, "second", second)
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, locationOrigin).AddDate(yearChg, monthChg, dayChg).Add(time.Duration(secondChg)*time.Second + time.Duration(minuteChg)*time.Minute + time.Duration(hourChg)*time.Hour).In(location), err
}

// DoubleReplace replaces keyword by a string representing a scanfformat (like %2d)
// in the string fmtstr. Finally the scanfformat is expandined to data
func DoubleReplace(fmtstr string, keyword string, scanfformat string, data int) string {
	finalValue := fmt.Sprintf(scanfformat, data)
	res := fmtstr
	if strings.ContainsAny(fmtstr, keyword) {
		res = strings.ReplaceAll(fmtstr, keyword, finalValue)
		// res = fmt.Sprintf(res, data)
	}

	return res
}

// ApplyFormat return a string representing the date date formtted by fmtstr
// Fields recognized : YYYY MM DD hh mm ss
func ApplyFormat(fmtstr string, date time.Time) string {
	res := DoubleReplace(fmtstr, "%YYYY", "%4d", date.Year())
	res = DoubleReplace(res, "%MM", "%02d", int(date.Month()))
	res = DoubleReplace(res, "%DD", "%02d", date.Day())
	res = DoubleReplace(res, "%hh", "%02d", date.Hour())
	res = DoubleReplace(res, "%mm", "%02d", date.Minute())
	res = DoubleReplace(res, "%ss", "%02d", date.Second())
	// res = strings.ReplaceAll(res, "@ts", strconv.FormatInt(date.Unix(), 10))
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

func CalcLine(tmpl string, beginTime time.Time, endTime time.Time) string {
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
		panic(err)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, d)
	s := doc.String()
	// err = t.Execute(os.Stdout, d)
	if err != nil {
		panic(err)
	}
	return s
}
