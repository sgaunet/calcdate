// Package main provides a command-line utility for date calculations and manipulations
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
	var expr, each, transform, format string
	// var endTime, beginTime time.Time
	var vOption, listTZ, skipWeekends bool
	var interval time.Duration
	var tmpl string
	var err error
	var rangeDate = false

	// Legacy flags
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
	
	// New expression flags
	flag.StringVar(&expr, "expr", "", "Date expression (e.g., 'today +1d', 'now | +2h | round hour')")
	flag.StringVar(&expr, "x", "", "Date expression (short form)")
	flag.StringVar(&each, "each", "", "Iteration interval for ranges (e.g., '1d', '1w', '1M')")
	flag.StringVar(&transform, "transform", "", "Transform expression for iterations (e.g., '$begin +8h, $end +20h')")
	flag.StringVar(&transform, "t", "", "Transform expression (short form)")
	flag.StringVar(&format, "format", "", "Output format: iso, sql, ts, human, compact, or custom Go format")
	flag.StringVar(&format, "f", "", "Output format (short form)")
	flag.BoolVar(&skipWeekends, "skip-weekends", false, "Skip weekend days in iterations")
	
	flag.Parse()

	if listTZ {
		calcdatelib.ListTZ()
		os.Exit(0)
	}

	if vOption {
		printVersion()
		os.Exit(0)
	}

	// Handle new expression syntax
	if expr != "" {
		processExpressionMode(expr, each, transform, format, tz, skipWeekends)
		return
	}

	// Legacy mode
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
		processRangeDateCase(begindate, enddate, ifmt, tz, interval, ofmt, separator, tmpl, format)
	} else {
		// Use format flag if provided, otherwise use ofmt
		if format != "" {
			// Get the timezone location
			loc, _ := time.LoadLocation(tz)
			output := formatOutput(beginTime.Time(), format, loc)
			fmt.Println(output)
		} else {
			fmt.Println(beginTime.Format(ofmt))
		}
	}
}

// processExpressionMode handles the new expression syntax
func processExpressionMode(expr, each, transform, format, tzStr string, skipWeekends bool) {
	// Parse timezone
	var tz *time.Location
	var err error
	if tzStr != "" {
		tz, err = time.LoadLocation(tzStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid timezone: %v\n", err)
			os.Exit(1)
		}
	} else {
		tz = time.Local
	}

	// Parse the expression
	parser := calcdatelib.NewExprParser(expr)
	node, err := parser.Parse(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse expression: %v\n", err)
		os.Exit(1)
	}

	// Check if it's a range expression (directly or within a pipe)
	if rangeNode, ok := node.(*calcdatelib.RangeNode); ok {
		processRangeExpression(rangeNode, each, transform, format, tz, skipWeekends)
		return
	}
	
	// Check if it's a pipe with a range as base
	if pipeNode, ok := node.(*calcdatelib.PipeNode); ok {
		if rangeNode, ok := pipeNode.Base.(*calcdatelib.RangeNode); ok {
			// This is a range with pipeline operations
			processRangeWithPipeline(rangeNode, pipeNode.Operations, each, transform, format, tz, skipWeekends)
			return
		}
	}

	// Single date expression
	ctx := &calcdatelib.EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}

	result, err := node.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate expression: %v\n", err)
		os.Exit(1)
	}

	// Format and print the result
	output := formatOutput(result, format, tz)
	fmt.Println(output)
}

// processRangeWithPipeline handles range expressions with pipeline operations
func processRangeWithPipeline(rangeNode *calcdatelib.RangeNode, operations []calcdatelib.ExprNode, each, transform, format string, tz *time.Location, skipWeekends bool) {
	ctx := &calcdatelib.EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}

	// Evaluate start and end
	start, err := rangeNode.Start.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate range start: %v\n", err)
		os.Exit(1)
	}

	end, err := rangeNode.End.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate range end: %v\n", err)
		os.Exit(1)
	}

	// Apply pipeline operations to the end date
	for _, op := range operations {
		switch opNode := op.(type) {
		case *calcdatelib.OperationNode:
			end, err = calcdatelib.ApplyOperation(end, opNode.Op, opNode.Value, tz)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to apply operation: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Continue with the rest of the range processing
	processRangeExpressionInternal(start, end, each, transform, format, tz, skipWeekends)
}

// processRangeExpression handles range expressions with optional iterations
func processRangeExpression(rangeNode *calcdatelib.RangeNode, each, transform, format string, tz *time.Location, skipWeekends bool) {
	ctx := &calcdatelib.EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}

	// Evaluate start and end
	start, err := rangeNode.Start.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate range start: %v\n", err)
		os.Exit(1)
	}

	end, err := rangeNode.End.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate range end: %v\n", err)
		os.Exit(1)
	}

	// Continue with common processing
	processRangeExpressionInternal(start, end, each, transform, format, tz, skipWeekends)
}

