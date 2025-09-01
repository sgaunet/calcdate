// Package calcdate provides date calculation and manipulation functionality.
package calcdate

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// SecondsPerHour is the number of seconds in an hour.
	SecondsPerHour = 3600
	// MinutesPerHour is the number of minutes in an hour.
	MinutesPerHour = 60
	// MonthsPerYear is the number of months in a year.
	MonthsPerYear = 12
	// HoursPerDay is the number of hours in a day.
	HoursPerDay = 24
)

// Date represents a date with timezone support and relative time calculations.
type Date struct {
	now      func() time.Time
	dateStr  string // represent the date string, ex: 2020/02/01 -1::
	ifmt     string // input format, default to %YYYY/%MM/%DD %hh:%mm:%ss
	tz       string // timezone
	location *time.Location

	submatch    []string
	subExpNames []string

	// time fields after updates
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

	// where fields are placed in submatch
	indexYear   int
	indexMonth  int
	indexDay    int
	indexHour   int
	indexMinute int
	indexSecond int
}

// NewDate creates a new Date instance from a date string, input format, and timezone.
func NewDate(date string, ifmt string, tz string) (*Date, error) {
	d := Date{
		now:     time.Now,
		dateStr: date,
		ifmt:    ifmt,
		tz:      tz,
	}

	err := d.initRe() // init submatch and subExpNames
	if err != nil {
		return nil, err
	}

	err = d.initIndex() // init index*
	if err != nil {
		return nil, err
	}

	err = d.initLocation()
	if err != nil {
		return nil, err
	}

	err = d.init() // init time fields and relative time fields
	if err != nil {
		return nil, err
	}

	d.applyRelativ() // apply relative time
	return &d, err
}

// String returns the date according the format YYYY/MM/DD hh:mm:ss.
func (d *Date) String() string {
	t := time.Date(d.year, time.Month(d.month), d.day, d.hour, d.minute, d.second, 0, d.location).In(d.location)
	return t.Format("2006/01/02 15:04:05")
	// return t.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
}

// Time returns the date as a time.Time.
func (d *Date) Time() time.Time {
	return time.Date(d.year, time.Month(d.month), d.day, d.hour, d.minute, d.second, 0, d.location).In(d.location)
}

// Before will return true if t is before d.
func (d *Date) Before(t *Date) bool {
	return d.Time().Before(t.Time())
}

// Add will add the duration.
func (d *Date) Add(dur time.Duration) {
	loc, _ := time.LoadLocation(d.tz)
	updatedTime := d.Time().In(loc).Add(dur)
	d.year = updatedTime.Year()
	d.month = int(updatedTime.Month())
	d.day = updatedTime.Day()
	d.hour = updatedTime.Hour()
	d.minute = updatedTime.Minute()
	d.second = updatedTime.Second()
}

// Format will return a string formatted with ofmt.
// Possible annotations : %YYYY %MM %DD %hh %mm %ss.
func (d *Date) Format(ofmt string) string {
	loc, _ := time.LoadLocation(d.tz)
	timeInLoc := d.Time().In(loc)
	res := doubleReplace(ofmt, "%YYYY", "%4d", timeInLoc.Year())
	res = doubleReplace(res, "%MM", "%02d", int(timeInLoc.Month()))
	res = doubleReplace(res, "%DD", "%02d", timeInLoc.Day())
	res = doubleReplace(res, "%hh", "%02d", timeInLoc.Hour())
	res = doubleReplace(res, "%mm", "%02d", timeInLoc.Minute())
	res = doubleReplace(res, "%ss", "%02d", timeInLoc.Second())
	res = strings.ReplaceAll(res, "%z", d.GetOffset())
	res = strings.ReplaceAll(res, "%Z", d.GetTZAbr())
	res = strings.ReplaceAll(res, "@ts", strconv.FormatInt(timeInLoc.Unix(), 10))
	return res
}

// GetOffset returns the timezone offset in the format +/-HH:MM.
func (d *Date) GetOffset() string {
	_, offset := d.Time().Zone()
	hours := offset / SecondsPerHour
	minutes := (offset % SecondsPerHour) / MinutesPerHour
	if offset < 0 {
		return fmt.Sprintf("-%02d:%02d", -hours, -minutes)
	}
	return fmt.Sprintf("+%02d:%02d", hours, minutes)
}

// GetTZAbr returns the timezone abbreviation.
func (d *Date) GetTZAbr() string {
	loc, _ := time.LoadLocation(d.tz)
	timeInLoc := d.Time().In(loc)
	name, _ := timeInLoc.Zone()
	return name
}

// SetBeginDate will calculate the begindate based on unknown specified time fields.
// For example, if NewDate has been created with // -2:: , seconds will be set to 0, minutes will be set to 0.
func (d *Date) SetBeginDate() *Date {
	if d.indexSecond == -1 {
		d.second = 0
	} else {
		return d
	}

	if d.indexMinute == -1 {
		d.minute = 0
	} else {
		return d
	}

	if d.indexHour == -1 {
		d.hour = 0
	} else {
		return d
	}

	if d.indexDay == -1 {
		d.day = 1
	} else {
		return d
	}

	if d.indexMonth == -1 {
		d.month = 1
	}
	return d
}

