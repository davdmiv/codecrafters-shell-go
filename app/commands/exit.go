package commands

import (
	"os"
)

type Exit struct{}

func NewExit() *Exit {
	return &Exit{}
}

func (e *Exit) Name() string {
	return "exit"
}

func (e *Exit) Execute(ctx *Context, args []string) error {
	os.Exit(0)
	return nil
}
