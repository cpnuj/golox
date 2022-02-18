package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var interpreter *Interpreter

func init() {
	interpreter = NewInterpreter()
}

func run(src string) error {
	logger.Reset(src, os.Stdout, os.Stderr)

	scanner := NewScanner(src)
	tokens, err := scanner.Tokens()
	if err != nil {
		return err
	}

	parser := NewParser(tokens)
	statements, err := parser.Parse()
	if err != nil {
		return err
	}

	resolver := NewResolver()
	locals, err := resolver.Resolve(statements)
	if err != nil {
		return err
	}

	interpreter.SetLocals(locals)
	return interpreter.Interprete(statements)
}

func runFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return run(string(data))
}

func prompt() {
	fmt.Printf(">>> ")
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
			fmt.Println(err)
		}
	}
}

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage %s [filename]", os.Args[0])
		return
	}

	// script file
	if len(os.Args) == 2 {
		if err := runFile(os.Args[1]); err != nil {
			fmt.Println(err)
		}
		return
	}

	// repl
	if err := runPrompt(); err != nil {
		fmt.Println(err)
	}
}
