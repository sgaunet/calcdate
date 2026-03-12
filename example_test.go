package calcdate_test

import (
	"fmt"
	"time"

	"github.com/sgaunet/calcdate/v2"
)

// Fixed reference time for deterministic examples.
var refTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

func ExampleEvaluateExpression() {
	// EvaluateExpression is the simplest way to evaluate a date expression.
	// Note: output depends on current time; this example shows the API pattern.
	_, err := calcdate.EvaluateExpression("today | +1d", time.UTC)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("ok")
	// Output: ok
}

func ExampleNewExprParser() {
	// Use the parser and AST for full control over evaluation context.
	parser := calcdate.NewExprParser("today | +1d")
	node, err := parser.Parse("today | +1d")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ctx := &calcdate.EvalContext{
		Now:      refTime,
		Timezone: time.UTC,
	}
	result, err := node.Evaluate(ctx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-01-16
}

func ExampleNewExprParser_pipeline() {
	// Pipelines chain operations: start from today, add 1 month, then snap to end of month.
	parser := calcdate.NewExprParser("today | +1M | endOfMonth")
	node, err := parser.Parse("today | +1M | endOfMonth")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ctx := &calcdate.EvalContext{
		Now:      refTime,
		Timezone: time.UTC,
	}
	result, err := node.Evaluate(ctx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-02-29
}

func ExampleNewExprParser_isoDate() {
	// Parse an explicit ISO date and apply arithmetic.
	parser := calcdate.NewExprParser("2024-03-01 | +7d")
	node, err := parser.Parse("2024-03-01 | +7d")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ctx := &calcdate.EvalContext{
		Now:      refTime,
		Timezone: time.UTC,
	}
	result, err := node.Evaluate(ctx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-03-08
}

func ExampleEvaluateRangeExpression() {
	// Range expressions return start and end times.
	// Note: output depends on current time; this example shows the API pattern.
	_, _, err := calcdate.EvaluateRangeExpression("today...+7d", time.UTC)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("ok")
	// Output: ok
}

func ExampleEvaluateRange() {
	// EvaluateRange works with parsed AST nodes for deterministic results.
	parser := calcdate.NewExprParser("2024-01-15...2024-01-22")
	node, err := parser.Parse("2024-01-15...2024-01-22")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ctx := &calcdate.EvalContext{
		Now:      refTime,
		Timezone: time.UTC,
	}
	start, end, err := calcdate.EvaluateRange(node, ctx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	// Output: 2024-01-15 to 2024-01-22
}

func ExampleApplyOperation() {
	// ApplyOperation applies a single operation to a date.
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	// Add 2 weeks
	result, err := calcdate.ApplyOperation(base, "+", "2w", time.UTC)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-06-29
}

func ExampleApplyOperation_boundary() {
	// Boundary operations snap to period start/end.
	base := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	result, err := calcdate.ApplyOperation(base, "startofmonth", "", time.UTC)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-06-01
}

func ExampleGetOperationRegistry() {
	// List all available operations by category.
	registry := calcdate.GetOperationRegistry()
	fmt.Printf("Categories: %d\n", len(registry))
	fmt.Printf("First category: %s\n", registry[0].Name)
	fmt.Printf("First operation: %s\n", registry[0].Operations[0].Name)
	// Output:
	// Categories: 13
	// First category: Date Values
	// First operation: today
}

func ExampleNewDate() {
	// Legacy Date API: create a date from a formatted string.
	d, err := calcdate.NewDate("2024/01/15 10:30:00", "%YYYY/%MM/%DD %hh:%mm:%ss", "UTC")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(d.Format("%YYYY-%MM-%DD"))
	// Output: 2024-01-15
}

func ExampleDate_AddMonth() {
	d, err := calcdate.NewDate("2024/01/15 00:00:00", "%YYYY/%MM/%DD %hh:%mm:%ss", "UTC")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	d.AddMonth(3)
	fmt.Println(d.Format("%YYYY-%MM-%DD"))
	// Output: 2024-04-15
}

func ExampleParseDateValue() {
	ctx := &calcdate.EvalContext{
		Now:      refTime,
		Timezone: time.UTC,
	}

	// Parse "tomorrow" relative to the reference time.
	result, err := calcdate.ParseDateValue("tomorrow", ctx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result.Format("2006-01-02"))
	// Output: 2024-01-16
}
