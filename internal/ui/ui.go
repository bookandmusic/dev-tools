package ui

type UI interface {
	Info(format string, a ...any)
	Success(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Note(format string, a ...any)
	Println(a ...any)
}
