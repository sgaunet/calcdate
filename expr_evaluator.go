package calcdate

import (
	"fmt"
	"time"
)

// Evaluate evaluates a DateNode.
func (n *DateNode) Evaluate(ctx *EvalContext) (time.Time, error) {
	return ParseDateValue(n.Value, ctx)
}

// Evaluate evaluates an OperationNode.
func (n *OperationNode) Evaluate(_ *EvalContext) (time.Time, error) {
	// Operations need a base date from context
	return time.Time{}, ErrOperationWithoutBaseDate
}

// Evaluate evaluates a PipeNode.
func (n *PipeNode) Evaluate(ctx *EvalContext) (time.Time, error) {
	// Evaluate base
	result, err := n.Base.Evaluate(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("base evaluation failed: %w", err)
	}
	
	// Apply each operation in sequence
	for _, op := range n.Operations {
		switch opNode := op.(type) {
		case *OperationNode:
			result, err = ApplyOperation(result, opNode.Op, opNode.Value, ctx.Timezone)
			if err != nil {
				return time.Time{}, err
			}
		default:
			return time.Time{}, fmt.Errorf("%w: %T", ErrInvalidPipelineOperation, op)
		}
	}
	
	return result, nil
}

// Evaluate evaluates a RangeNode.
func (n *RangeNode) Evaluate(_ *EvalContext) (time.Time, error) {
	// Range nodes don't evaluate to a single time, they need special handling
	return time.Time{}, ErrRangeNodesSeparateHandling
}

// Evaluate evaluates a VariableNode.
func (n *VariableNode) Evaluate(ctx *EvalContext) (time.Time, error) {
	if ctx.Variables == nil {
		return time.Time{}, fmt.Errorf("%w: %s", ErrVariableNotFound, n.Name)
	}
	
	if t, ok := ctx.Variables[n.Name]; ok {
		return t, nil
	}
	
	return time.Time{}, fmt.Errorf("%w: %s", ErrVariableNotFound, n.Name)
}

// Evaluate evaluates a TransformNode (used for iterations).
func (n *TransformNode) Evaluate(_ *EvalContext) (time.Time, error) {
	// Transform nodes don't evaluate to a single time, they need special handling
	return time.Time{}, ErrTransformNodesSeparate
}

// EvaluateRange evaluates a range expression and returns start and end times.
func EvaluateRange(node ExprNode, ctx *EvalContext) (time.Time, time.Time, error) {
	rangeNode, ok := node.(*RangeNode)
	if !ok {
		return time.Time{}, time.Time{}, ErrNotRangeExpression
	}
	
	start, err := rangeNode.Start.Evaluate(ctx)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to evaluate range start: %w", err)
	}
	
	end, err := rangeNode.End.Evaluate(ctx)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to evaluate range end: %w", err)
	}
	
	return start, end, nil
}

// EvaluateTransform evaluates a transform expression for iterations
//nolint:lll // long function signature is readable
func EvaluateTransform(transform *TransformNode, beginTime, endTime time.Time, index int, ctx *EvalContext) (time.Time, time.Time, error) {
	// Create context with variables
	transformCtx := &EvalContext{
		Now:      ctx.Now,
		Timezone: ctx.Timezone,
		Variables: map[string]time.Time{
			"$begin": beginTime,
			"$end":   endTime,
		},
		Index: index,
	}
	
	// Evaluate begin expression
	newBegin, err := transform.BeginExpr.Evaluate(transformCtx)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to evaluate begin transform: %w", err)
	}
	
	// Evaluate end expression
	newEnd, err := transform.EndExpr.Evaluate(transformCtx)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to evaluate end transform: %w", err)
	}
	
	return newBegin, newEnd, nil
}

// EvaluateExpression is the main entry point for evaluating expressions.
func EvaluateExpression(input string, tz *time.Location) (time.Time, error) {
	parser := NewExprParser(input)
	node, err := parser.Parse(input)
	if err != nil {
		return time.Time{}, err
	}
	
	ctx := &EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}
	
	t, err := node.Evaluate(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("expression evaluation failed: %w", err)
	}
	return t, nil
}

// EvaluateRangeExpression evaluates a range expression.
func EvaluateRangeExpression(input string, tz *time.Location) (time.Time, time.Time, error) {
	parser := NewExprParser(input)
	node, err := parser.Parse(input)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	
	ctx := &EvalContext{
		Now:      time.Now(),
		Timezone: tz,
	}
	
	return EvaluateRange(node, ctx)
}