package main

import (
	"fmt"
	"io"
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
