// Package calcdate provides date calculation and manipulation functionality
// with an intuitive expression-based syntax.
//
// # Expression Engine
//
// The primary API is the expression evaluator, which supports a rich syntax
// for date calculations:
//
//   - Natural language dates: today, now, yesterday, tomorrow
//   - ISO 8601 dates: 2024-01-15, 2024-01-15T14:30:00
//   - Arithmetic: +1d, -2w, +3M, -1Y
//   - Pipeline operations: today | +1M | endOfMonth
//   - Range expressions: today...+7d
//   - Boundary operations: startOfMonth, endOfWeek, etc.
//
// Basic usage:
//
//	result, err := calcdate.EvaluateExpression("today | +1M | endOfMonth", time.UTC)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Format("2006-01-02"))
//
// # AST-Based Evaluation
//
// For more control, use the parser and AST directly:
//
//	parser := calcdate.NewExprParser(input)
//	node, err := parser.Parse("today | +1d")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := &calcdate.EvalContext{
//	    Now:      time.Now(),
//	    Timezone: time.UTC,
//	}
//	result, err := node.Evaluate(ctx)
//
// # Range Expressions
//
// Range expressions generate start/end pairs:
//
//	start, end, err := calcdate.EvaluateRangeExpression("today...+7d", time.UTC)
//
// # Legacy Date API
//
// The [Date] type provides lower-level date manipulation with format strings
// and relative time calculations. For new code, prefer the expression engine.
//
// # Operations
//
// Use [GetOperationRegistry] to discover all available operations at runtime,
// or use [ApplyOperation] to apply operations programmatically.
package calcdate
