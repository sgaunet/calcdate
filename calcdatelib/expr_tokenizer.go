package calcdatelib

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenDate
	TokenOperator
	TokenUnit
	TokenPipe
	TokenRange
	TokenVariable
	TokenKeyword
	TokenNumber
	TokenTime
	TokenComma
	TokenLParen
	TokenRParen
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Tokenizer tokenizes date expressions
type Tokenizer struct {
	input string
	pos   int
	tokens []Token
}

// NewTokenizer creates a new tokenizer
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: input,
		pos:   0,
		tokens: []Token{},
	}
}

// Tokenize tokenizes the input string
func (t *Tokenizer) Tokenize() ([]Token, error) {
	for t.pos < len(t.input) {
		if err := t.nextToken(); err != nil {
			return nil, err
		}
	}
	t.tokens = append(t.tokens, Token{Type: TokenEOF, Pos: t.pos})
	return t.tokens, nil
}

func (t *Tokenizer) nextToken() error {
	// Skip whitespace
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}

	if t.pos >= len(t.input) {
		return nil
	}

	startPos := t.pos
	ch := t.input[t.pos]

	switch ch {
	case '|':
		t.tokens = append(t.tokens, Token{Type: TokenPipe, Value: "|", Pos: startPos})
		t.pos++
		return nil
	case ',':
		t.tokens = append(t.tokens, Token{Type: TokenComma, Value: ",", Pos: startPos})
		t.pos++
		return nil
	case '(':
		t.tokens = append(t.tokens, Token{Type: TokenLParen, Value: "(", Pos: startPos})
		t.pos++
		return nil
	case ')':
		t.tokens = append(t.tokens, Token{Type: TokenRParen, Value: ")", Pos: startPos})
		t.pos++
		return nil
	case '+', '-':
		// Could be operator or part of a date/time
		if t.pos+1 < len(t.input) && unicode.IsDigit(rune(t.input[t.pos+1])) {
			return t.readNumberWithUnit()
		}
		t.tokens = append(t.tokens, Token{Type: TokenOperator, Value: string(ch), Pos: startPos})
		t.pos++
		return nil
	case '$':
		// Variable
		return t.readVariable()
	case '.':
		// Check for range operator ...
		if t.pos+2 < len(t.input) && t.input[t.pos:t.pos+3] == "..." {
			t.tokens = append(t.tokens, Token{Type: TokenRange, Value: "...", Pos: startPos})
			t.pos += 3
			return nil
		}
		return fmt.Errorf("unexpected character '.' at position %d", t.pos)
	default:
		if unicode.IsDigit(rune(ch)) {
			return t.readDateOrNumberWithUnit()
		}
		if unicode.IsLetter(rune(ch)) {
			return t.readKeywordOrDate()
		}
		return fmt.Errorf("unexpected character '%c' at position %d", ch, t.pos)
	}
}

func (t *Tokenizer) readVariable() error {
	startPos := t.pos
	t.pos++ // Skip $
	
	for t.pos < len(t.input) && (unicode.IsLetter(rune(t.input[t.pos])) || unicode.IsDigit(rune(t.input[t.pos]))) {
		t.pos++
	}
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenVariable, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) readNumberWithUnit() error {
	startPos := t.pos
	
	// Read sign if present
	if t.pos < len(t.input) && (t.input[t.pos] == '+' || t.input[t.pos] == '-') {
		t.pos++
	}
	
	// Read number
	for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
		t.pos++
	}
	
	// Read unit if present (d, w, M, Y, h, m, s, q)
	if t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == 'd' || ch == 'w' || ch == 'M' || ch == 'Y' || ch == 'h' || ch == 'm' || ch == 's' || ch == 'q' {
			t.pos++
		}
	}
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenUnit, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) readDateOrNumberWithUnit() error {
	// Try to read ISO date (YYYY-MM-DD) or time (HH:MM:SS)
	if t.isISODate() {
		return t.readISODate()
	}
	
	if t.isTime() {
		return t.readTime()
	}
	
	// Otherwise, treat as number with optional unit
	return t.readNumberWithUnit()
}

