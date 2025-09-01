package calcdate

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token.
type TokenType int

// Token types for expression parsing.
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

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Tokenizer tokenizes date expressions.
type Tokenizer struct {
	input string
	pos   int
	tokens []Token
}

// NewTokenizer creates a new tokenizer.
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: input,
		pos:   0,
		tokens: []Token{},
	}
}

// Tokenize tokenizes the input string.
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
	t.skipWhitespace()

	if t.pos >= len(t.input) {
		return nil
	}

	startPos := t.pos
	ch := t.input[t.pos]

	// Handle single character tokens
	if tokenType, isSingle := t.getSingleCharToken(ch); isSingle {
		t.tokens = append(t.tokens, Token{Type: tokenType, Value: string(ch), Pos: startPos})
		t.pos++
		return nil
	}

	// Handle special cases
	return t.handleSpecialChar(ch, startPos)
}

func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
}

func (t *Tokenizer) getSingleCharToken(ch byte) (TokenType, bool) {
	switch ch {
	case '|':
		return TokenPipe, true
	case ',':
		return TokenComma, true
	case '(':
		return TokenLParen, true
	case ')':
		return TokenRParen, true
	default:
		return TokenEOF, false
	}
}

func (t *Tokenizer) handleSpecialChar(ch byte, startPos int) error {
	switch ch {
	case '+', '-':
		return t.handlePlusMinusToken(ch, startPos)
	case '$':
		return t.readVariable()
	case '.':
		return t.handleDotToken(startPos)
	default:
		if unicode.IsDigit(rune(ch)) {
			return t.readDateOrNumberWithUnit()
		}
		if unicode.IsLetter(rune(ch)) {
			return t.readKeywordOrDate()
		}
		return fmt.Errorf("%w: '%c' at position %d", ErrUnexpectedCharacter, ch, t.pos)
	}
}

func (t *Tokenizer) handlePlusMinusToken(ch byte, startPos int) error {
	if t.pos+1 < len(t.input) && unicode.IsDigit(rune(t.input[t.pos+1])) {
		return t.readNumberWithUnit()
	}
	t.tokens = append(t.tokens, Token{Type: TokenOperator, Value: string(ch), Pos: startPos})
	t.pos++
	return nil
}

func (t *Tokenizer) handleDotToken(startPos int) error {
	if t.pos+2 < len(t.input) && t.input[t.pos:t.pos+3] == "..." {
		t.tokens = append(t.tokens, Token{Type: TokenRange, Value: "...", Pos: startPos})
		t.pos += 3
		return nil
	}
	return fmt.Errorf("%w: '.' at position %d", ErrUnexpectedCharacter, t.pos)
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
	
	t.readSign()
	t.readDigits()
	t.readUnit()
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenUnit, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) readSign() {
	if t.pos < len(t.input) && (t.input[t.pos] == '+' || t.input[t.pos] == '-') {
		t.pos++
	}
}

func (t *Tokenizer) readDigits() {
	for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
		t.pos++
	}
}

func (t *Tokenizer) readUnit() {
	if t.pos < len(t.input) && t.isUnitChar(t.input[t.pos]) {
		t.pos++
	}
}

func (t *Tokenizer) isUnitChar(ch byte) bool {
	return ch == 'd' || ch == 'w' || ch == 'M' || ch == 'Y' || 
		ch == 'h' || ch == 'm' || ch == 's' || ch == 'q'
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
	
	const isoDateLength = 10
	str := t.input[t.pos:minInt(t.pos+isoDateLength, len(t.input))]
	if len(str) >= 10 && str[4] == '-' && str[7] == '-' {
		return true
	}
	return false
}

func (t *Tokenizer) readISODate() error {
	startPos := t.pos
	
	// Read YYYY-MM-DD
	const datePartLength = 10
	t.pos += datePartLength
	
	// Check for optional time part
	hasTimePart := t.readOptionalTimePart()
	
	// Check for optional timezone (only if we read a time part)
	if hasTimePart {
		t.readOptionalTimezone()
	}
	
	value := t.input[startPos:t.pos]
	t.tokens = append(t.tokens, Token{Type: TokenDate, Value: value, Pos: startPos})
	return nil
}