// SetEndDate will calculate the end of datefor the empty field.
// For example, if NewDate has been created with // -2:: , seconds will be set to 59, minutes will be set to 59.
func (d *Date) SetEndDate() *Date {
	if d.indexSecond == -1 {
		d.second = 59
	} else {
		return d
	}

	if d.indexMinute == -1 {
		d.minute = 59
	} else {
		return d
	}

	if d.indexHour == -1 {
		d.hour = 23
	} else {
		return d
	}

	if d.indexDay == -1 {
		d.day = DayInMonth(d.year, d.month)
	} else {
		return d
	}

	if d.indexMonth == -1 {
		d.month = MonthsPerYear
	}
	return d
}

// SetEndMonth sets the date to the end of the month.
func (d *Date) SetEndMonth() *Date {
	d.second = 59
	d.minute = 59
	d.hour = 23

	if d.month == 0 {
		d.month = MonthsPerYear
	}

	d.day = DayInMonth(d.year, d.month)
	return d
}

// SetBeginMonth sets the date to the beginning of the month.
func (d *Date) SetBeginMonth() *Date {
	d.second = 0
	d.minute = 0
	d.hour = 0

	d.day = 1
	if d.month == 0 {
		d.month = 1
	}
	return d
}

// AddMonth adds n months to the date.
func (d *Date) AddMonth(n int) *Date {
	d.year += n / MonthsPerYear
	if d.month+(n%MonthsPerYear) > MonthsPerYear {
		d.year++
		d.month = d.month + (n % MonthsPerYear) - MonthsPerYear
	} else {
		d.month += n
	}
	daysInMonth := DayInMonth(d.year, d.month)
	if d.day > daysInMonth {
		d.day = daysInMonth
	}
	return d
}

// AddYear adds n years to the date.
func (d *Date) AddYear(n int) *Date {
	updatedTime := d.Time().AddDate(n, 0, 0)
	d.year = updatedTime.Year()
	d.month = int(updatedTime.Month())
	d.day = updatedTime.Day()
	d.hour = updatedTime.Hour()
	d.minute = updatedTime.Minute()
	d.second = updatedTime.Second()
	return d
}

// AddDay adds n days to the date.
func (d *Date) AddDay(n int) *Date {
	updatedTime := d.Time().AddDate(0, 0, n)
	d.year = updatedTime.Year()
	d.month = int(updatedTime.Month())
	d.day = updatedTime.Day()
	d.hour = updatedTime.Hour()
	d.minute = updatedTime.Minute()
	d.second = updatedTime.Second()
	return d
}

// GetInterval returns the absolute duration between two dates.
func GetInterval(d1 *Date, d2 *Date) time.Duration {
	diff := d1.Time().Sub(d2.Time())
	if diff > 0 {
		return diff
	}
	return -diff
}

// ListTZ prints all available timezones to stdout.
func ListTZ() {
	var zoneDirs = map[string]string{
		"android":   "/system/usr/share/zoneinfo/",
		"darwin":    "/usr/share/zoneinfo/",
		"dragonfly": "/usr/share/zoneinfo/",
		"freebsd":   "/usr/share/zoneinfo/",
		"linux":     "/usr/share/zoneinfo/",
		"netbsd":    "/usr/share/zoneinfo/",
		"openbsd":   "/usr/share/zoneinfo/",
		"solaris":   "/usr/share/lib/zoneinfo/",
	}

	var timeZones []string

	// Reads the Directory corresponding to the OS
	dirFile, _ := os.ReadDir(zoneDirs[runtime.GOOS])
	for _, i := range dirFile {
		// Checks if starts with Capital Letter
		if i.Name() == (strings.ToUpper(i.Name()[:1]) + i.Name()[1:]) {
			if i.IsDir() {
				// Recursive read if directory
				subFiles, err := os.ReadDir(zoneDirs[runtime.GOOS] + i.Name())
				if err != nil {
					log.Fatal(err)
				}

				for _, s := range subFiles {
					// Appends the path to timeZones var
					timeZones = append(timeZones, i.Name()+"/"+s.Name())
				}
			}
			// Appends the path to timeZones var
			timeZones = append(timeZones, i.Name())
		}
	}
	// Loop over timezones and Check Validity, Delete entry if invalid.
	// Range function doesnt work with changing length.
	for i := 0; i < len(timeZones); i++ {
		_, err := time.LoadLocation(timeZones[i])
		if err != nil {
			// newSlice = timeZones[:n]  timeZones[n+1:]
			timeZones = append(timeZones[:i], timeZones[i+1:]...)
			continue
		}
	}
	// Now we Can range timeZones for printing
	for _, i := range timeZones {
		fmt.Println(i)
	}
}

