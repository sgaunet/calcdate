package calcdatelib

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Date struct {
	dateStr  string // represent the date string, ex: 2020/02/01 -1::
	ifmt     string // input format, default to %YYYY/%MM/%DD %hh:%mm:%ss
	tz       string // timezone
	location *time.Location

	// time fields
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int

	// relative time fields +<int> or -<int>
	relativeYear   int
	relativeMonth  int
	relativeDay    int
	relativeHour   int
	relativeMinute int
	relativeSecond int

	// where fields are placed in dateStr
	indexYear   int
	indexMonth  int
	indexDay    int
	indexHour   int
	indexMinute int
	indexSecond int
}

func NewDate(date string, ifmt string, tz string) (*Date, error) {
	d := Date{
		dateStr: date,
		ifmt:    ifmt,
		tz:      tz,
	}
	err := d.initLocation()
	if err != nil {
		return nil, err
	}
	d.initIndex() // to intialize index*
	if isFormatOK, _ := d.checkDateFormat(d.dateStr, d.ifmt); !isFormatOK {
		return nil, errors.New("date and format are incompatibles")
	}
	err = d.initRelativeTimeFields()
	if err != nil {
		return nil, err
	}
	err = d.initTimeFields()
	if err != nil {
		return nil, err
	}
	// err = d.scanIfmt()
	d.applyRelativ()
	return &d, err
}

// initIndex will initialise where fields of time are placed based on ifmt format
func (d *Date) initIndex() {
	fmtDate := splitDate(d.ifmt)
	d.indexYear = indexArray(fmtDate, "%YYYY")
	d.indexMonth = indexArray(fmtDate, "%MM")
	d.indexDay = indexArray(fmtDate, "%DD")
	d.indexHour = indexArray(fmtDate, "%hh")
	d.indexMinute = indexArray(fmtDate, "%mm")
	d.indexSecond = indexArray(fmtDate, "%ss")
}

// initLocation will initialize the location time (local for default)
func (d *Date) initLocation() error {
	locationOrigin, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}
	if d.tz != "" {
		d.location, err = time.LoadLocation(d.tz)
	} else {
		d.location = locationOrigin
	}
	return err
}

// String return the date according the format YYYY/%%/DD hh:mm:ss
func (d *Date) String() string {
	t := time.Date(d.year, time.Month(d.month), d.day, d.hour, d.minute, d.second, 0, d.location).In(d.location)
	return t.Format("2006/01/02 15:04:05")
	// return t.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
}

// Time returns the date as a time.Time
func (d *Date) Time() time.Time {
	return time.Date(d.year, time.Month(d.month), d.day, d.hour, d.minute, d.second, 0, d.location).In(d.location)
}

// Before will return true if t is before d
func (d *Date) Before(t *Date) bool {
	return d.Time().Before(t.Time())
}

// Add will add the duration
func (d *Date) Add(dur time.Duration) {
	loc, _ := time.LoadLocation(d.tz)
	new := d.Time().In(loc).Add(dur)
	d.year = new.Year()
	d.month = int(new.Month())
	d.day = new.Day()
	d.hour = new.Hour()
	d.minute = new.Minute()
	d.second = new.Second()
}

// Format will return a string formatted with ofmt
// Possible annotations : %YYYY %MM %DD %hh %mm %ss
func (d *Date) Format(ofmt string) string {
	loc, _ := time.LoadLocation(d.tz)
	new := d.Time().In(loc)
	res := doubleReplace(ofmt, "%YYYY", "%4d", new.Year())
	res = doubleReplace(res, "%MM", "%02d", int(new.Month()))
	res = doubleReplace(res, "%DD", "%02d", new.Day())
	res = doubleReplace(res, "%hh", "%02d", new.Hour())
	res = doubleReplace(res, "%mm", "%02d", new.Minute())
	res = doubleReplace(res, "%ss", "%02d", new.Second())
	// res = strings.ReplaceAll(res, "@ts", strconv.FormatInt(date.Unix(), 10))
	return res
}

// extractRelaiveNumber will return an int if string represent an integer with + ot - before the interger
func extractRelativeNumber(str string) (relativNumber int, err error) {
	r, err := regexp.Compile(`(\+|-){1}[0-9]+`)
	if err != nil {
		return
	}
	isNumber, err := regexp.Compile("^[0-9]+$")
	if err != nil {
		return
	}

	if !r.MatchString(str) && !isNumber.MatchString(str) {
		return 0, fmt.Errorf("%s is not a number", str)
	}
	if r.MatchString(str) {
		relativNumber, err = strconv.Atoi(str)
		if err != nil {
			return
		}
	}
	return
}

// initTimeFieldsRelativeNumbers will initialize all relative* fields based on the specified date string
func (d *Date) initRelativeTimeFields() (err error) {
	tabDate := splitDate(d.dateStr)
	if d.indexYear != -1 && len(tabDate[d.indexYear]) != 0 {
		d.relativeYear, err = extractRelativeNumber(tabDate[d.indexYear])
		if err != nil {
			return
		}
	}
	if d.indexMonth != -1 && len(tabDate[d.indexMonth]) != 0 {
		d.relativeMonth, err = extractRelativeNumber(tabDate[d.indexMonth])
		if err != nil {
			return
		}
	}
	if d.indexDay != -1 && len(tabDate[d.indexDay]) != 0 {
		d.relativeDay, err = extractRelativeNumber(tabDate[d.indexDay])
		if err != nil {
			return
		}
	}
	if d.indexHour != -1 && len(tabDate[d.indexHour]) != 0 {
		d.relativeHour, err = extractRelativeNumber(tabDate[d.indexHour])
		if err != nil {
			return
		}
	}
	if d.indexMinute != -1 && len(tabDate[d.indexMinute]) != 0 {
		d.relativeMinute, err = extractRelativeNumber(tabDate[d.indexMinute])
		if err != nil {
			return
		}
	}
	if d.indexSecond != -1 && len(tabDate[d.indexSecond]) != 0 {
		d.relativeSecond, err = extractRelativeNumber(tabDate[d.indexSecond])
		if err != nil {
			return
		}
	}
	return
}

