package main

import (
	"errors"
	"fmt"
	"strconv"
)

// tokens
type TokenType int32

const (
	EOF TokenType = iota

	// single-character tokens
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Keywords
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
)

func (t TokenType) String() string {
	switch t {
	// single-character tokens
	case LEFT_PAREN:
		return "LEFT_PAREN"
	case RIGHT_PAREN:
		return "RIGHT_PAREN"
	case LEFT_BRACE:
		return "LEFT_BRACE"
	case RIGHT_BRACE:
		return "RIGHT_BRACE"
	case COMMA:
		return "COMMA"
	case DOT:
		return "DOT"
	case MINUS:
		return "MINUS"
	case PLUS:
		return "PLUS"
	case SEMICOLON:
		return "SEMICOLON"
	case SLASH:
		return "SLASH"
	case STAR:
		return "STAR"

	// One or two character tokens
	case BANG:
		return "BANG"
	case BANG_EQUAL:
		return "BANG_EQUAL"
	case EQUAL:
		return "EQUAL"
	case EQUAL_EQUAL:
		return "EQUAL_EQUAL"
	case GREATER:
		return "GREATER"
	case GREATER_EQUAL:
		return "GREATER_EQUAL"
	case LESS:
		return "LESS"
	case LESS_EQUAL:
		return "LESS_EQUAL"

	// Literals
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"

	// Keywords
	case AND:
		return "AND"
	case CLASS:
		return "CLASS"
	case ELSE:
		return "ELSE"
	case FALSE:
		return "FALSE"
	case FUN:
		return "FUN"
	case FOR:
		return "FOR"
	case IF:
		return "IF"
	case NIL:
		return "NIL"
	case OR:
		return "OR"
	case PRINT:
		return "PRINT"
	case RETURN:
		return "RETURN"
	case SUPER:
		return "SUPER"
	case THIS:
		return "THIS"
	case TRUE:
		return "TRUE"
	case VAR:
		return "VAR"
	case WHILE:
		return "WHILE"

	case EOF:
		return "EOF"

	default:
		return ""
	}
}

type Token struct {
	typ    TokenType
	lexeme string
	lexval interface{}
}

func (token Token) Type() TokenType {
	return token.typ
}

func (token Token) Value() interface{} {
	if token.lexval == nil {
		return token.lexeme
	}
	return token.lexval
}

func (token Token) String() string {
	return fmt.Sprintf("%s %s %v", token.typ, token.lexeme, token.lexval)
}

// helper for lexer
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		b == '_'
}

func isSpace(b byte) bool {
	return b == '\t' || b == ' ' || b == '\n'
}

// lexer
type Scanner struct {
	src     []byte
	start   int
	current int
	row     int
	col     int
	scanned bool
	errors  []error
	tokens  []Token
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src:     []byte(src),
		start:   0,
		current: 0,
		row:     1,
		col:     1,
		scanned: false,
		errors:  make([]error, 0),
		tokens:  make([]Token, 0),
	}
}

func (s *Scanner) Tokens() ([]Token, error) {
	s.scan()
	if s.hasError() {
		for _, err := range s.errors {
			logger.EPrintf("%s", err)
		}
		return nil, errors.New("Scanner Error")
	}
	for _, token := range s.tokens {
		logger.DPrintf(lexdebug, "%s\n", token)
	}
	return s.tokens, nil
}

func (s *Scanner) scan() {
	for !s.atEnd() {
		s.scanToken()
	}
	s.addToken(EOF, nil)
}

func (s *Scanner) hasError() bool {
	return len(s.errors) > 0
}

func (s *Scanner) scanToken() {
	b := s.advance()

	if isSpace(b) {
		// do nothing
	} else if isDigit(b) {
		s.number()
	} else if isAlpha(b) {
		s.keywordOrIdent()
	} else {
		s.other(b)
	}

	s.start = s.current
}

func (s *Scanner) addToken(typ TokenType, val interface{}) {
	s.tokens = append(s.tokens, Token{
		typ:    typ,
		lexeme: s.lexeme(),
		lexval: val,
	})
}

func (s *Scanner) atEnd() bool {
	return s.current >= len(s.src)
}

func (s *Scanner) advance() byte {
	if s.peek() == '\n' {
		s.row++
		s.col = 1
	} else if s.peek() == '\t' {
		s.col = s.col + (8 - s.col/8)
	} else {
		s.col++
	}
	c := s.src[s.current]
	s.current++
	return c
}

func (s *Scanner) peek() byte {
	if s.current >= len(s.src) {
		return 0
	}
	return s.src[s.current]
}

func (s *Scanner) lexeme() string {
	return string(s.src[s.start:s.current])
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' {
		s.advance()

		if !isDigit(s.peek()) {
			s.errors = append(s.errors, logger.NewError(s.row, s.col, "expect digit"))
			return
		}

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	f, err := strconv.ParseFloat(s.lexeme(), 64)
	if err != nil {
		panic(err)
	}

	s.addToken(NUMBER, f)
}

func (s *Scanner) string() {
	// row and col for report errors
	row, col := s.row, s.col-1

	var str string
	for !s.atEnd() {
		c := s.advance()
		if c == '"' {
			s.addToken(STRING, str)
			return
		}
		str += string(c)
	}

	s.errors = append(s.errors, logger.NewError(
		row, col, "Unterminated string",
	))
}

func (s *Scanner) other(b byte) {
	switch b {
	// single character
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '.':
		s.addToken(DOT, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case '*':
		s.addToken(STAR, nil)
	case '/':
		s.addToken(SLASH, nil)
	case ';':
		s.addToken(SEMICOLON, nil)

	// one or two characters
	case '!':
		if s.peek() == '=' {
			s.advance()
			s.addToken(BANG_EQUAL, nil)
		}
		s.addToken(BANG, nil)
	case '=':
		if s.peek() == '=' {
			s.advance()
			s.addToken(EQUAL_EQUAL, nil)
		} else {
			s.addToken(EQUAL, nil)
		}
	case '>':
		if s.peek() == '=' {
			s.advance()
			s.addToken(GREATER_EQUAL, nil)
		} else {
			s.addToken(GREATER, nil)
		}
	case '<':
		if s.peek() == '=' {
			s.advance()
			s.addToken(LESS_EQUAL, nil)
		} else {
			s.addToken(LESS, nil)
		}

	// string literal
	case '"':
		s.string()

	default:
		// NB: is safe to set col = s.col-1 here?
		s.errors = append(s.errors, logger.NewError(
			s.row, s.col-1, "Unknown character "+string(b),
		))
	}
}

var scannerKeywords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"fun":    FUN,
	"for":    FOR,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

func (s *Scanner) keywordOrIdent() {
	for isAlpha(s.peek()) || isDigit(s.peek()) {
		s.advance()
	}
	token, isKeyword := scannerKeywords[s.lexeme()]
	if isKeyword {
		// keyword with value
		switch token {
		case TRUE:
			s.addToken(TRUE, true)
		case FALSE:
			s.addToken(FALSE, false)
		default:
			s.addToken(token, nil)
		}
	} else {
		s.addToken(IDENTIFIER, s.lexeme())
	}
}
