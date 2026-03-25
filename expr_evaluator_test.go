package calcdate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariableNodeEvaluate(t *testing.T) {
	beginTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("valid variable", func(t *testing.T) {
		ctx := &EvalContext{
			Variables: map[string]time.Time{"$begin": beginTime},
		}
		node := &VariableNode{Name: "$begin"}
		result, err := node.Evaluate(ctx)
		require.NoError(t, err)
		assert.True(t, beginTime.Equal(result))
	})

	t.Run("missing variable", func(t *testing.T) {
		ctx := &EvalContext{
			Variables: map[string]time.Time{"$begin": beginTime},
		}
		node := &VariableNode{Name: "$missing"}
		_, err := node.Evaluate(ctx)
		assert.ErrorIs(t, err, ErrVariableNotFound)
	})

	t.Run("nil variables map", func(t *testing.T) {
		ctx := &EvalContext{}
		node := &VariableNode{Name: "$begin"}
		_, err := node.Evaluate(ctx)
		assert.ErrorIs(t, err, ErrVariableNotFound)
	})
}

func TestOperationNodeEvaluate(t *testing.T) {
	node := &OperationNode{Op: "+", Value: "1d"}
	_, err := node.Evaluate(&EvalContext{})
	assert.ErrorIs(t, err, ErrOperationWithoutBaseDate)
}

func TestRangeNodeEvaluate(t *testing.T) {
	node := &RangeNode{
		Start: &DateNode{Value: "today"},
		End:   &DateNode{Value: "tomorrow"},
	}
	_, err := node.Evaluate(&EvalContext{Now: time.Now(), Timezone: time.UTC})
	assert.ErrorIs(t, err, ErrRangeNodesSeparateHandling)
}

func TestTransformNodeEvaluate(t *testing.T) {
	node := &TransformNode{
		BeginExpr: &DateNode{Value: "today"},
		EndExpr:   &DateNode{Value: "tomorrow"},
	}
	_, err := node.Evaluate(&EvalContext{})
	assert.ErrorIs(t, err, ErrTransformNodesSeparate)
}

func TestEvaluateRange(t *testing.T) {
	ctx := &EvalContext{Now: time.Now(), Timezone: time.UTC}

	t.Run("valid range", func(t *testing.T) {
		node := &RangeNode{
			Start: &DateNode{Value: "2024-01-01"},
			End:   &DateNode{Value: "2024-12-31"},
		}
		start, end, err := EvaluateRange(node, ctx)
		require.NoError(t, err)
		assert.Equal(t, 2024, start.Year())
		assert.Equal(t, time.January, start.Month())
		assert.Equal(t, 1, start.Day())
		assert.Equal(t, time.December, end.Month())
		assert.Equal(t, 31, end.Day())
	})

	t.Run("not a range node", func(t *testing.T) {
		node := &DateNode{Value: "today"}
		_, _, err := EvaluateRange(node, ctx)
		assert.ErrorIs(t, err, ErrNotRangeExpression)
	})

	t.Run("invalid start", func(t *testing.T) {
		node := &RangeNode{
			Start: &DateNode{Value: "invaliddate"},
			End:   &DateNode{Value: "2024-01-01"},
		}
		_, _, err := EvaluateRange(node, ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "range start")
	})

	t.Run("invalid end", func(t *testing.T) {
		node := &RangeNode{
			Start: &DateNode{Value: "2024-01-01"},
			End:   &DateNode{Value: "invaliddate"},
		}
		_, _, err := EvaluateRange(node, ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "range end")
	})
}

func TestEvaluateTransform(t *testing.T) {
	beginTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	ctx := &EvalContext{Now: time.Now(), Timezone: time.UTC}

	t.Run("shift begin and end", func(t *testing.T) {
		parser := NewExprParser("")
		transform, err := parser.ParseTransform("$begin +8h, $end -1h")
		require.NoError(t, err)

		newBegin, newEnd, err := EvaluateTransform(transform, beginTime, endTime, 0, ctx)
		require.NoError(t, err)
		assert.Equal(t, 8, newBegin.Hour())  // 00:00 + 8h
		assert.Equal(t, 23, newEnd.Hour())   // 00:00 (next day) - 1h
	})

	t.Run("error in begin expression", func(t *testing.T) {
		transform := &TransformNode{
			BeginExpr: &VariableNode{Name: "$nonexistent"},
			EndExpr:   &VariableNode{Name: "$end"},
		}
		_, _, err := EvaluateTransform(transform, beginTime, endTime, 0, ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "begin transform")
	})

	t.Run("error in end expression", func(t *testing.T) {
		transform := &TransformNode{
			BeginExpr: &VariableNode{Name: "$begin"},
			EndExpr:   &VariableNode{Name: "$nonexistent"},
		}
		_, _, err := EvaluateTransform(transform, beginTime, endTime, 0, ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end transform")
	})
}

func TestEvaluateRangeExpression(t *testing.T) {
	t.Run("fixed dates", func(t *testing.T) {
		start, end, err := EvaluateRangeExpression("2024-01-01...2024-12-31", time.UTC)
		require.NoError(t, err)
		assert.Equal(t, time.January, start.Month())
		assert.Equal(t, 1, start.Day())
		assert.Equal(t, time.December, end.Month())
		assert.Equal(t, 31, end.Day())
	})

	t.Run("parse error", func(t *testing.T) {
		_, _, err := EvaluateRangeExpression("|||", time.UTC)
		assert.Error(t, err)
	})

	t.Run("not a range expression", func(t *testing.T) {
		_, _, err := EvaluateRangeExpression("today", time.UTC)
		assert.Error(t, err)
	})
}