func (t *Tokenizer) readOptionalTimePart() bool {
	if t.pos >= len(t.input) || (t.input[t.pos] != 'T' && t.input[t.pos] != ' ') {
		return false
	}
	
	return t.tryReadTimePart()
}

func (t *Tokenizer) tryReadTimePart() bool {
	const timePatternOffset = 3
	const timePartLength = 8
	
	// Verify it's actually a time (HH:MM pattern) before consuming
	if t.pos+timePatternOffset <= len(t.input) && t.input[t.pos+timePatternOffset] == ':' {
		t.pos++
		if t.pos+timePartLength <= len(t.input) {
			t.pos += timePartLength
		}
		return true
	}
	
	if t.input[t.pos] == 'T' {
		t.pos++
		if t.pos+timePartLength <= len(t.input) {
			t.pos += timePartLength
		}
		return true
	}
	
	return false
}

func (t *Tokenizer) readOptionalTimezone() {
	if t.pos >= len(t.input) {
		return
	}
	
	ch := t.input[t.pos]
	switch ch {
	case 'Z':
		t.pos++
	case '+', '-':
		t.pos++
		for t.pos < len(t.input) && (unicode.IsDigit(rune(t.input[t.pos])) || t.input[t.pos] == ':') {
			t.pos++
		}
	case ' ':
		// Check if there's a timezone name after the space
		t.readOptionalTimezoneNameAfterSpace()
	default:
		// Check if we're at the start of a timezone name (no space prefix)
		t.readOptionalTimezoneName()
	}
}

func (t *Tokenizer) readOptionalTimezoneNameAfterSpace() {
	if t.pos >= len(t.input) || t.input[t.pos] != ' ' {
		return
	}
	
	// Skip the space
	spacePos := t.pos
	t.pos++
	
	// Try to read timezone name
	if t.tryReadTimezoneName() {
		// Successfully read timezone name, keep the position
		return
	}
	
	// If we couldn't read a timezone name, revert to before the space
	t.pos = spacePos
}

func (t *Tokenizer) readOptionalTimezoneName() {
	t.tryReadTimezoneName()
}

func (t *Tokenizer) tryReadTimezoneName() bool {
	startPos := t.pos
	
	const maxTimezoneNameLength = 5
	const minTimezoneNameLength = 2
	
	// Read potential timezone name (letters only, 2-5 characters)
	for t.pos < len(t.input) && unicode.IsLetter(rune(t.input[t.pos])) && (t.pos-startPos) < maxTimezoneNameLength {
		t.pos++
	}
	
	// Check if we read a reasonable timezone name length
	nameLength := t.pos - startPos
	if nameLength < minTimezoneNameLength || nameLength > maxTimezoneNameLength {
		t.pos = startPos
		return false
	}
	
	// Validate that this looks like a timezone name
	potentialTZ := t.input[startPos:t.pos]
	if t.isValidTimezoneName(potentialTZ) {
		return true
	}
	
	t.pos = startPos
	return false
}

func (t *Tokenizer) isValidTimezoneName(name string) bool {
	// Common timezone abbreviations and names
	validTimezones := []string{
		"UTC", "GMT", "EST", "CST", "MST", "PST", "EDT", "CDT", "MDT", "PDT",
		"CET", "CEST", "JST", "IST", "BST", "AEST", "AEDT",
	}
	
	upperName := strings.ToUpper(name)
	for _, tz := range validTimezones {
		if upperName == tz {
			return true
		}
	}
	return false
}

func (t *Tokenizer) isTime() bool {
	// Check for HH:MM or HH:MM:SS pattern
	if t.pos+5 > len(t.input) {
		return false
	}
	
	// Simple check for time pattern
	const maxTimeLength = 8
	str := t.input[t.pos:minInt(t.pos+maxTimeLength, len(t.input))]
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
		"startOfHour", "endOfHour", "startOfMinute", "endOfMinute", "startOfSecond", "endOfSecond",
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

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}