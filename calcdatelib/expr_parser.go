package calcdatelib

import (
	"fmt"
	"strings"
)

// ExprParser parses date expressions
type ExprParser struct {
	tokens []Token
	pos    int
}

// NewExprParser creates a new expression parser
func NewExprParser(input string) *ExprParser {
	return &ExprParser{}
}

// Parse parses a date expression string
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

// ParseTransform parses a transform expression with comma-separated begin and end expressions
func (p *ExprParser) ParseTransform(input string) (*TransformNode, error) {
	// Split by comma to get begin and end expressions
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("transform must have exactly two parts separated by comma")
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

func (p *ExprParser) parseExpression() (ExprNode, error) {
	// Check for range expression (date...date)
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	
	// Check for range operator
	if p.current().Type == TokenRange {
		p.advance()
		endNode, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		rangeNode := &RangeNode{Start: node, End: endNode}
		
		// Check for pipe operations after the range
		if p.current().Type == TokenPipe {
			return p.parsePipeline(rangeNode)
		}
		
		return rangeNode, nil
	}
	
	// Check for pipe operations
	if p.current().Type == TokenPipe {
		return p.parsePipeline(node)
	}
	
	// Check for operations without pipe (e.g., "today +1d")
	ops := []ExprNode{}
	for p.current().Type == TokenOperator || p.current().Type == TokenUnit || p.current().Type == TokenKeyword {
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
		
		// Continue parsing operations without pipe (e.g., "| endOfMonth +1d +1s")
		for p.current().Type == TokenOperator || p.current().Type == TokenUnit || p.current().Type == TokenKeyword {
			op, err := p.parseOperation()
			if err != nil {
				return nil, err
			}
			operations = append(operations, op)
		}
	}
	
	return &PipeNode{Base: base, Operations: operations}, nil
}

func (p *ExprParser) parsePrimary() (ExprNode, error) {
	token := p.current()
	
	switch token.Type {
	case TokenVariable:
		p.advance()
		return &VariableNode{Name: token.Value}, nil
		
	case TokenKeyword:
		// Check if it's a date keyword (today, now, yesterday, tomorrow, weekday names)
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
		
	case TokenDate:
		p.advance()
		return &DateNode{Value: token.Value}, nil
		
	case TokenTime:
		p.advance()
		return &DateNode{Value: token.Value}, nil
		
	case TokenOperator, TokenUnit:
		// Relative date like "+1d" or "-2w"
		p.advance()
		return &DateNode{Value: token.Value}, nil
		
	case TokenEOF:
		return nil, fmt.Errorf("unexpected end of expression")
		
	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

func (p *ExprParser) parseOperation() (ExprNode, error) {
	token := p.current()
	
	switch token.Type {
	case TokenOperator:
		// Operator followed by unit (e.g., "+1d", "-2w")
		op := token.Value
		p.advance()
		
		if p.current().Type != TokenUnit && p.current().Type != TokenNumber {
			return nil, fmt.Errorf("expected unit or number after operator %s", op)
		}
		
		value := p.current().Value
		p.advance()
		
		// Combine operator and value
		return &OperationNode{Op: op, Value: value}, nil
		
	case TokenUnit:
		// Direct unit like "1d", "+2w"
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
		
	case TokenKeyword:
		keyword := token.Value
		p.advance()
		
		// Check for operations that take arguments
		if keyword == "day" || keyword == "time" || keyword == "round" || keyword == "trunc" {
			// These operations expect an argument
			if p.current().Type == TokenNumber || p.current().Type == TokenTime || p.current().Type == TokenKeyword {
				value := p.current().Value
				p.advance()
				return &OperationNode{Op: keyword, Value: value}, nil
			}
			// If no argument, use default
			return &OperationNode{Op: keyword, Value: ""}, nil
		}
		
		// Boundary operations (startOf*, endOf*)
		if strings.HasPrefix(keyword, "startof") || strings.HasPrefix(keyword, "endof") ||
		   keyword == "start" || keyword == "end" {
			return &OperationNode{Op: keyword, Value: ""}, nil
		}
		
		return nil, fmt.Errorf("unknown operation keyword: %s", keyword)
		
	default:
		return nil, fmt.Errorf("expected operation, got %v", token)
	}
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

func (p *ExprParser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+1]
}