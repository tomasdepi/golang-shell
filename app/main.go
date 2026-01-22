package main

import (
	"fmt"
	"os"

	"github.com/tomasdepi/golang-shell/app/lexer"
	"github.com/tomasdepi/golang-shell/app/parser"
	"github.com/tomasdepi/golang-shell/app/shell"
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

		shell.Prompt.Render()
		input := shell.ReadlineFromShell()

		err := term.Restore(fd, oldState)

		if err != nil {
			panic(err)
		}

		if len(input) == 0 {
			continue
		}

		// TODO: figure out if this is the optimal place to put
		shell.HistoryManager.Add(input)

		tokens, _ := lexer.Lex(input)

		p := parser.New(tokens)
		pipeline, parseErr := p.ParseCommand()

		if parseErr != nil {
			fmt.Println(parseErr)
			continue
		}

		err = shell.Execute(pipeline)

		if err != nil {
			fmt.Println(err)
		}

	}
}
