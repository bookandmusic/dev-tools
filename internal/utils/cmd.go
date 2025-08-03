package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// RunCommand runs the command with real-time stdout/stderr to terminal.
func RunCommand(name string, args ...string) error {
	_, err := runCommandInternal(name, args, true, true)
	return err
}

// RunCommandCapture runs the command and returns combined stdout+stderr without printing to terminal.
func RunCommandCapture(name string, args ...string) (string, error) {
	return runCommandInternal(name, args, false, true)
}

// Internal executor
func runCommandInternal(name string, args []string, printToTerminal bool, captureOutput bool) (string, error) {
	cmd := exec.Command(name, args...)

	var buf bytes.Buffer
	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	if captureOutput && printToTerminal {
		// MultiWriter: print & capture
		stdout = io.MultiWriter(os.Stdout, &buf)
		stderr = io.MultiWriter(os.Stderr, &buf)
	} else if captureOutput && !printToTerminal {
		stdout = &buf
		stderr = &buf
	} // else: default to os.Stdout/os.Stderr

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if captureOutput {
		return buf.String(), err
	}
	return "", err
}

// RunSystemctlCommand 执行单条 systemctl 命令，阻塞等待完成
func RunSystemctlCommand(args ...string) error {
	return RunCommand("systemctl", args...)
}
