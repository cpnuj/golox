package main

import "errors"

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() (Expr, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	logger.DPrintf(parsedebug, "%v\n", expr.Accept(&ExprPrintVisitor{}))
	return expr, nil
}

func (p *Parser) atEnd() bool {
	return p.current >= len(p.tokens)
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) advance() Token {
	if !p.atEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) checkOne(token TokenType) bool {
	if p.atEnd() {
		return false
	}
	return p.peek().Type() == token
}

func (p *Parser) check(tokens ...TokenType) bool {
	for _, token := range tokens {
		if p.checkOne(token) {
			return true
		}
	}
	return false
}

func (p *Parser) consume(t TokenType, msg string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}
	return Token{}, errors.New(msg)
}

//
// CFG for expression:
//
// expression     → equality ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary
//                | primary ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")" ;
//

func (p *Parser) expression() (Expr, error) {
	return p.equality()
}

func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.check(BANG_EQUAL, EQUAL_EQUAL) {
		operator := p.advance()

		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = &ExprBinary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.check(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.advance()

		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = &ExprBinary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.check(MINUS, PLUS) {
		operator := p.advance()

		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = &ExprBinary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.check(SLASH, STAR) {
		operator := p.advance()

		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr = &ExprBinary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.check(BANG, MINUS) {
		operator := p.advance()

		unary, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr := &ExprUnary{
			UnaryOperator: operator,
			Expression:    unary,
		}

		return expr, nil
	}

	return p.primary()
}

func (p *Parser) primary() (Expr, error) {
	if p.check(NUMBER, STRING, TRUE, FALSE, NIL) {
		return &ExprLiteral{Value: p.advance()}, nil
	}

	if p.check(LEFT_PAREN) {
		p.advance()

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if _, err := p.consume(
			RIGHT_PAREN, "Expect ')' after expression",
		); err != nil {
			return nil, err
		}

		return &ExprGrouping{Expression: expr}, nil
	}

	return nil, errors.New("expect expression")
}
