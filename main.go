package main

import (
	"fmt"
	"os"

	"compiler/config"
	"compiler/lexer"
	"compiler/parser"
)

func main() {
	config.Init()
	// Initialize and run the lexer
	lex := lexer.New()
	lexerSuccess := lex.Tokenize()

	if !lexerSuccess {
		fmt.Fprintln(os.Stderr,
			"Compilation aborted due to lexer error. A complete log of this run can be found in: output.err")
		os.Exit(1)
	}

	// Initialize and run the parser
	pars := parser.New()
	parserSuccess := pars.Parse()

	if !parserSuccess {
		pars.ListErrors()
		fmt.Fprintln(os.Stderr,
			"Compilation aborted due to parser error. A complete log of this run can be found in: output.err")
		os.Exit(1)
	} else {
		fmt.Println("Compilation successful.")
	}
}
