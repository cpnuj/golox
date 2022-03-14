package main

import (
	"fmt"
)

type LoxErrorType int

const (
	ParseError LoxErrorType = iota
	ResolveError
	RuntimeError
)

type LoxError struct {
	t   LoxErrorType
	msg string
	tk  Token // used by parse error
}

func (e *LoxError) String() string {
	var ret string
	switch e.t {
	case ParseError, ResolveError:
		ret = fmt.Sprintf("[line %d] Error at '%s': %s", e.tk.row, e.tk.lexeme, e.msg)
	case RuntimeError:
		ret = fmt.Sprintf("%s\n[line %d]", e.msg, e.tk.row)
	default:
		ret = "Unknown error type"
	}
	return ret
}

func NewLoxError(t LoxErrorType, tk Token, msg string) *LoxError {
	return &LoxError{
		t:   t,
		msg: msg,
		tk:  tk,
	}
}