//nolint:unused // Used in tests
func newDateWithSpecificNowFct(date string, ifmt string, tz string, nowFct func() time.Time) (*Date, error) {
	d := Date{
		now:     nowFct,
		dateStr: date,
		ifmt:    ifmt,
		tz:      tz,
	}

	err := d.initRe() // init submatch and subExpNames
	if err != nil {
		return nil, err
	}

	err = d.initIndex() // init index*
	if err != nil {
		return nil, err
	}

	err = d.initLocation()
	if err != nil {
		return nil, err
	}

	err = d.init() // init time fields and relative time fields
	if err != nil {
		return nil, err
	}

	d.applyRelativ() // apply relative time
	return &d, err
}

func (d *Date) initIndex() error {
	err := d.initIndexSecond()
	if err != nil {
		return err
	}

	err = d.initIndexMinute()
	if err != nil {
		return err
	}

	err = d.initIndexHour()
	if err != nil {
		return err
	}

	err = d.initIndexDay()
	if err != nil {
		return err
	}

	err = d.initIndexMonth()
	if err != nil {
		return err
	}

	err = d.initIndexYear()
	if err != nil {
		return err
	}

	return nil
}

func (d *Date) initIndexYear() error {
	var err error

	d.indexYear, err = d.findIndexOf("Year")
	return err
}

func (d *Date) initIndexMonth() error {
	var err error

	d.indexMonth, err = d.findIndexOf("Month")
	return err
}

func (d *Date) initIndexDay() error {
	var err error

	d.indexDay, err = d.findIndexOf("Day")
	return err
}

func (d *Date) initIndexHour() error {
	var err error

	d.indexHour, err = d.findIndexOf("Hour")
	return err
}

func (d *Date) initIndexMinute() error {
	var err error

	d.indexMinute, err = d.findIndexOf("Minute")
	return err
}

func (d *Date) initIndexSecond() error {
	var err error

	d.indexSecond, err = d.findIndexOf("Second")
	return err
}

// initUnitTime initialise ptrToInit and ptrRelativToInit.
// if d.submatch[idx] is empty, ptrToInit=defaultValue and ptrRelativToInit=0
// if d.submatch[idx] is (+-){1}(\d)+, ptrToInit=defaultValue and ptrRelativToInit=(+-){1}(\d)+
// if d.submatch[idx] is (\d)+, ptrToInit=(\d)+ and ptrRelativToInit=0.
func (d *Date) initUnitTime(defaultValue int, ptrToInit *int, ptrRelativToInit *int, idx int) error {
	*ptrToInit = defaultValue
	*ptrRelativToInit = 0
	if idx <= 0 {
		return nil
	}

	value := d.submatch[idx]
	if value == "" {
		return nil
	}

	convert, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}

	*ptrToInit = convert
	*ptrRelativToInit = 0

	if len(value) > 0 && (value[0] == '+' || value[0] == '-') {
		*ptrToInit = defaultValue
		*ptrRelativToInit = convert
	}

	return nil
}

func (d *Date) initRe() error {
	regexpFromIfmt := createRegexpFromIfmt(d.ifmt)
	rr, err := regexp.Compile(regexpFromIfmt)
	if err != nil {
		return fmt.Errorf("regexp.Compile: %w", err)
	}
	d.submatch = rr.FindStringSubmatch(d.dateStr)
	d.subExpNames = rr.SubexpNames()
	return nil
}

// initIndex initialise time fields.
func (d *Date) init() error {
	now := d.now()
	if err := d.initUnitTime(now.Second(), &d.second, &d.relativeSecond, d.indexSecond); err != nil {
		return err
	}
	if err := d.initUnitTime(now.Minute(), &d.minute, &d.relativeMinute, d.indexMinute); err != nil {
		return err
	}
	if err := d.initUnitTime(now.Hour(), &d.hour, &d.relativeHour, d.indexHour); err != nil {
		return err
	}
	if err := d.initUnitTime(now.Day(), &d.day, &d.relativeDay, d.indexDay); err != nil {
		return err
	}
	if err := d.initUnitTime(int(now.Month()), &d.month, &d.relativeMonth, d.indexMonth); err != nil {
		return err
	}
	if err := d.initUnitTime(now.Year(), &d.year, &d.relativeYear, d.indexYear); err != nil {
		return err
	}

	return nil
}

// initLocation will initialize the location time (local for default).
func (d *Date) initLocation() error {
	locationOrigin, err := time.LoadLocation("Local")
	if err != nil {
		return fmt.Errorf("time.LoadLocation: %w", err)
	}

	if d.tz != "" {
		d.location, err = time.LoadLocation(d.tz)
		if err != nil {
			return fmt.Errorf("time.LoadLocation: %w", err)
		}
	} else {
		d.location = locationOrigin
	}
	return nil
}

func (d *Date) applyRelativ() {
	t := d.Time()
	updatedTime := t.AddDate(d.relativeYear, d.relativeMonth, d.relativeDay).Add(
		time.Duration(d.relativeHour)*time.Hour +
			time.Duration(d.relativeMinute)*time.Minute +
			time.Duration(d.relativeSecond)*time.Second)
	d.year = updatedTime.Year()
	d.month = int(updatedTime.Month())
	d.day = updatedTime.Day()
	d.hour = updatedTime.Hour()
	d.minute = updatedTime.Minute()
	d.second = updatedTime.Second()
}