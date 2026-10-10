package commands

import (
	"fmt"
	"strings"
)

type Echo struct{}

func NewEcho() *Echo {
	return &Echo{}
}

func (e *Echo) Name() string {
	return "echo"
}

func (e *Echo) Execute(ctx *Context, args []string) error {
	output := strings.Join(args, " ")
	_, err := fmt.Fprintln(ctx.Stdout, output)
	return err
}
