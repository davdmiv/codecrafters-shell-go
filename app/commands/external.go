package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

type External struct {
	command string
}

func NewExternal(command string) *External {
	return &External{command: command}
}

func (e *External) Name() string {
	return e.command
}

func (e *External) Execute(ctx *Context, args []string) error {
	cmd := exec.Command(e.command, args...)

	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &StatusError{Code: exitErr.ExitCode()}
		}
		// Процесс вообще не удалось запустить (нет файла, permission denied и т.п.)
		return fmt.Errorf("failed to execute %s: %w", e.command, err)
	}

	return nil
}

// StartExternal starts an external command without waiting for it.
// When background is true, stdin is connected to /dev/null.
func StartExternal(name string, args []string, ctx *Context, background bool) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if background {
		devNull, err := os.Open(os.DevNull)
		if err != nil {
			return nil, err
		}
		cmd.Stdin = devNull
	} else {
		cmd.Stdin = ctx.Stdin
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to execute %s: %w", name, err)
	}
	return cmd, nil
}
