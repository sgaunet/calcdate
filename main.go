// Package main provides a command-line utility for date calculations and manipulations
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
)

func completeDate(adate string) string {
	resDate := adate
	// If date format is like "HH:MM:SS" without date part, add "//" prefix
	r := regexp.MustCompile(`^(\+|-)*[0-9]*:(\+|-)*[0-9]*:(\+|-)*[0-9]*$`)
	if r.MatchString(resDate) {
		resDate = "// " + resDate
	}

	// If date is like "YYYY/MM/DD" without time part, add "::" time part
	r = regexp.MustCompile(`^(\+|-)*[0-9]*/(\+|-)*[0-9]*/(\+|-)*[0-9]*$`)
	if r.MatchString(resDate) {
		resDate += " ::"
	}
	return resDate
}

var version = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var begindate, enddate, separator, ifmt, ofmt, tz string
	// var endTime, beginTime time.Time
	var vOption, listTZ bool
	var interval time.Duration
	var tmpl string
	var err error
	var rangeDate = false

	flag.StringVar(&begindate, "b", "// ::", "Begin date")
	flag.StringVar(&enddate, "e", "", "End date")
	flag.StringVar(&separator, "s", " ", "Separator")
	flag.StringVar(&tz, "tz", "Local", "Input timezone")
	flag.StringVar(&ifmt, "ifmt", "%YYYY/%MM/%DD %hh:%mm:%ss", "Input Format (%YYYY/%MM/%DD %hh:%mm:%ss)")
	// Define output format with timestamp and timezone support
	flag.StringVar(&ofmt, "ofmt", "%YYYY/%MM/%DD %hh:%mm:%ss",
		"Input Format (%YYYY/%MM/%DD %hh:%mm:%ss), use @ts for timestamp %z for offset %Z for timezone")
	flag.DurationVar(&interval, "i", 0, "Interval (Ex: 1m or 1h or 15s)")
	// Define default template for interval rendering
	defaultTemplate := "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - "
	defaultTemplate += "{{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} "
	defaultTemplate += "{{ .BeginTime.Unix }} {{ .EndTime.Unix }}"
	flag.StringVar(&tmpl, "tmpl", defaultTemplate, "Used only with -i option")
	flag.BoolVar(&listTZ, "list-tz", false, "List timezones")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.Parse()

	if listTZ {
		calcdatelib.ListTZ()
		os.Exit(0)
	}

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if enddate != "" && begindate != "" {
		rangeDate = true
	}

	// -i option can be used only with tho dates (begin/end)
	if interval != 0 && !rangeDate {
		fmt.Println("specify a range date")
		os.Exit(1)
	}
	// If date is like :: or //  TODO control the format
	begindate = completeDate(begindate)
	enddate = completeDate(enddate)

	beginTime, err := calcdatelib.NewDate(begindate, ifmt, tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format date begindate KO: %v\n", err)
		os.Exit(1)
	}

	if rangeDate {
		processRangeDateCase(begindate, enddate, ifmt, tz, interval, ofmt, separator, tmpl)
	} else {
		fmt.Println(beginTime.Format(ofmt))
	}
}

// processRangeDateCase handles the logic for date range processing
// processRangeDateCase handles date ranges with optional interval processing.
func processRangeDateCase(
	begindate, enddate, ifmt, tz string,
	interval time.Duration,
	ofmt, separator, tmpl string,
) {
	var err error
	beginTime, err := calcdatelib.NewDate(begindate, ifmt, tz)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	endTime, err := calcdatelib.NewDate(enddate, ifmt, tz)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	endTime.SetEndDate()
	beginTime.SetBeginDate()
	if interval == 0 {
		// Print range
		fmt.Printf("%s%s%s\n", beginTime.Format(ofmt), separator, endTime.Format(ofmt))
	} else {
		intervals, err := calcdatelib.RenderIntervalLines(*beginTime, *endTime, interval, tmpl)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for idx := range intervals {
			fmt.Println(intervals[idx])
		}
	}
}
