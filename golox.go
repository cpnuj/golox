package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// debug flags
var lexdebug = 1 // lexer debug

type Logger struct {
	lines   []string
	dwriter io.Writer // debug writer
	ewriter io.Writer // error writer
}

var logger = &Logger{}

func (l *Logger) Reset(src string, dwriter, ewriter io.Writer) {
	l.lines = []string{""}
	l.lines = append(l.lines, strings.Split(src, "\n")...)
	l.dwriter = dwriter
	l.ewriter = ewriter
}

func (l *Logger) NewError(row, col int, errmsg string) error {
	prefix := fmt.Sprintf("    %d | ", row)
	lineMsg := fmt.Sprintf("%s%s", prefix, l.lines[row])
	pointer := ""
	for i := 1; i < len(prefix)+col; i++ {
		pointer += " "
	}
	pointer += "^"
	return fmt.Errorf("Error: %s\n%s\n%s\n", errmsg, lineMsg, pointer)
}

func (l *Logger) DPrintf(dflag int, format string, a ...interface{}) {
	if dflag > 0 {
		fmt.Fprintf(l.dwriter, format, a...)
	}
}

func (l *Logger) EPrintf(format string, a ...interface{}) {
	fmt.Fprintf(l.ewriter, format, a...)
}

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
	default:
		return ""
	}
}

type Token struct {
	typ    TokenType
	lexeme string
	lexval interface{}
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

func (s *Scanner) scan() {
	for !s.atEnd() {
		s.scanToken()
	}
	s.scanned = true
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
		}
		s.addToken(EQUAL, nil)
	case '>':
		if s.peek() == '=' {
			s.advance()
			s.addToken(GREATER_EQUAL, nil)
		}
		s.addToken(GREATER, nil)
	case '<':
		if s.peek() == '=' {
			s.advance()
			s.addToken(LESS_EQUAL, nil)
		}
		s.addToken(LESS, nil)

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

func (s *Scanner) Token() (Token, error) {
	if !s.scanned {
		s.scan()
	}
	if len(s.tokens) == 0 {
		return Token{}, fmt.Errorf("")
	}
	t := s.tokens[0]
	s.tokens = s.tokens[1:]
	return t, nil
}

func run(src string) error {
	logger.Reset(src, os.Stdout, os.Stderr)

	scanner := NewScanner(src)
	token, err := scanner.Token()
	for ; err == nil; token, err = scanner.Token() {
		logger.DPrintf(lexdebug, "%s\n", token)
	}

	if scanner.hasError() {
		for _, err := range scanner.errors {
			logger.EPrintf("%s", err)
		}
		return errors.New("scanner error")
	}

	return nil
}

func runFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return run(string(data))
}

func prompt() {
	fmt.Printf("> ")
}

func runPrompt() error {
	var s string
	var err error
	reader := bufio.NewReader(os.Stdin)
	for {
		prompt()
		s, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			return err
		}

		err := run(s)
		if err != nil {
			return err
		}
	}
}

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage %s [filename]", os.Args[0])
	} else if len(os.Args) == 2 {
		if err := runFile(os.Args[1]); err != nil {
			fmt.Println(err)
		}
	} else {
		if err := runPrompt(); err != nil {
			fmt.Println(err)
		}
	}
}
