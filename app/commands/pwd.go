package commands

import (
	"fmt"
	"os"
)

type Pwd struct{}

func NewPwd() *Pwd {
	return &Pwd{}
}

func (p *Pwd) Name() string {
	return "pwd"
}

func (p *Pwd) Execute(ctx *Context, args []string) error {
	output, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(ctx.Stdout, output)
	return err
}
