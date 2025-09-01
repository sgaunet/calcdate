package calcdate

import (
	"time"
)

// ExprNode represents a node in the expression AST.
type ExprNode interface {
	Evaluate(ctx *EvalContext) (time.Time, error)
}

// EvalContext provides context for expression evaluation.
type EvalContext struct {
	Now      time.Time
	Timezone *time.Location
	Variables map[string]time.Time
	Index    int
}

// DateNode represents a date/time value.
type DateNode struct {
	Value string // "today", "now", "yesterday", ISO date, etc.
}

// OperationNode represents an operation on a date.
type OperationNode struct {
	Op    string // "+", "-", "startOf", "endOf", "round", "trunc", "day", "time"
	Value string // "1d", "2w", "month", "15", "14:30:00", etc.
}

// PipeNode represents a pipeline of operations.
type PipeNode struct {
	Base       ExprNode
	Operations []ExprNode
}

// RangeNode represents a date range.
type RangeNode struct {
	Start ExprNode
	End   ExprNode
}

// VariableNode represents a variable reference.
type VariableNode struct {
	Name string // "$begin", "$end", "$index"
}

// TransformNode represents a transformation expression for iterations.
type TransformNode struct {
	BeginExpr ExprNode
	EndExpr   ExprNode
}