package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
)

func completeDate(adate string) (resDate string) {
	resDate = adate
	r := regexp.MustCompile(`^(\+|-)*[0-9]*:(\+|-)*[0-9]*:(\+|-)*[0-9]*$`)
	if r.MatchString(resDate) {
		resDate = fmt.Sprintf("// %s", resDate)
	}

	// If date is like "//"
	r = regexp.MustCompile(`^(\+|-)*[0-9]*/(\+|-)*[0-9]*/(\+|-)*[0-9]*$`)
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
	var begindate, enddate, separator, ifmt, ofmt string
	// var endTime, beginTime time.Time
	var vOption bool
	var interval time.Duration
	var tmpl string
	var err error
	var rangeDate bool = false

	flag.StringVar(&begindate, "b", "// ::", "Begin date")
	flag.StringVar(&enddate, "e", "", "End date")
	flag.StringVar(&separator, "s", " ", "Separator")
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
	// If date is like :: or //  TODO control the format
	begindate = completeDate(begindate)
	enddate = completeDate(enddate)

	beginTime, err := calcdatelib.NewDate(begindate, ifmt)
	if err != nil {
		fmt.Println("Format date begindate KO")
		os.Exit(1)
	}

	if rangeDate {
		beginTime, err = calcdatelib.NewDate(begindate, ifmt)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		endTime, err := calcdatelib.NewDate(enddate, ifmt)
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
	} else {
		fmt.Println(beginTime.Format(ofmt))
	}
}
