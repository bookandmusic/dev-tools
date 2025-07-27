package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// NewConsoleUI 创建 UI 实例
func NewConsoleUI(debug bool) UI {
	return &ConsoleUI{debugEnabled: debug}
}

var (
	timeColor    = color.New(color.FgWhite).SprintFunc() // 时间统一颜色
	infoColor    = color.New(color.FgCyan).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	debugColor   = color.New(color.FgMagenta).SprintFunc()
)

// ConsoleUI 控制台输出实现
type ConsoleUI struct {
	debugEnabled bool
}

// formatMessage 增加时间戳、对齐级别标志和颜色
func (c ConsoleUI) formatMessage(level string, colorFunc func(a ...interface{}) string, msg string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMsg := fmt.Sprintf(msg, args...)

	// 固定级别宽度 7 个字符
	levelStr := fmt.Sprintf("%-7s", level)

	return fmt.Sprintf("%s %s %s", timeColor("["+timestamp+"]"), colorFunc("["+levelStr+"]"), formattedMsg)
}

func (c ConsoleUI) Info(msg string, args ...interface{}) {
	fmt.Println(c.formatMessage("INFO", infoColor, msg, args...))
}

func (c ConsoleUI) Success(msg string, args ...interface{}) {
	fmt.Println(c.formatMessage("SUCCESS", successColor, msg, args...))
}

func (c ConsoleUI) Warning(msg string, args ...interface{}) {
	fmt.Println(c.formatMessage("WARN", warnColor, msg, args...))
}

func (c ConsoleUI) Error(msg string, args ...interface{}) {
	fmt.Println(c.formatMessage("ERROR", errorColor, msg, args...))
}

func (c ConsoleUI) Debug(msg string, args ...interface{}) {
	if c.debugEnabled {
		fmt.Println(c.formatMessage("DEBUG", debugColor, msg, args...))
	}
}

func (c ConsoleUI) Println(msg string, args ...interface{}) {
	// 直接用Printf格式化并加换行
	fmt.Println(fmt.Sprintf(msg, args...))
}
