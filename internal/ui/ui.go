package ui

// UI 定义日志输出接口
type UI interface {
	Info(msg string, args ...interface{})
	Success(msg string, args ...interface{})
	Warning(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Println(msg string, args ...interface{})
}
