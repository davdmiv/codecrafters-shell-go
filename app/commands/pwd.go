package commands

import (
	"fmt"
	"os"
)

type Pwd struct{}

func NewPwd() *Pwd {
	return &Pwd{}
}

func (e *Pwd) Name() string {
	return "pwd"
}

func (e *Pwd) Execute(ctx *Context, args []string) error {
	output, err := os.Getwd()
	_, err = fmt.Fprintln(ctx.Stdout, output)
	return err
}
