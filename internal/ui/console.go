package ui

import (
	"fmt"

	"github.com/fatih/color"
)

func NewConsoleUI(debug bool) UI {
	return &ConsoleUI{debugEnabled: debug}
}

var (
	infoColor    = color.New(color.FgCyan).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	noteColor    = color.New(color.FgWhite).Add(color.Italic).SprintFunc()
)

// ConsoleUI 控制台输出实现
type ConsoleUI struct {
	debugEnabled bool
}

func (c ConsoleUI) Info(msg string, args ...interface{}) {
	fmt.Printf(infoColor("ℹ️ "+msg+"\n"), args...)
}

func (c ConsoleUI) Success(msg string, args ...interface{}) {
	fmt.Printf(successColor("⚠️ "+msg+"\n"), args...)
}

func (c ConsoleUI) Warning(msg string, args ...interface{}) {
	fmt.Printf(warnColor("⚠️ "+msg+"\n"), args...)
}

func (c ConsoleUI) Error(msg string, args ...interface{}) {
	fmt.Printf(errorColor("❌ "+msg+"\n"), args...)
}

func (c ConsoleUI) Debug(msg string, args ...interface{}) {
	if c.debugEnabled {
		fmt.Printf(noteColor("🐛 "+msg+"\n"), args...)
	}
}

func (c ConsoleUI) Println(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}