// checkDateFormat checks that format and adate share the same format
// format should have a format based on %YYYY %MM %DD %hh %mm %ss
func (d *Date) checkDateFormat(adate string, format string) (formatOK bool, err error) {
	tabDate := splitDate(adate)
	fmtDate := splitDate(format)

	if len(tabDate) != len(fmtDate) {
		// fmt.Println("Error format and date don't have the same fields.")
		return false, err
	}

	if d.indexYear == -1 && d.indexMonth == -1 && d.indexDay == -1 && d.indexHour == -1 && d.indexMinute == -1 && d.indexSecond == -1 {
		return false, err
	}
	return true, err
}

func (d *Date) initTimeFields() (err error) {
	tabDate := splitDate(d.dateStr)

	if d.indexYear != -1 {
		d.year, err = calculateUnitOfTime(tabDate[d.indexYear], time.Now().In(d.location).Year())
		if err != nil {
			return err
		}
	}
	if d.indexMonth != -1 {
		d.month, err = calculateUnitOfTime(tabDate[d.indexMonth], int(time.Now().In(d.location).Month()))
		if err != nil {
			return err
		}
	}
	if d.indexDay != -1 {
		d.day, err = calculateUnitOfTime(tabDate[d.indexDay], time.Now().In(d.location).Day())
		if err != nil {
			return err
		}
	}
	if d.indexHour != -1 {
		d.hour, err = calculateUnitOfTime(tabDate[d.indexHour], time.Now().In(d.location).Hour())
		if err != nil {
			return err
		}
	}
	if d.indexMinute != -1 {
		d.minute, err = calculateUnitOfTime(tabDate[d.indexMinute], time.Now().In(d.location).Minute())
		if err != nil {
			return err
		}
	}
	if d.indexSecond != -1 {
		d.second, err = calculateUnitOfTime(tabDate[d.indexSecond], time.Now().In(d.location).Second())
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Date) applyRelativ() {
	t := d.Time()
	t.AddDate(d.relativeYear, d.relativeMonth, d.relativeDay).Add(time.Duration(d.relativeHour)*time.Hour + time.Duration(d.relativeMinute)*time.Minute + time.Duration(d.relativeSecond)*time.Second)
	d.year = t.Year()
	d.month = int(t.Month())
	d.day = t.Day()
	d.hour = t.Hour()
	d.minute = t.Minute()
	d.second = t.Second()
}

// SetBeginDate will calculate the begindate based on unknown specified time fields
// For example, if NewDate has been created with // -2:: , seconds will be set to 0, minutes will be set to 0
func (d *Date) SetBeginDate() *Date {
	tabDate := splitDate(d.dateStr)
	// fmtDate := splitDate(ifmt)
	// treatFields: stop to calculate the begin and end of date fields when a value have been precised
	treatFields := false
	// If second is not specified
	if d.indexSecond != -1 {
		if len(tabDate[d.indexSecond]) == 0 {
			d.second = 0
		} else {
			treatFields = true
		}
	}

	if d.indexMinute != -1 {
		if !treatFields && len(tabDate[d.indexMinute]) == 0 {
			d.minute = 0
		} else {
			treatFields = true
		}
	}

	if d.indexHour != -1 {
		if !treatFields && len(tabDate[d.indexHour]) == 0 {
			d.hour = 0
		} else {
			treatFields = true
		}
	}

	if d.indexDay != -1 {
		if !treatFields && len(tabDate[d.indexDay]) == 0 {
			d.day = 1
		} else {
			treatFields = true
		}
	}

	if d.indexMonth != -1 {
		if !treatFields && len(tabDate[d.indexMonth]) == 0 {
			d.month = 1
		} else {
			treatFields = true
		}
	}
	return d
}

// SetEndDate will calculate the end of datefor the empty field
// For example, if NewDate has been created with // -2:: , seconds will be set to 59, minutes will be set to 59
func (d *Date) SetEndDate() *Date {
	tabDate := splitDate(d.dateStr)
	// fmtDate := splitDate(ifmt)
	// treatFields: stop to calculate the begin and end of date fields when a value have been precised
	treatFields := false
	// If second is not specified
	if d.indexSecond != -1 {
		if len(tabDate[d.indexSecond]) == 0 {
			d.second = 59
		} else {
			treatFields = true
		}
	}

	if d.indexMinute != -1 {
		if !treatFields && len(tabDate[d.indexMinute]) == 0 {
			d.minute = 59
		} else {
			treatFields = true
		}
	}

	if d.indexHour != -1 {
		if !treatFields && len(tabDate[d.indexHour]) == 0 {
			d.hour = 23
		} else {
			treatFields = true
		}
	}

	if d.indexMonth != -1 {
		if !treatFields && len(tabDate[d.indexMonth]) == 0 {
			d.month = 12
		}
	}

	if d.indexDay != -1 {
		if !treatFields && len(tabDate[d.indexDay]) == 0 {
			d.day = DayInMonth(d.year, d.month)
		} else {
			treatFields = true
		}
	}
	return d
}