// processRangeExpressionInternal handles the common logic for range processing
func processRangeExpressionInternal(start, end time.Time, each, transform, format string, tz *time.Location, skipWeekends bool) {
	ctx := &calcdatelib.EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}

	// Parse transform if provided
	var transformNode *calcdatelib.TransformNode
	var err error
	if transform != "" {
		parser := calcdatelib.NewExprParser("")
		transformNode, err = parser.ParseTransform(transform)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse transform: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle iterations
	if each != "" {
		// Check if it's a special interval (M, Y, q)
		if strings.HasSuffix(each, "M") || strings.HasSuffix(each, "Y") || strings.HasSuffix(each, "q") {
			results, err := calcdatelib.IterateWithSpecialInterval(start, end, each, transformNode, tz)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to iterate: %v\n", err)
				os.Exit(1)
			}
			for _, result := range results {
				if skipWeekends && isWeekend(result.BeginTime) {
					continue
				}
				printIterationResult(result, format, tz)
			}
		} else {
			// Parse regular interval
			interval, err := calcdatelib.ParseInterval(each)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse interval: %v\n", err)
				os.Exit(1)
			}

			iterator := calcdatelib.NewRangeIterator(start, end, interval, transformNode, tz)
			results, err := iterator.Iterate()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to iterate: %v\n", err)
				os.Exit(1)
			}

			for _, result := range results {
				if skipWeekends && isWeekend(result.BeginTime) {
					continue
				}
				printIterationResult(result, format, tz)
			}
		}
	} else {
		// No iteration, just print the range
		if transformNode != nil {
			start, end, err = calcdatelib.EvaluateTransform(transformNode, start, end, 0, ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to apply transform: %v\n", err)
				os.Exit(1)
			}
		}
		startStr := formatOutput(start, format, tz)
		endStr := formatOutput(end, format, tz)
		fmt.Printf("%s - %s\n", startStr, endStr)
	}
}

// formatOutput formats a time according to the specified format
func formatOutput(t time.Time, format string, tz *time.Location) string {
	if tz != nil {
		t = t.In(tz)
	}

	switch format {
	case "iso":
		return t.Format(time.RFC3339)
	case "sql":
		return t.Format("2006-01-02 15:04:05")
	case "ts":
		return fmt.Sprintf("%d", t.Unix())
	case "human":
		return t.Format("Monday, January 2, 2006")
	case "compact":
		return t.Format("20060102")
	case "":
		// Default format
		return t.Format("2006/01/02 15:04:05")
	default:
		// Custom format
		return t.Format(format)
	}
}

// printIterationResult prints a single iteration result
func printIterationResult(result calcdatelib.IterationResult, format string, tz *time.Location) {
	beginStr := formatOutput(result.BeginTime, format, tz)
	endStr := formatOutput(result.EndTime, format, tz)
	fmt.Printf("%s - %s\n", beginStr, endStr)
}

// isWeekend checks if a date is a weekend
func isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// processRangeDateCase handles the logic for date range processing
// processRangeDateCase handles date ranges with optional interval processing.
func processRangeDateCase(
	begindate, enddate, ifmt, tz string,
	interval time.Duration,
	ofmt, separator, tmpl, format string,
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
		if format != "" {
			// Use format flag if provided
			loc, _ := time.LoadLocation(tz)
			beginStr := formatOutput(beginTime.Time(), format, loc)
			endStr := formatOutput(endTime.Time(), format, loc)
			fmt.Printf("%s%s%s\n", beginStr, separator, endStr)
		} else {
			fmt.Printf("%s%s%s\n", beginTime.Format(ofmt), separator, endTime.Format(ofmt))
		}
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
