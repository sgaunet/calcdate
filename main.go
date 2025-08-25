// Package main provides a command-line utility for date calculations and manipulations
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sgaunet/calcdate/v2/calcdatelib"
)


var version = "development"

// errNoExpressionProvided is returned when no expression is provided via stdin.
var errNoExpressionProvided = errors.New("no expression provided via stdin")

func printVersion() {
	fmt.Println(version)
}

// isStdinRedirected checks if stdin is redirected (piped or from file).
// Returns true if stdin is redirected, false if it's a terminal.
func isStdinRedirected() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a character device (terminal)
	// If it's NOT a character device, it's redirected
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// readExprFromStdin reads the date expression from stdin.
// It reads the first non-empty line and returns it as the expression.
func readExprFromStdin() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}
	return "", errNoExpressionProvided
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

	// Handle expression syntax (required via parameter or stdin)
	if config.expr == "" {
		// Check if stdin is redirected (piped or from file)
		if isStdinRedirected() {
			// Input is piped or from file, read from stdin
			var err error
			config.expr, err = readExprFromStdin()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		} else {
			// No input redirection and no expression provided
			// Use default behavior: show current date and time
			config.expr = "now"
		}
	}

	processExpressionMode(config.expr, config.each, config.transform, config.format, config.tz, config.skipWeekends)
}

type cliConfig struct {
	tz                                            string
	expr, each, transform, format                 string
	vOption, listTZ, skipWeekends                 bool
}

func parseCommandLineFlags() cliConfig {
	var config cliConfig

	// Legacy flags (kept)
	flag.StringVar(&config.tz, "tz", "Local", "Input timezone")
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
	flag.StringVar(&config.format, "format", "",
		"Output format: iso, sql, ts, human, compact, or Unix date format "+
		"(e.g., '%Y-%m-%d %H:%M:%S', '%Y-%m-%d %H:%M:%S %Z')")
	flag.StringVar(&config.format, "f", "", "Output format (short form)")
	flag.BoolVar(&config.skipWeekends, "skip-weekends", false, "Skip weekend days in iterations")

	flag.Parse()
	return config
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
		// Check if this is a Unix date format (contains %)
		if strings.Contains(format, "%") {
			// Convert Unix format to Go format
			goFormat := calcdatelib.ConvertUnixFormatToGolang(format)
			return t.Format(goFormat)
		}
		// Otherwise treat as Go format (backward compatibility)
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

