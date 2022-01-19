package main

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

func (p *Parser) match(tokens ...TokenType) bool {
	if p.check(tokens...) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(t TokenType, msg string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}
	row, col := p.peek().Pos()
	return Token{}, logger.NewError(row, col, msg)
}

//
// CFG for program:
//
// program        → declaration* EOF ;
//
// declaration    → varDecl
//                | statement ;
//
// varDecl        → VAR IDENTIFIER "=" expression ;
//
// statement      → exprStmt
//                | printStmt
//                | blockStmt
//                | ifStmt ;
//
// exprStmt       → expression ";" ;
//
// printStmt      → "print" expression ";" ;
//
// blockStmt      → "{" declaration* "}" ;
//
// ifStmt         → "if" "(" expression ")" statement ("else" statement)? ;
//

func (p *Parser) Parse() ([]Stmt, error) {
	statements := make([]Stmt, 0)
	for !p.match(EOF) {
		statement, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	return statements, nil
}

func (p *Parser) declaration() (Stmt, error) {
	if p.match(VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() (Stmt, error) {
	name, err := p.consume(IDENTIFIER, "need identifier")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if p.match(EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(SEMICOLON, "expect ; after statement")
	if err != nil {
		return nil, err
	}

	return &StmtVar{Name: name, Initializer: initializer}, nil
}

func (p *Parser) statement() (Stmt, error) {
	if p.match(PRINT) {
		return p.printStmt()
	}
	if p.match(LEFT_BRACE) {
		return p.blockStmt()
	}
	if p.match(IF) {
		return p.ifStmt()
	}
	return p.exprStmt()
}

func (p *Parser) printStmt() (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(SEMICOLON, "expect ; after statement")
	if err != nil {
		return nil, err
	}

	return &StmtPrint{Expression: value}, nil
}

func (p *Parser) exprStmt() (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(SEMICOLON, "expect ; after statement")
	if err != nil {
		return nil, err
	}

	return &StmtExpression{Expression: value}, nil
}

func (p *Parser) blockStmt() (Stmt, error) {
	statements := make([]Stmt, 0)
	for !p.atEnd() && !p.check(RIGHT_BRACE) {
		statement, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}

	_, err := p.consume(RIGHT_BRACE, "expect }")
	if err != nil {
		return nil, err
	}

	return &StmtBlock{
		Statements: statements,
	}, nil
}

func (p *Parser) ifStmt() (Stmt, error) {
	_, err := p.consume(LEFT_PAREN, "expect ( after if")
	if err != nil {
		return nil, err
	}

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(RIGHT_PAREN, "expect )")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if p.match(ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &StmtIf{
		Cond: cond,
		Then: thenBranch,
		Else: elseBranch,
	}, nil
}

//
// CFG for expression:
//
// expression     → assignment ;
// assignment     → IDENTIFIER "=" assignment
//                | equality ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary
//                | primary ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")"
//                | IDENTIFIER ;
//

func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		left, ok := expr.(*ExprVariable)
		if !ok {
			row, col := p.previous().Pos()
			return nil, logger.NewError(row, col, "invalid assign target")
		}

		right, err := p.assignment()
		if err != nil {
			return nil, err
		}

		expr = &ExprAssign{
			Name:  left.Name,
			Value: right,
		}
	}

	return expr, nil
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

		_, err = p.consume(RIGHT_PAREN, "Expect ')' after expression")
		if err != nil {
			return nil, err
		}

		return &ExprGrouping{Expression: expr}, nil
	}

	if p.check(IDENTIFIER) {
		return &ExprVariable{
			Name: p.advance(),
		}, nil
	}

	row, col := p.peek().Pos()
	return nil, logger.NewError(row, col, "expect expression")
}
