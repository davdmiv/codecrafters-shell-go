package commands

import (
	"errors"
	"fmt"
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
