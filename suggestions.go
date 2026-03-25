package calcdate

import (
	"fmt"
	"sort"
	"strings"
)

// Suggestion engine constants.
const (
	maxSuggestions          = 3
	shortInputLen          = 3
	mediumInputLen         = 6
	shortInputMaxDistance   = 1
	mediumInputMaxDistance  = 2
	longInputMaxDistance    = 3
)

// ExpressionError is a custom error type that provides suggestions and context
// for invalid expressions.
type ExpressionError struct {
	Wrapped      error
	Message      string
	Suggestions  []string
	ValidOptions []string
	Position     int
	Expression   string
}

func (e *ExpressionError) Error() string {
	var b strings.Builder
	b.WriteString(e.Message)

	if len(e.Suggestions) > 0 {
		b.WriteString(". Did you mean: ")
		b.WriteString(strings.Join(e.Suggestions, ", "))
		b.WriteString("?")
	}

	if len(e.ValidOptions) > 0 {
		b.WriteString(". Valid options: ")
		b.WriteString(strings.Join(e.ValidOptions, ", "))
	}

	if e.Position >= 0 && e.Expression != "" {
		b.WriteString("\n")
		b.WriteString(formatPositionHighlight(e.Expression, e.Position))
	}

	return b.String()
}

// Unwrap returns the wrapped sentinel error for errors.Is() compatibility.
func (e *ExpressionError) Unwrap() error {
	return e.Wrapped
}

// levenshteinDistance computes the Levenshtein distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

type suggestion struct {
	name     string
	distance int
}

// findSuggestions returns up to maxSuggestions closest matches within the
// threshold derived from input length.
func findSuggestions(input string, candidates []string) []string {
	maxDistance := autoThreshold(len(input))

	var matches []suggestion
	lower := strings.ToLower(input)

	for _, c := range candidates {
		d := levenshteinDistance(lower, strings.ToLower(c))
		if d > 0 && d <= maxDistance {
			matches = append(matches, suggestion{c, d})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].distance < matches[j].distance
	})

	result := make([]string, 0, maxSuggestions)
	for i := 0; i < len(matches) && i < maxSuggestions; i++ {
		result = append(result, matches[i].name)
	}

	return result
}

func autoThreshold(inputLen int) int {
	switch {
	case inputLen <= shortInputLen:
		return shortInputMaxDistance
	case inputLen <= mediumInputLen:
		return mediumInputMaxDistance
	default:
		return longInputMaxDistance
	}
}

// validOperationNames returns all recognized operation keywords.
func validOperationNames() []string {
	return []string{
		"start", "startofday", "end", "endofday",
		"startofweek", "endofweek",
		"startofmonth", "endofmonth",
		"startofyear", "endofyear",
		"startofquarter", "endofquarter",
		"startofhour", "endofhour",
		"startofminute", "endofminute",
		"startofsecond", "endofsecond",
		"day", "time", "round", "trunc",
	}
}

// validDateKeywords returns all recognized date keywords.
func validDateKeywords() []string {
	return []string{
		"today", "now", "yesterday", "tomorrow",
		"monday", "tuesday", "wednesday", "thursday",
		"friday", "saturday", "sunday",
	}
}

// validUnits returns all recognized time units with descriptions.
func validUnits() []string {
	return []string{
		"s (seconds)", "m (minutes)", "h (hours)", "d (days)",
		"w (weeks)", "M (months)", "Y (years)", "q (quarters)",
	}
}

// validUnitLetters returns just the unit letters for suggestion matching.
func validUnitLetters() []string {
	return []string{"s", "m", "h", "d", "w", "M", "Y", "q"}
}

func newUnknownOperationError(op string, pos int, expr string) *ExpressionError {
	suggestions := findSuggestions(op, validOperationNames())

	return &ExpressionError{
		Wrapped:     ErrUnknownOperation,
		Message:     fmt.Sprintf("%s: %s", ErrUnknownOperation.Error(), op),
		Suggestions: suggestions,
		Position:    pos,
		Expression:  expr,
	}
}

func newUnknownUnitError(unit string) *ExpressionError {
	suggestions := findSuggestions(unit, validUnitLetters())

	return &ExpressionError{
		Wrapped:      ErrUnknownUnit,
		Message:      fmt.Sprintf("%s: %s", ErrUnknownUnit.Error(), unit),
		Suggestions:  suggestions,
		ValidOptions: validUnits(),
		Position:     -1,
	}
}

func newInvalidDateValueError(value string) *ExpressionError {
	suggestions := findSuggestions(value, validDateKeywords())

	return &ExpressionError{
		Wrapped:     ErrInvalidDateValue,
		Message:     fmt.Sprintf("%s: %s", ErrInvalidDateValue.Error(), value),
		Suggestions: suggestions,
		Position:    -1,
	}
}

// formatPositionHighlight generates a caret marker under the problematic position.
func formatPositionHighlight(expr string, pos int) string {
	if pos < 0 || pos > len(expr) {
		return ""
	}

	var b strings.Builder
	b.WriteString("  ")
	b.WriteString(expr)
	b.WriteString("\n  ")
	b.WriteString(strings.Repeat(" ", pos))
	b.WriteString("^")

	return b.String()
}
