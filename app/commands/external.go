package commands

import (
	"fmt"
	"os/exec"
)

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
		return fmt.Errorf("failed to execute %s: %w", e.command, err)
	}

	return nil
}
