// Package main provides a command-line utility for date calculations and manipulations
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sgaunet/calcdate/v2/calcdatelib"
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
	config := parseCommandLineFlags()

	if config.listTZ {
		calcdatelib.ListTZ()
		os.Exit(0)
	}

	if config.vOption {
		printVersion()
		os.Exit(0)
	}

	// Handle new expression syntax
	if config.expr != "" {
		processExpressionMode(config.expr, config.each, config.transform, config.format, config.tz, config.skipWeekends)
		return
	}

	processLegacyMode(config)
}

type cliConfig struct {
	begindate, enddate, separator, ifmt, ofmt, tz string
	expr, each, transform, format                 string
	vOption, listTZ, skipWeekends                 bool
	interval                                      time.Duration
	tmpl                                          string
}

func parseCommandLineFlags() cliConfig {
	var config cliConfig

	// Legacy flags
	flag.StringVar(&config.begindate, "b", "// ::", "Begin date")
	flag.StringVar(&config.enddate, "e", "", "End date")
	flag.StringVar(&config.separator, "s", " ", "Separator")
	flag.StringVar(&config.tz, "tz", "Local", "Input timezone")
	flag.StringVar(&config.ifmt, "ifmt", "%YYYY/%MM/%DD %hh:%mm:%ss", "Input Format (%YYYY/%MM/%DD %hh:%mm:%ss)")
	// Define output format with timestamp and timezone support
	flag.StringVar(&config.ofmt, "ofmt", "%YYYY/%MM/%DD %hh:%mm:%ss",
		"Input Format (%YYYY/%MM/%DD %hh:%mm:%ss), use @ts for timestamp %z for offset %Z for timezone")
	flag.DurationVar(&config.interval, "i", 0, "Interval (Ex: 1m or 1h or 15s)")
	// Define default template for interval rendering
	defaultTemplate := "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - "
	defaultTemplate += "{{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} "
	defaultTemplate += "{{ .BeginTime.Unix }} {{ .EndTime.Unix }}"
	flag.StringVar(&config.tmpl, "tmpl", defaultTemplate, "Used only with -i option")
	flag.BoolVar(&config.listTZ, "list-tz", false, "List timezones")
	flag.BoolVar(&config.vOption, "v", false, "Get version")

	// New expression flags
	flag.StringVar(&config.expr, "expr", "",
		"Date expression (e.g., 'today +1d', 'now | +2h | round hour', 'today...+7d')")
	flag.StringVar(&config.expr, "x", "", "Date expression (short form)")
	flag.StringVar(&config.each, "each", "", "Iteration interval for ranges (e.g., '1d', '1w', '1M')")
	flag.StringVar(&config.transform, "transform", "",
		"Transform expression for iterations (e.g., '$begin +8h, $end +20h')")
	flag.StringVar(&config.transform, "t", "", "Transform expression (short form)")
	flag.StringVar(&config.format, "format", "", "Output format: iso, sql, ts, human, compact, or custom Go format")
	flag.StringVar(&config.format, "f", "", "Output format (short form)")
	flag.BoolVar(&config.skipWeekends, "skip-weekends", false, "Skip weekend days in iterations")

	flag.Parse()
	return config
}

func processLegacyMode(config cliConfig) {
	rangeDate := config.enddate != "" && config.begindate != ""

	// -i option can be used only with two dates (begin/end)
	if config.interval != 0 && !rangeDate {
		fmt.Println("specify a range date")
		os.Exit(1)
	}

	config.begindate = completeDate(config.begindate)
	config.enddate = completeDate(config.enddate)

	beginTime, err := calcdatelib.NewDate(config.begindate, config.ifmt, config.tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format date begindate KO: %v\n", err)
		os.Exit(1)
	}

	if rangeDate {
		processRangeDateCase(config.begindate, config.enddate, config.ifmt, config.tz,
			config.interval, config.ofmt, config.separator, config.tmpl, config.format)
	} else {
		printSingleDate(beginTime, config.format, config.ofmt, config.tz)
	}
}

func printSingleDate(beginTime *calcdatelib.Date, format, ofmt, tz string) {
	if format != "" {
		loc, _ := time.LoadLocation(tz)
		output := formatOutput(beginTime.Time(), format, loc)
		fmt.Println(output)
	} else {
		fmt.Println(beginTime.Format(ofmt))
	}
}

// processExpressionMode handles the new expression syntax.
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
		tz = time.Local //nolint:gosmopolitan // intentional default to local timezone
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
//
//nolint:lll // long function signature is readable
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
		if opNode, ok := op.(*calcdatelib.OperationNode); ok {
			end, err = calcdatelib.ApplyOperation(end, opNode.Op, opNode.Value, tz)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to apply operation: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown operation type\n")
			os.Exit(1)
		}
	}

	// Continue with the rest of the range processing
	processRangeExpressionInternal(start, end, each, transform, format, tz, skipWeekends)
}

