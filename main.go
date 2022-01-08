package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/sgaunet/calcdate/calcdate"
)

func completeDate(adate string) (resDate string) {
	resDate = adate
	r, _ := regexp.Compile("^(\\+|-)*[0-9]*:(\\+|-)*[0-9]*:(\\+|-)*[0-9]*$")
	if r.MatchString(resDate) {
		resDate = fmt.Sprintf("// %s", resDate)
	}

	// If date is like "//"
	r, _ = regexp.Compile("^(\\+|-)*[0-9]*/(\\+|-)*[0-9]*/(\\+|-)*[0-9]*$")
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
	var endtime, begintime time.Time
	var vOption bool
	var err error
	var rangeDate bool = false

	flag.StringVar(&begindate, "b", "// ::", "Begin date")
	flag.StringVar(&enddate, "e", "", "End date")
	flag.StringVar(&separator, "s", " ", "Separator")
	flag.StringVar(&tz, "tz", "Local", "Timezone")
	flag.StringVar(&ifmt, "ifmt", "YYYY/MM/DD hh:mm:ss", "Input Format (YYYY/MM/DD hh:mm:ss)")
	flag.StringVar(&ofmt, "ofmt", "YYYY/MM/DD hh:mm:ss", "Input Format (YYYY/MM/DD hh:mm:ss)")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if enddate != "" && begindate != "" {
		rangeDate = true
	}

	// If date is like :: or //
	//begindate = completeDate(begindate)
	//enddate = completeDate(enddate)

	begintime, err = calcdate.CreateDate(begindate, ifmt, tz, false, false)
	if err != nil {
		fmt.Println("Format date begindate KO")
		os.Exit(1)
	}

	if rangeDate {
		begintime, err = calcdate.CreateDate(begindate, ifmt, tz, true, false)
		endtime, err = calcdate.CreateDate(enddate, ifmt, tz, false, true)
		if err != nil {
			fmt.Println("Format date enddate KO")
			os.Exit(1)
		}

		fmt.Printf("%s%s%s\n", calcdate.ApplyFormat(ofmt, begintime), separator, calcdate.ApplyFormat(ofmt, endtime))
	} else {
		fmt.Println(calcdate.ApplyFormat(ofmt, begintime))
	}

}