func (t *Tokenizer) isISODate() bool {
	// Simple check for YYYY-MM-DD pattern
	if t.pos+10 > len(t.input) {
		return false
	}
	
	str := t.input[t.pos:min(t.pos+10, len(t.input))]
	if len(str) >= 10 && str[4] == '-' && str[7] == '-' {
		return true
	}
	return false
}

func (t *Tokenizer) readISODate() error {
	startPos := t.pos
	
	// Read YYYY-MM-DD
	t.pos += 10
	
	// Check for optional time part
	if t.pos < len(t.input) && (t.input[t.pos] == 'T' || t.input[t.pos] == ' ') {
		hasTimePart := false
		// Verify it's actually a time (HH:MM pattern) before consuming
		if t.pos+3 <= len(t.input) && t.input[t.pos+3] == ':' {
			t.pos++
			// Read time part HH:MM:SS
			if t.pos+8 <= len(t.input) {
				t.pos += 8
			}
			hasTimePart = true
		} else if t.input[t.pos] == 'T' {
			// If it's 'T', still advance and try to read time
			t.pos++
			// Read time part HH:MM:SS
			if t.pos+8 <= len(t.input) {
				t.pos += 8
			}
			hasTimePart = true
		}
		// Otherwise, don't consume the space as it's not a time part
		
		// Check for optional timezone (only if we read a time part)
		if hasTimePart && t.pos < len(t.input) && (t.input[t.pos] == 'Z' || t.input[t.pos] == '+' || t.input[t.pos] == '-') {
			if t.input[t.pos] == 'Z' {
				t.pos++
			} else {
				// Read timezone offset
				t.pos++
				for t.pos < len(t.input) && (unicode.IsDigit(rune(t.input[t.pos])) || t.input[t.pos] == ':') {
					t.pos++
				}
			}
		}
	}
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenDate, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) isTime() bool {
	// Check for HH:MM or HH:MM:SS pattern
	if t.pos+5 > len(t.input) {
		return false
	}
	
	// Simple check for time pattern
	str := t.input[t.pos:min(t.pos+8, len(t.input))]
	if len(str) >= 5 && str[2] == ':' {
		return true
	}
	return false
}

func (t *Tokenizer) readTime() error {
	startPos := t.pos
	
	// Read HH:MM
	t.pos += 5
	
	// Check for optional :SS
	if t.pos+3 <= len(t.input) && t.input[t.pos] == ':' {
		t.pos += 3
	}
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenTime, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) readKeywordOrDate() error {
	startPos := t.pos
	
	// Read word
	for t.pos < len(t.input) && (unicode.IsLetter(rune(t.input[t.pos])) || unicode.IsDigit(rune(t.input[t.pos]))) {
		t.pos++
	}
	
	value := t.input[startPos:t.pos]
	lowerValue := strings.ToLower(value)
	
	// Check if it's a keyword
	keywords := []string{
		"today", "now", "yesterday", "tomorrow",
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
		"start", "end", "startOf", "endOf",
		"startOfDay", "endOfDay", "startOfWeek", "endOfWeek",
		"startOfMonth", "endOfMonth", "startOfYear", "endOfYear",
		"startOfQuarter", "endOfQuarter",
		"round", "trunc", "day", "time", "month", "year", "week", "quarter",
		"hour", "minute", "second",
	}
	
	for _, kw := range keywords {
		if lowerValue == strings.ToLower(kw) {
			t.tokens = append(t.tokens, Token{Type: TokenKeyword, Value: lowerValue, Pos: startPos})
			return nil
		}
	}
	
	// Otherwise treat as a date/string
	t.tokens = append(t.tokens, Token{Type: TokenDate, Value: value, Pos: startPos})
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}