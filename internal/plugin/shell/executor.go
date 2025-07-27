package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
)

type ShellExecutor struct{}

func (s *ShellExecutor) ScriptPath(basePath string, names ...string) string {
	return filepath.Join(append([]string{basePath}, names...)...) + ".sh"
}

func (s *ShellExecutor) Exec(scriptPath string, c *cli.Context) error {
	args := []string{scriptPath}

	// Add all flags
	for _, name := range c.LocalFlagNames() {
		if c.IsSet(name) {
			args = append(args, "--"+name, c.String(name))
		}
	}
	// Add raw args
	args = append(args, c.Args().Slice()...)

	cmd := exec.Command("bash", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func (s *ShellExecutor) NotFoundError(path string) error {
	return fmt.Errorf("script not found: %s", path)
}
