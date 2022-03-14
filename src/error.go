package main

import (
	"fmt"
)

const (
	ParseError int = iota
	ResolveError
	RuntimeError
)

type LoxError struct {
	t    int
	line int
	msg  string
	tk   Token // used by parse error
}

func (e *LoxError) String() string {
	var ret string
	switch e.t {
	case ParseError:
		ret = fmt.Sprintf("[line %d] Error at '%s': %s", e.line, e.tk.lexeme, e.msg)
	case RuntimeError:
		ret = fmt.Sprintf("%s\n[line %d]", e.msg, e.line)
	default:
		ret = "Unknown error type"
	}
	return ret
}

func NewParseError(tk Token, msg string) *LoxError {
	return &LoxError{
		t:    ParseError,
		line: tk.row,
		msg:  msg,
		tk:   tk,
	}
}

func NewRuntimeError(line int, msg string) *LoxError {
	return &LoxError{
		t:    RuntimeError,
		line: line,
		msg:  msg,
	}
}
