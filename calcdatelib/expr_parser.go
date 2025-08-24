package calcdatelib

import (
	"fmt"
	"strings"
)

// ExprParser parses date expressions.
type ExprParser struct {
	tokens []Token
	pos    int
}

// NewExprParser creates a new expression parser.
func NewExprParser(_ string) *ExprParser {
	return &ExprParser{}
}

// Parse parses a date expression string
//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) Parse(input string) (ExprNode, error) {
	// Tokenize input
	tokenizer := NewTokenizer(input)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	
	p.tokens = tokens
	p.pos = 0
	
	return p.parseExpression()
}

// ParseTransform parses a transform expression with comma-separated begin and end expressions.
func (p *ExprParser) ParseTransform(input string) (*TransformNode, error) {
	// Split by comma to get begin and end expressions
	parts := strings.Split(input, ",")
	if len(parts) != TransformParts {
		return nil, ErrTransformPartsInvalid
	}
	
	// Parse begin expression
	beginExpr, err := p.Parse(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse begin expression: %w", err)
	}
	
	// Parse end expression
	endExpr, err := p.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse end expression: %w", err)
	}
	
	return &TransformNode{
		BeginExpr: beginExpr,
		EndExpr:   endExpr,
	}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseExpression() (ExprNode, error) {
	// Check for range expression (date...date)
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	
	// Check for range operator
	if p.current().Type == TokenRange {
		return p.parseRangeExpression(node)
	}
	
	// Check for pipe operations
	if p.current().Type == TokenPipe {
		return p.parsePipeline(node)
	}
	
	// Check for operations without pipe
	return p.parseOperationsAfterNode(node)
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseRangeExpression(startNode ExprNode) (ExprNode, error) {
	p.advance() // consume range operator
	endNode, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	rangeNode := &RangeNode{Start: startNode, End: endNode}
	
	// Check for pipe operations after the range
	if p.current().Type == TokenPipe {
		return p.parsePipeline(rangeNode)
	}
	
	return rangeNode, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseOperationsAfterNode(node ExprNode) (ExprNode, error) {
	ops := []ExprNode{}
	for p.isOperationToken() {
		op, err := p.parseOperation()
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	
	if len(ops) > 0 {
		// Check if there's a pipe after the operations
		if p.current().Type == TokenPipe {
			// Create a pipe node with the operations collected so far
			pipeNode := &PipeNode{Base: node, Operations: ops}
			// Continue parsing the pipeline
			return p.parsePipeline(pipeNode)
		}
		return &PipeNode{Base: node, Operations: ops}, nil
	}
	
	return node, nil
}

func (p *ExprParser) isOperationToken() bool {
	t := p.current().Type
	return t == TokenOperator || t == TokenUnit || t == TokenKeyword
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parsePipeline(base ExprNode) (ExprNode, error) {
	operations := []ExprNode{}
	
	for p.current().Type == TokenPipe {
		p.advance() // consume pipe
		
		// Parse the operation after the pipe
		op, err := p.parseOperation()
		if err != nil {
			return nil, err
		}
		operations = append(operations, op)
		
		// Continue parsing operations without pipe
		for p.isOperationToken() {
			op, err := p.parseOperation()
			if err != nil {
				return nil, err
			}
			operations = append(operations, op)
		}
	}
	
	return &PipeNode{Base: base, Operations: operations}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parsePrimary() (ExprNode, error) {
	token := p.current()
	
	switch token.Type {
	case TokenVariable:
		return p.parseVariableToken(token)
	case TokenKeyword:
		return p.parseKeywordToken(token)
	case TokenDate, TokenTime:
		return p.parseDateTimeToken(token)
	case TokenOperator, TokenUnit:
		return p.parseOperatorUnitToken(token)
	case TokenEOF:
		return nil, ErrUnexpectedEndOfExpression
	case TokenPipe, TokenRange, TokenNumber, TokenComma, TokenLParen, TokenRParen:
		return nil, fmt.Errorf("%w: %v", ErrUnexpectedToken, token)
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnexpectedToken, token)
	}
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseVariableToken(token Token) (ExprNode, error) {
	p.advance()
	return &VariableNode{Name: token.Value}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseKeywordToken(token Token) (ExprNode, error) {
	// Check if it's a date keyword
	dateKeywords := []string{"today", "now", "yesterday", "tomorrow", 
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, kw := range dateKeywords {
		if token.Value == kw {
			p.advance()
			return &DateNode{Value: token.Value}, nil
		}
	}
	// Otherwise it's an operation keyword
	return p.parseOperation()
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseDateTimeToken(token Token) (ExprNode, error) {
	p.advance()
	return &DateNode{Value: token.Value}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseOperatorUnitToken(token Token) (ExprNode, error) {
	// Relative date like "+1d" or "-2w"
	p.advance()
	return &DateNode{Value: token.Value}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseOperation() (ExprNode, error) {
	token := p.current()
	
	switch token.Type {
	case TokenOperator:
		return p.parseOperatorOperation(token)
	case TokenUnit:
		return p.parseUnitOperation(token)
	case TokenKeyword:
		return p.parseKeywordOperation(token)
	case TokenEOF, TokenDate, TokenPipe, TokenRange, TokenVariable,
		TokenNumber, TokenTime, TokenComma, TokenLParen, TokenRParen:
		return nil, fmt.Errorf("%w: got %v", ErrExpectedOperationAfterPipe, token)
	default:
		return nil, fmt.Errorf("%w: got %v", ErrExpectedOperationAfterPipe, token)
	}
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseOperatorOperation(token Token) (ExprNode, error) {
	op := token.Value
	p.advance()
	
	if p.current().Type != TokenUnit && p.current().Type != TokenNumber {
		return nil, fmt.Errorf("%w %s", ErrExpectedUnitAfterOperator, op)
	}
	
	value := p.current().Value
	p.advance()
	
	return &OperationNode{Op: op, Value: value}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseUnitOperation(token Token) (ExprNode, error) {
	value := token.Value
	p.advance()
	
	// Determine operator from the value
	op := "+"
	if strings.HasPrefix(value, "-") {
		op = "-"
		value = value[1:]
	} else if strings.HasPrefix(value, "+") {
		value = value[1:]
	}
	
	return &OperationNode{Op: op, Value: value}, nil
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseKeywordOperation(token Token) (ExprNode, error) {
	keyword := token.Value
	p.advance()
	
	// Check for operations that take arguments
	if p.isArgumentOperation(keyword) {
		return p.parseArgumentOperation(keyword)
	}
	
	// Boundary operations (startOf*, endOf*)
	if p.isBoundaryOperation(keyword) {
		return &OperationNode{Op: keyword, Value: ""}, nil
	}
	
	return nil, fmt.Errorf("%w: %s", ErrUnknownOperation, keyword)
}

func (p *ExprParser) isArgumentOperation(keyword string) bool {
	const (
		dayKeyword   = "day"
		timeKeyword  = "time"
		roundKeyword = "round"
		truncKeyword = "trunc"
	)
	return keyword == dayKeyword || keyword == timeKeyword || keyword == roundKeyword || keyword == truncKeyword
}

func (p *ExprParser) isBoundaryOperation(keyword string) bool {
	return strings.HasPrefix(keyword, "startof") || strings.HasPrefix(keyword, "endof") ||
		keyword == "start" || keyword == "end"
}

//nolint:ireturn // returns interface by design for AST nodes
func (p *ExprParser) parseArgumentOperation(keyword string) (ExprNode, error) {
	if p.current().Type == TokenNumber || p.current().Type == TokenTime || p.current().Type == TokenKeyword {
		value := p.current().Value
		p.advance()
		return &OperationNode{Op: keyword, Value: value}, nil
	}
	// If no argument, use default
	return &OperationNode{Op: keyword, Value: ""}, nil
}

func (p *ExprParser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *ExprParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// peek returns the next token without advancing
//nolint:unused // may be used in future
func (p *ExprParser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+1]
}