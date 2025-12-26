package shell

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"

	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

const PROMP = "$ "

const (
	EXIT_COMMAND = "exit"
	ECHO_COMMAND = "echo"
	TYPE_COMMAND = "type"
	PWD_COMMAND  = "pwd"
	CD_COMMAND   = "cd"
)

type Shell struct {
	CurrentDir string
}

func (s *Shell) changeDir(newDir string) {
	s.CurrentDir = newDir
}

func (s *Shell) PrintPrompt() {
	// fmt.Print(s.currentDir, PROMP)
	fmt.Print(PROMP)
}

func (s *Shell) isBuiltInCommand(cmd string) bool {
	return slices.Contains([]string{
		ECHO_COMMAND,
		EXIT_COMMAND,
		TYPE_COMMAND,
		PWD_COMMAND,
		CD_COMMAND,
	}, cmd)
}

func (s *Shell) getBuiltins() map[string]func([]string) {

	builtin := map[string]func([]string){
		ECHO_COMMAND: s.echo,
		EXIT_COMMAND: func([]string) { s.exit() },
		TYPE_COMMAND: s.typeCmd,
		PWD_COMMAND:  s.pwd,
		CD_COMMAND:   s.cd,
	}

	return builtin
}

func (s *Shell) echo(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func (s *Shell) exit() {
	os.Exit(0)
}

func (s *Shell) typeCmd(args []string) {
	if len(args) == 0 {
		fmt.Print("")
	}

	for _, arg := range args {

		// TODO check commands more dynamic
		if s.isBuiltInCommand(arg) {
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
}

func (s *Shell) pwd(args []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(dir)
}

func (s *Shell) cd(args []string) {

	var newDir string

	if len(args) == 0 {
		newDir = os.Getenv("HOME")
	}

	newDir = args[0]

	// TODO: implement also the "-" functionality
	if newDir == "~" {
		newDir = os.Getenv("HOME") // consider also os.UserHomeDir()
	}

	err := os.Chdir(newDir)

	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", newDir)
	}

	s.changeDir(newDir)
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

func (s *Shell) Execute(sc *parser.SingleCommand) error {
	restoreFns, err := s.applyRedirections(sc.Redirs)
	if err != nil {
		return err
	}

	// Always restore FDs
	defer func() {
		for i := len(restoreFns) - 1; i >= 0; i-- {
			restoreFns[i]()
		}
	}()

	// No command, only redirections (valid shell behavior)
	if len(sc.Args) == 0 {
		return nil
	}

	cmd := sc.Args[0]
	args := sc.Args[1:]

	if builtin, ok := s.getBuiltins()[cmd]; ok {
		builtin(args)
		return nil
	}

	return s.execExternal(cmd, args)
}

func (s *Shell) execExternal(cmd string, args []string) error {

	// search the command in PATH
	_, err := exec.LookPath(cmd)

	if err != nil {
		return fmt.Errorf("%s: command not found", cmd)
	}

	command := exec.Command(cmd, args...)

	command.Stdout = os.Stdout
	//command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("execution error: %s", err)
		}
	}

	return nil
}

func (s *Shell) applyRedirections(redirs []parser.Redirection) ([]func(), error) {
	var restoreFns []func()

	for _, r := range redirs {

		// Save original FD
		origFD, err := syscall.Dup(r.FD)
		if err != nil {
			return nil, err
		}

		restoreFns = append(restoreFns, func(fd, saved int) func() {
			return func() {
				syscall.Dup2(saved, fd)
				syscall.Close(saved)
			}
		}(r.FD, origFD))

		flags := os.O_CREATE | os.O_WRONLY

		if r.Append {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		file, err := os.OpenFile(r.To, flags, 0644)
		if err != nil {
			return nil, err
		}

		// Attach file to target FD
		if err := syscall.Dup2(int(file.Fd()), r.FD); err != nil {
			file.Close()
			return nil, err
		}

		// Close original FD (dup2 keeps it alive)
		file.Close()
	}

	return restoreFns, nil
}
