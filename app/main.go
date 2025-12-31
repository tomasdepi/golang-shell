package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/shell-starter-go/app/lexer"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
	"golang.org/x/term"
)

func main() {
	REPL()
}

func REPL() {

	currentDir, _ := os.Getwd()

	shell := shell.NewShell(currentDir)

	lexer := lexer.Lexer{}

	for {

		fd := int(os.Stdin.Fd())

		oldState, _ := term.MakeRaw(fd)

		input := shell.ReadlineFromShell()

		err := term.Restore(fd, oldState)

		if err != nil {
			panic(err)
		}

		if len(input) == 0 {
			continue
		}

		tokens, _ := lexer.Lex(input)

		p := parser.New(tokens)
		sc, parseErr := p.ParseSingleCommand()

		if parseErr != nil {
			fmt.Println(parseErr)
			continue
		}

		err = shell.Execute(sc)

		if err != nil {
			fmt.Println(err)
		}

	}
}
