package shell

import (
	"fmt"
	"os"
	"slices"
)

const (
	RETURN_LINE     = "\r\n"
	DELETE_BYTE     = "\x1b[D\x1b[P"
	CURSOR_TO_LEFT  = "\x1b[D"
	CURSOR_TO_RIGHT = "\x1b[C"
	CLEAR_LINE      = "\x1b[2K"
)

type ReadLine struct {
	cursorPos int
	buff      []byte
	prompt    string
	history   History
}

func (rl *ReadLine) Print() {
	fmt.Print(string(rl.buff))
}

func (rl *ReadLine) AddPrintable(b byte) {
	rl.cursorPos++
	rl.buff = append(rl.buff, b)
}

func (rl *ReadLine) Reset() {
	rl.buff = []byte(rl.prompt)
	rl.cursorPos = len(rl.prompt)
}

func NewReadLine(prompt string, history History) *ReadLine {
	return &ReadLine{
		prompt:    prompt,
		buff:      []byte{},
		cursorPos: 0,
		history:   history,
	}
}

func (rl *ReadLine) setBuffer(s string) {
	rl.buff = []byte(s)
	rl.cursorPos = len(rl.buff)
}

func (rl *ReadLine) redraw() {
	fmt.Print("\r")
	fmt.Print(CLEAR_LINE)
	fmt.Print(rl.prompt)
	fmt.Print(string(rl.buff))

	moveLeft := len(rl.buff) - rl.cursorPos
	if moveLeft > 0 {
		fmt.Printf("\x1b[%dD", moveLeft)
	}
}

func (rl *ReadLine) Readline() string {

	rl.redraw()

	byteBuf := make([]byte, 1)

	for {
		os.Stdin.Read(byteBuf)

		input := byteBuf[0]

		if isPrintable(input) {

			if rl.cursorPos == len(rl.buff) {
				// append case (fast path)
				rl.buff = append(rl.buff, input)
				rl.cursorPos++
				fmt.Print(string(input))
			} else {
				// insert in the middle
				rl.buff = append(
					rl.buff[:rl.cursorPos],
					append([]byte{input}, rl.buff[rl.cursorPos:]...)...,
				)
				rl.cursorPos++

				rl.redraw()

			}

		}

		if input == '\r' || input == '\n' {

			fmt.Print("\r\n")

			line := string(rl.buff)

			// responsibilty of adding the entry to history is shell's
			/* if len(line) > 0 {
				rl.history.Add(line)
			} */

			rl.buff = nil
			rl.cursorPos = 0

			return line
		}

		if input == 127 { //backspace

			if rl.cursorPos > len(rl.prompt) {

				rl.cursorPos--

				rl.buff = slices.Delete(rl.buff, rl.cursorPos, rl.cursorPos+1)

				fmt.Print(DELETE_BYTE)
			}

		}

		if input == 27 { // ESC

			os.Stdin.Read(byteBuf)

			if byteBuf[0] == '[' {
				os.Stdin.Read(byteBuf)
				switch byteBuf[0] {
				case 'A': // UP
					cmd, ok := rl.history.Prev()
					if ok {
						rl.setBuffer(cmd)
						rl.redraw()
					}
				case 'B': // DOWN
					if cmd, ok := rl.history.Next(); ok {
						rl.setBuffer(cmd)
					} else {
						rl.setBuffer("")
					}
					rl.redraw()
				case 'C':
					if rl.cursorPos < len(rl.buff) {
						rl.cursorPos++
						fmt.Print(CURSOR_TO_RIGHT)
					}
				case 'D':
					if rl.cursorPos > len(rl.prompt) {
						rl.cursorPos--
						fmt.Print(CURSOR_TO_LEFT)
					}
				}
			}

		}
	}

}

func isPrintable(b byte) bool {
	return b >= 32 && b <= 126
}