// processRangeExpression handles range expressions with optional iterations
//
//nolint:lll // long function signature is readable
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
//
//nolint:lll // long function signature is readable
func processRangeExpressionInternal(start, end time.Time, each, transform, format string, tz *time.Location, skipWeekends bool) {
	transformNode, err := parseTransformIfProvided(transform)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse transform: %v\n", err)
		os.Exit(1)
	}

	if each != "" {
		processIterations(start, end, each, transformNode, format, tz, skipWeekends)
	} else {
		processSingleRange(start, end, transformNode, format, tz)
	}
}

func parseTransformIfProvided(transform string) (*calcdatelib.TransformNode, error) {
	if transform == "" {
		return nil, nil //nolint:nilnil // returning nil transform and nil error is correct for empty input
	}

	parser := calcdatelib.NewExprParser("")
	node, err := parser.ParseTransform(transform)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transform: %w", err)
	}
	return node, nil
}

//nolint:lll // long function signature is readable
func processIterations(start, end time.Time, each string, transformNode *calcdatelib.TransformNode, format string, tz *time.Location, skipWeekends bool) {
	if isSpecialInterval(each) {
		processSpecialIntervalIterations(start, end, each, transformNode, format, tz, skipWeekends)
	} else {
		processRegularIntervalIterations(start, end, each, transformNode, format, tz, skipWeekends)
	}
}

func isSpecialInterval(each string) bool {
	return strings.HasSuffix(each, "M") || strings.HasSuffix(each, "Y") || strings.HasSuffix(each, "q")
}

//nolint:lll // long function signature is readable
func processSpecialIntervalIterations(start, end time.Time, each string, transformNode *calcdatelib.TransformNode, format string, tz *time.Location, skipWeekends bool) {
	results, err := calcdatelib.IterateWithSpecialInterval(start, end, each, transformNode, tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to iterate: %v\n", err)
		os.Exit(1)
	}
	printFilteredResults(results, format, tz, skipWeekends)
}

//nolint:lll // long function signature is readable
func processRegularIntervalIterations(start, end time.Time, each string, transformNode *calcdatelib.TransformNode, format string, tz *time.Location, skipWeekends bool) {
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
	printFilteredResults(results, format, tz, skipWeekends)
}

func printFilteredResults(results []calcdatelib.IterationResult, format string, tz *time.Location, skipWeekends bool) {
	for _, result := range results {
		if skipWeekends && isWeekend(result.BeginTime) {
			continue
		}
		printIterationResult(result, format, tz)
	}
}

//nolint:lll // long function signature is readable
func processSingleRange(start, end time.Time, transformNode *calcdatelib.TransformNode, format string, tz *time.Location) {
	if transformNode != nil {
		ctx := &calcdatelib.EvalContext{
			Now:      time.Now(),
			Timezone: tz,
		}
		var err error
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

// formatOutput formats a time according to the specified format.
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
		return strconv.FormatInt(t.Unix(), 10)
	case "human":
		return t.Format("Monday, January 2, 2006")
	case "compact":
		return t.Format("20060102")
	case "":
		// Default format (sql)
		return t.Format("2006-01-02 15:04:05")
	default:
		// Custom format
		return t.Format(format)
	}
}

// printIterationResult prints a single iteration result.
func printIterationResult(result calcdatelib.IterationResult, format string, tz *time.Location) {
	beginStr := formatOutput(result.BeginTime, format, tz)
	endStr := formatOutput(result.EndTime, format, tz)
	fmt.Printf("%s - %s\n", beginStr, endStr)
}

// isWeekend checks if a date is a weekend.
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
	beginTime, endTime := parseBeginEndDates(begindate, enddate, ifmt, tz)

	if interval == 0 {
		printDateRange(beginTime, endTime, format, ofmt, separator, tz)
	} else {
		printIntervalLines(beginTime, endTime, interval, tmpl)
	}
}

func parseBeginEndDates(begindate, enddate, ifmt, tz string) (*calcdatelib.Date, *calcdatelib.Date) {
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
	return beginTime, endTime
}

func printDateRange(beginTime, endTime *calcdatelib.Date, format, ofmt, separator, tz string) {
	if format != "" {
		loc, _ := time.LoadLocation(tz)
		beginStr := formatOutput(beginTime.Time(), format, loc)
		endStr := formatOutput(endTime.Time(), format, loc)
		fmt.Printf("%s%s%s\n", beginStr, separator, endStr)
	} else {
		fmt.Printf("%s%s%s\n", beginTime.Format(ofmt), separator, endTime.Format(ofmt))
	}
}

func printIntervalLines(beginTime, endTime *calcdatelib.Date, interval time.Duration, tmpl string) {
	intervals, err := calcdatelib.RenderIntervalLines(*beginTime, *endTime, interval, tmpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	for idx := range intervals {
		fmt.Println(intervals[idx])
	}
}
