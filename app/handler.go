package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const (
	EXIT_COMMAND = "exit"
	ECHO_COMMAND = "echo"
	TYPE_COMMAND = "type"
	PWD_COMMAND  = "pwd"
	CD_COMMAND   = "cd"
)

type CommandHandler func(args []string) error

var builtInCommands = map[string]CommandHandler{
	EXIT_COMMAND: func(args []string) error {
		os.Exit(0)
		return nil
	},
	ECHO_COMMAND: func(args []string) error {
		fmt.Println(strings.Join(args, " "))
		return nil
	},
	TYPE_COMMAND: typeCommand,
	PWD_COMMAND:  pwdCommand,
	CD_COMMAND:   cdCommand,
}

func pwdCommand(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(dir)
	return nil
}

func typeCommand(args []string) error {
	if len(args) == 0 {
		fmt.Print("")
		return nil
	}

	for _, arg := range args {

		// TODO check commands more dynamic
		if slices.Contains([]string{ECHO_COMMAND, EXIT_COMMAND, TYPE_COMMAND, PWD_COMMAND, CD_COMMAND}, arg) {
			fmt.Println(arg, "is a shell builtin")
			continue
		}

		path, err := exec.LookPath(arg)

		if err != nil {
			fmt.Printf("%s: not found\n", arg)
		} else {
			fmt.Println(arg, "is", path)
		}
	}

	return nil
}

func cdCommand(args []string) error {

	if len(args) == 0 {
		fmt.Print("")
		return nil
	}

	newDir := args[0]

	// TODO: implement also the "-" functionality
	if newDir == "~" {
		newDir = os.Getenv("HOME") // consider also os.UserHomeDir()
	}

	err := os.Chdir(newDir)

	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", newDir)
	}

	// TODO: implement a sort of shell.RunCommand, so bultin commands can change the shell in a better way
	shell.changeDir(newDir)

	return nil

	/*
		 		err := os.Chdir(argv[1])
			switch {
			case err == nil:
				return
			case errors.Is(err, fs.ErrNotExist):
				fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", argv[1])
			case errors.Is(err, fs.ErrPermission):
				fmt.Fprintf(os.Stderr, "cd: %s: Permission denied\n", argv[1])
			default:
				fmt.Fprintf(os.Stderr, "cd: %s: Not a directory\n", argv[1])
			}
	*/
}

// keeping this approach commented just for fun, what I've done before knowing the existence of exec.LookPath
/* func customLookPath(file string) string {
	PATH_ENV := os.Getenv("PATH")
	paths := strings.Split(PATH_ENV, string(os.PathListSeparator))

	for _, path := range paths {

		fullPath := path + "/" + file

		if fileInfo, err := os.Stat(fullPath); !os.IsNotExist(err) {
			if fileInfo.Mode()&0100 != 0 {
				return fullPath
			}
		}
	}
} */
