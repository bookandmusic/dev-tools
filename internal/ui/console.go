package ui

import (
	"fmt"

	"github.com/fatih/color"
)

type ConsoleUI struct{}

var (
	infoColor    = color.New(color.FgCyan).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	noteColor    = color.New(color.FgWhite).Add(color.Italic).SprintFunc()
)

func (c ConsoleUI) Info(format string, a ...any) {
	fmt.Println(infoColor("[INFO] "), fmt.Sprintf(format, a...))
}

func (c ConsoleUI) Success(format string, a ...any) {
	fmt.Println(successColor("[ OK ] "), fmt.Sprintf(format, a...))
}

func (c ConsoleUI) Warning(format string, a ...any) {
	fmt.Println(warnColor("[WARN] "), fmt.Sprintf(format, a...))
}

func (c ConsoleUI) Error(format string, a ...any) {
	fmt.Println(errorColor("[FAIL] "), fmt.Sprintf(format, a...))
}

func (c ConsoleUI) Note(format string, a ...any) {
	fmt.Println(noteColor("[NOTE] "), fmt.Sprintf(format, a...))
}

func (c ConsoleUI) Println(a ...any) {
	fmt.Println(a...)
}
