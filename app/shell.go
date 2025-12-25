package main

import "fmt"

const PROMP = "$ "

type Shell struct {
	CurrentDir string
}

func (s *Shell) changeDir(newDir string) {
	s.CurrentDir = newDir
}

func (s *Shell) printPrompt() {
	// fmt.Print(s.currentDir, PROMP)
	fmt.Print(PROMP)
}
