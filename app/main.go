package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/codecrafters-io/shell-starter-go/app/lexer"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

var shell Shell
var savedFD int
var fdToOverride int

func main() {
	REPL()
}

func readInput(reader *bufio.Reader) (string, int) {

	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}

	trimmedInput := strings.TrimRight(input, "\n")

	return trimmedInput, len(trimmedInput)
}

func REPL() {

	stdinReader := bufio.NewReader(os.Stdin)

	currentDir, _ := os.Getwd()

	shell = Shell{
		CurrentDir: currentDir,
	}

	lexer := lexer.Lexer{}

	for {

		shell.printPrompt()

		input, count := readInput(stdinReader)

		if count == 0 {
			continue
		}

		tokens, _ := lexer.Lex(input)

		/* 		for _, t := range tokens {
		   			t.PrintDebug()
		   		}
		   		continue */

		p := parser.New(tokens)
		sc, parseErr := p.ParseSingleCommand()

		if parseErr != nil {
			fmt.Println(parseErr)
			continue
		}

		command := sc.Args[0]
		args := sc.Args[1:]

		// handle redirections
		if len(sc.Redirs) > 0 {

			lastRedir := sc.Redirs[len(sc.Redirs)-1]

			fdToOverride = lastRedir.FD

			// save original fd
			savedFD, _ = syscall.Dup(fdToOverride)

			// TODO: implment multios

			flags := os.O_CREATE | os.O_WRONLY

			if lastRedir.Append {
				flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
			}

			file, _ := os.OpenFile(
				lastRedir.To,
				flags,
				0644,
			)

			syscall.Dup2(int(file.Fd()), fdToOverride)

			//defer file.Close()

		}

		if handler, ok := builtInCommands[command]; ok {
			handler(args)
		} else {

			// search the command in PATH
			_, err := exec.LookPath(command)

			if err != nil {
				fmt.Printf("%s: command not found\n", command)
			} else {

				cmd := exec.Command(command, args...)

				cmd.Stdout = os.Stdout
				//cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				cmd.Run()
				/* if execErr != nil {
					if _, ok := err.(*exec.ExitError); !ok {
						fmt.Println("Execution error: ", execErr)
					}
				} */
			}
		}

		// restore fd
		if len(sc.Redirs) > 0 {
			syscall.Dup2(savedFD, fdToOverride)
			syscall.Close(savedFD)
		}
	}
}
