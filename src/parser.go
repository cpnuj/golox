package main

import "fmt"

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

func (p *Parser) consume(t TokenType, msg string) Token {
	if p.check(t) {
		return p.advance()
	}
	panic(NewLoxError(ParseError, p.peek(), msg))
}

//
// CFG for program:
//
// program        → declaration* EOF ;
//
// declaration    → varDecl
//                → funDecl
//                | statement ;
//
// varDecl        → VAR IDENTIFIER "=" expression ;
//
// funDecl        → "fun" function ;
// function       → IDENTIFIER "(" parameters? ")" blockStmt ;
// parameters     → IDENTIFIER ("," IDENTIFIER)* ;
//
// statement      → exprStmt
//                | printStmt
//                | blockStmt
//                | ifStmt
//                | whileStmt
//                | forStmt
//                | returnStmt
//                | classStmt;
//
// exprStmt       → expression ";" ;
//
// printStmt      → "print" expression ";" ;
//
// blockStmt      → "{" declaration* "}" ;
//
// ifStmt         → "if" "(" expression ")" statement ("else" statement)? ;
//
// whileStmt      → "while" "(" expression ")" statement;
//
// forStmt        → "for" "(" (varDecl | exprStmt | ";") expression? ";" expression? ")" statement ;
//
// returnStmt     → "return" expression? ";" ;
//
// classStmt      → "class" IDENTIFIER ( "<" IDENTIFIER )? "{" function* "}" ;
//

func (p *Parser) Parse() ([]Stmt, error) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()

	statements := make([]Stmt, 0)
	for !p.match(EOF) {
		statement, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	NewAstPrinter().Print(statements)
	return statements, nil
}

