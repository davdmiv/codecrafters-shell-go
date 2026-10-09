package commands

import (
	"fmt"
	"os/exec"
)

type Type struct {
	registry *Registry
}

func NewType(registry *Registry) *Type {
	return &Type{
		registry: registry,
	}
}

func (t *Type) Name() string {
	return "type"
}

func (t *Type) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("type: missing argument")
	}

	target := args[0]

	if _, ok := t.registry.Get(target); ok {
		fmt.Fprintf(ctx.Stdout, "%s is a shell builtin\n", target)
		return nil
	}

	if path, err := exec.LookPath(target); err == nil {
		fmt.Fprintf(ctx.Stdout, "%s is %s\n", target, path)
		return nil
	}

	return fmt.Errorf("type: %s: not found", target)
}
