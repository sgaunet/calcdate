package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
)

func completeDate(adate string) (resDate string) {
	resDate = adate
	r := regexp.MustCompile("^(\\+|-)*[0-9]*:(\\+|-)*[0-9]*:(\\+|-)*[0-9]*$")
	if r.MatchString(resDate) {
		resDate = fmt.Sprintf("// %s", resDate)
	}

	// If date is like "//"
	r = regexp.MustCompile("^(\\+|-)*[0-9]*/(\\+|-)*[0-9]*/(\\+|-)*[0-9]*$")
	if r.MatchString(resDate) {
		resDate = fmt.Sprintf("%s ::", resDate)
	}
	return
}

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var begindate, enddate, separator, ifmt, ofmt, tz string
	// var endTime, beginTime time.Time
	var vOption bool
	var interval time.Duration
	var tmpl string
	var err error
	var rangeDate bool = false

	flag.StringVar(&begindate, "b", "// ::", "Begin date")
	flag.StringVar(&enddate, "e", "", "End date")
	flag.StringVar(&separator, "s", " ", "Separator")
	flag.StringVar(&tz, "tz", "Local", "Timezone")
	flag.StringVar(&ifmt, "ifmt", "%YYYY/%MM/%DD %hh:%mm:%ss", "Input Format (%YYYY/%MM/%DD %hh:%mm:%ss)")
	flag.StringVar(&ofmt, "ofmt", "%YYYY/%MM/%DD %hh:%mm:%ss", "Input Format (%YYYY/%MM/%DD %hh:%mm:%ss), use @ts for timestamp")
	flag.DurationVar(&interval, "i", 0, "Interval (Ex: 1m or 1h or 15s)")
	flag.StringVar(&tmpl, "tmpl", "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} {{ .BeginTime.Unix }} {{ .EndTime.Unix }}", "Used only with -i option")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.Parse()

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
	// If date is like :: or //
	begindate = completeDate(begindate)
	enddate = completeDate(enddate)

	beginTime, err := calcdatelib.NewDate(begindate, ifmt, tz)
	if err != nil {
		fmt.Println("Format date begindate KO")
		os.Exit(1)
	}

	if rangeDate {
		beginTime, err = calcdatelib.NewDate(begindate, ifmt, tz)
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
			// calc range with interval
			tmpInterval, _ := calcdatelib.NewDate(begindate, ifmt, tz)    // no need to control err again
			tmpEndInterval, _ := calcdatelib.NewDate(begindate, ifmt, tz) // no need to control err again
			for tmpInterval.Before(endTime) {
				tmpEndInterval.Add(interval)
				l, err := calcdatelib.RenderTemplate(tmpl, tmpInterval.Time(), tmpEndInterval.Time())
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}
				res := strings.Split(l, "\n")
				for _, val := range res {
					fmt.Println(val)
				}
				tmpInterval.Add(interval)
			}
		}
	} else {
		fmt.Println(beginTime.Format(ofmt))
	}
}