func (p *Parser) declaration() (Stmt, error) {
	if p.match(VAR) {
		return p.varDeclaration()
	}
	if p.match(FUN) {
		return p.funDecl()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() (Stmt, error) {
	name := p.consume(IDENTIFIER, "need identifier")

	var initializer Expr
	var err error
	if p.match(EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(SEMICOLON, "expect ; after statement")

	return &StmtVar{Name: name, Initializer: initializer}, nil
}

func (p *Parser) funDecl() (Stmt, error) {
	value := p.consume(IDENTIFIER, "expect identifier")
	name := value.Value().(string)

	p.consume(LEFT_PAREN, "expect (")

	var params []string
	if p.check(RIGHT_PAREN) {
		params = make([]string, 0)
	} else {
		var err error
		params, err = p.parameters()
		if err != nil {
			return nil, err
		}
	}

	p.consume(RIGHT_PAREN, "expect )")
	p.consume(LEFT_BRACE, "expect {")

	body, err := p.blockStmt()
	if err != nil {
		return nil, err
	}

	return &StmtFun{
		Name:   name,
		Params: params,
		Body:   body.(*StmtBlock).Statements,
	}, nil
}

func (p *Parser) parameters() ([]string, error) {
	params := make([]string, 0)
	for {
		param := p.consume(IDENTIFIER, "expect identifier")
		params = append(params, param.Value().(string))
		if !p.match(COMMA) {
			break
		}
	}
	return params, nil
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
	if p.match(WHILE) {
		return p.whileStmt()
	}
	if p.match(FOR) {
		return p.forStmt()
	}
	if p.match(RETURN) {
		return p.returnStmt()
	}
	if p.match(CLASS) {
		return p.classStmt()
	}
	return p.exprStmt()
}

func (p *Parser) printStmt() (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(SEMICOLON, "expect ; after statement")
	return &StmtPrint{Expression: value}, nil
}

func (p *Parser) exprStmt() (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(SEMICOLON, "expect ; after statement")
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
	p.consume(RIGHT_BRACE, "expect }")
	return &StmtBlock{
		Statements: statements,
	}, nil
}

func (p *Parser) ifStmt() (Stmt, error) {
	p.consume(LEFT_PAREN, "expect ( after if")

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	p.consume(RIGHT_PAREN, "expect )")

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

func (p *Parser) whileStmt() (Stmt, error) {
	p.consume(LEFT_PAREN, "expect ( after while")

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	p.consume(RIGHT_PAREN, "expect )")

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &StmtWhile{
		Cond: cond,
		Body: body,
	}, nil
}

func (p *Parser) forStmt() (Stmt, error) {
	p.consume(LEFT_PAREN, "expect ( after for")

	var initializer Stmt
	var err error
	if p.match(SEMICOLON) {
		initializer = nil
	} else if p.match(VAR) {
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.exprStmt()
		if err != nil {
			return nil, err
		}
	}

	var condition Expr
	if !p.check(SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(SEMICOLON, "expect ;")

	var increment Expr
	if !p.check(RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(RIGHT_PAREN, "expect )")

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = &StmtBlock{
			Statements: []Stmt{
				body,
				&StmtExpression{increment},
			},
		}
	}

	if condition == nil {
		condition = &ExprLiteral{true}
	}

	body = &StmtWhile{
		Cond: condition,
		Body: body,
	}

	if initializer != nil {
		body = &StmtBlock{
			Statements: []Stmt{
				initializer,
				body,
			},
		}
	}

	return body, nil
}

func (p *Parser) returnStmt() (Stmt, error) {
	keyword := p.previous()

	var expr Expr
	var err error
	if !p.check(SEMICOLON) {
		expr, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(SEMICOLON, "expect ;")

	return &StmtReturn{
		Keyword: keyword,
		Value:   expr,
	}, nil
}

func (p *Parser) classStmt() (Stmt, error) {
	token := p.consume(IDENTIFIER, "expect indentifier")

	var superclass *ExprVariable
	if p.match(LESS) {
		t := p.consume(IDENTIFIER, "Expect superclass name")
		superclass = &ExprVariable{t}
	}

	p.consume(LEFT_BRACE, "expect {")

	methods := make([]*StmtFun, 0)
	for !p.check(RIGHT_BRACE) && !p.atEnd() {
		fun, err := p.funDecl()
		if err != nil {
			return nil, err
		}

		// TODO: error to logger
		funNode, ok := fun.(*StmtFun)
		if !ok {
			panic("Programming error: Expect type StmtFun")
		}

		methods = append(methods, funNode)
	}

	p.consume(RIGHT_BRACE, "expect }")

	return &StmtClass{
		Name:       token.lexeme,
		Superclass: superclass,
		Methods:    methods,
	}, nil
}

//
// CFG for expression:
//
// expression     → assignment ;
// assignment     → ( call "." )? IDENTIFIER "=" assignment ;
//                | logic_or ;
// logic_or       → logic_and ("or" logic_and)* ;
// logic_and      → equality ("and" equality)* ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary
//                | call ;
// call           → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")"
//                | IDENTIFIER ;
//                | "super" "." IDENTIFIER ;
//
// arguments      → expression ("," expression)*
//

func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		// record position of token "=" to report error
		tk := p.previous()

		right, err := p.assignment()
		if err != nil {
			return nil, err
		}

		switch left := expr.(type) {
		case *ExprVariable:
			expr = &ExprAssign{
				Name:  left.Name,
				Value: right,
			}
		case *ExprGet:
			expr = &ExprSet{
				Object: left.Object,
				Field:  left.Field,
				Value:  right,
				Dot:    left.Dot,
			}
		default:
			row, col := tk.Pos()
			return nil, logger.NewError(row, col, "invalid assign target")
		}
	}

	return expr, nil
}

func (p *Parser) or() (Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.check(OR) {
		operator := p.advance()

		right, err := p.or()
		if err != nil {
			return nil, err
		}

		expr = &ExprLogical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.check(AND) {
		operator := p.advance()

		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		expr = &ExprLogical{
			Left:     expr,
			Operator: operator,
			Right:    right,
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

	return p.call()
}

func (p *Parser) call() (Expr, error) {
	callee, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.check(LEFT_PAREN) {
			paren := p.advance()
			args, err := p.arguments()
			if err != nil {
				return nil, err
			}
			callee = &ExprCall{
				Callee: callee,
				Paren:  paren,
				Args:   args,
			}
		} else if p.check(DOT) {
			dot := p.advance()
			field := p.consume(IDENTIFIER, "expect identifier after dot")
			callee = &ExprGet{
				Object: callee,
				Field:  field,
				Dot:    dot,
			}
		} else {
			break
		}
	}

	return callee, nil
}

func (p *Parser) primary() (Expr, error) {
	if p.check(NUMBER, STRING, TRUE, FALSE, NIL) {
		return &ExprLiteral{Value: p.advance().Value()}, nil
	}

	if p.check(LEFT_PAREN) {
		p.advance()

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		p.consume(RIGHT_PAREN, "Expect ')' after expression")

		return &ExprGrouping{Expression: expr}, nil
	}

	if p.check(IDENTIFIER) {
		return &ExprVariable{
			Name: p.advance(),
		}, nil
	}

	if p.check(THIS) {
		return &ExprThis{
			Keyword: p.advance(),
		}, nil
	}

	if p.check(SUPER) {
		super := p.advance()
		p.consume(DOT, "Expect '.' after 'super'.")
		name := p.consume(IDENTIFIER, "Expect superclass method name")
		return &ExprSuper{
			Keyword: super,
			Method:  name,
		}, nil
	}

	row, col := p.peek().Pos()
	return nil, logger.NewError(row, col, "expect expression")
}

func (p *Parser) arguments() ([]Expr, error) {
	// no args
	args := make([]Expr, 0)
	if p.match(RIGHT_PAREN) {
		return args, nil
	}

	arg, err := p.expression()
	if err != nil {
		return nil, err
	}

	args = append(args, arg)
	for p.match(COMMA) {
		arg, err = p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	p.consume(RIGHT_PAREN, "expect )")

	return args, nil
}
