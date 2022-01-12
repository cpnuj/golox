package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

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
