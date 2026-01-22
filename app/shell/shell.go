package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/tomasdepi/golang-shell/app/parser"
)

const PROMP = "$ "

const (
	EXIT_COMMAND    = "exit"
	ECHO_COMMAND    = "echo"
	TYPE_COMMAND    = "type"
	PWD_COMMAND     = "pwd"
	CD_COMMAND      = "cd"
	HISTORY_COMMAND = "history"
)

type Command interface {
	Run() error
	SetStdout(*os.File)
}

type ExternalCommand struct {
	cmd *exec.Cmd
	ctx *ExecutionContext
}

type BuiltinCommand struct {
	builtin func(io.Writer, []string)
	args    []string
	ctx     *ExecutionContext
}

func (ec *ExternalCommand) Run() error {
	ec.cmd.Stdout = ec.ctx.stdout
	ec.cmd.Stdin = ec.ctx.stdin
	ec.cmd.Start()

	return nil
}

func (ec *ExternalCommand) Wait() error {
	return ec.cmd.Wait()
}

func (ec *ExternalCommand) SetStdout(out *os.File) {
	ec.ctx.stdout = out
}

func (bc *BuiltinCommand) Run() error {
	bc.builtin(bc.ctx.stdout, bc.args)

	if bc.ctx.stdout != os.Stdout {
		bc.ctx.stdout.Close()
	}

	return nil
}

func (bc *BuiltinCommand) SetStdout(out *os.File) {
	bc.ctx.stdout = out
}

type ExecutionContext struct {
	stdout *os.File
	stdin  *os.File
	stderr *os.File
}

type Pipe struct {
	r *os.File
	w *os.File
}

type Shell struct {
	CurrentDir     string
	HistoryManager *HistoryManager
	readline       *ReadLine
}

func NewShell(currDir string) *Shell {

	h := NewHistory()

	// TODO: quite ugly, need refactor
	rl := NewReadLine(PROMP, h, []string{
		ECHO_COMMAND,
		EXIT_COMMAND,
		TYPE_COMMAND,
		PWD_COMMAND,
		CD_COMMAND,
		HISTORY_COMMAND,
	})

	return &Shell{
		CurrentDir:     currDir,
		HistoryManager: h,
		readline:       rl,
	}
}

func (s *Shell) ReadlineFromShell() string {
	return s.readline.Readline()
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
		HISTORY_COMMAND,
	}, cmd)
}

func (s *Shell) getBuiltins() map[string]func(io.Writer, []string) {

	builtin := map[string]func(io.Writer, []string){
		ECHO_COMMAND:    s.echo,
		EXIT_COMMAND:    s.exit,
		TYPE_COMMAND:    s.typeCmd,
		PWD_COMMAND:     s.pwd,
		CD_COMMAND:      s.cd,
		HISTORY_COMMAND: s.history,
	}

	return builtin
}

func (s *Shell) echo(out io.Writer, args []string) {
	fmt.Fprintln(out, strings.Join(args, " "))
}

func (s *Shell) exit(out io.Writer, args []string) {

	s.HistoryManager.SaveToFile(os.Getenv("HISTFILE"))
	os.Exit(0)
}

func (s *Shell) typeCmd(out io.Writer, args []string) {
	if len(args) == 0 {
		fmt.Fprint(out, "")
	}

	for _, arg := range args {

		// TODO check commands more dynamic
		if s.isBuiltInCommand(arg) {
			fmt.Fprintln(out, arg, "is a shell builtin")
			continue
		}

		path, err := exec.LookPath(arg)

		if err != nil {
			fmt.Fprintf(out, "%s: not found\n", arg)
		} else {
			fmt.Fprintln(out, arg, "is", path)
		}
	}
}

func (s *Shell) pwd(out io.Writer, args []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintln(out, dir)
}

func (s *Shell) cd(out io.Writer, args []string) {

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
		fmt.Fprintf(out, "cd: %s: No such file or directory\n", newDir)
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

func (s *Shell) history(out io.Writer, args []string) {

	var history []string

	if len(args) == 0 {
		history = s.HistoryManager.ReadAll()
		for i, entry := range history {
			fmt.Fprintln(out, "\t", i+1, entry)
		}
		return
	}

	param := args[0]

	n, err := strconv.Atoi(param)

	if err == nil { // param is valid number
		history = s.HistoryManager.ReadLastN(n)
		countFrom := s.HistoryManager.GetHistoryLen() - n + 1

		for i, entry := range history {
			fmt.Println("\t", countFrom+i, entry)
		}
		return
	}

	switch param {
	case "-a":
		err := s.HistoryManager.AppendToFile(args[1])
		if err != nil {
			fmt.Println(err)
		}
		return
	case "-w":
		err := s.HistoryManager.SaveToFile(args[1])
		if err != nil {
			fmt.Println(err)
		}
		return
	case "-r":
		err := s.HistoryManager.LoadFromFile(args[1])
		if err != nil {
			fmt.Println(err)
		}
	}

}

func (s *Shell) Execute(pipeline *parser.Pipeline) error {

	if len(pipeline.Commands) == 0 {
		return nil
	}

	if len(pipeline.Commands) == 1 {
		return s.excuteSingleCommand(pipeline.Commands[0])
	}

	pipes := []Pipe{}

	c := make([]Command, len(pipeline.Commands))

	if s.isBuiltInCommand(pipeline.Commands[0].Args[0]) {
		c[0] = &BuiltinCommand{
			ctx: &ExecutionContext{
				stdout: os.Stdout,
				stdin:  os.Stdin,
				stderr: os.Stderr,
			},
			builtin: s.getBuiltins()[pipeline.Commands[0].Args[0]],
			args:    pipeline.Commands[0].Args[1:],
		}
	} else {
		c[0] = &ExternalCommand{
			ctx: &ExecutionContext{
				stdout: os.Stdout,
				stdin:  os.Stdin,
				stderr: os.Stderr,
			},
			cmd: exec.Command(pipeline.Commands[0].Args[0], pipeline.Commands[0].Args[1:]...),
		}
	}

	for i, cmd := range pipeline.Commands[1:] {

		index := i + 1

		r, w, err := os.Pipe()
		if err != nil {
			panic(err)
		}

		pipes = append(pipes, Pipe{w: w, r: r})

		c[index-1].SetStdout(w)

		if s.isBuiltInCommand(cmd.Args[0]) {
			c[index] = &BuiltinCommand{
				ctx: &ExecutionContext{
					stdout: os.Stdout,
					stdin:  r,
					stderr: os.Stderr,
				},
				builtin: s.getBuiltins()[cmd.Args[0]],
				args:    cmd.Args[1:],
			}
		} else {
			c[index] = &ExternalCommand{
				ctx: &ExecutionContext{
					stdout: os.Stdout,
					stdin:  r,
					stderr: os.Stderr,
				},
				cmd: exec.Command(cmd.Args[0], cmd.Args[1:]...),
			}
		}
	}

	for _, e := range c {
		e.Run()
	}

	for _, pipe := range pipes {
		pipe.r.Close()
		pipe.w.Close()
	}

	// Wait all
	for _, cmd := range c {
		if ec, ok := cmd.(*ExternalCommand); ok {
			ec.Wait()
		}
	}

	return nil
}

func (s *Shell) excuteSingleCommand(sc *parser.SingleCommand) error {
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

	s.HistoryManager.Add(strings.Join(sc.Args, " "))

	cmd := sc.Args[0]
	args := sc.Args[1:]

	if builtin, ok := s.getBuiltins()[cmd]; ok {
		builtin(os.Stdout, args)
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
