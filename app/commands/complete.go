package commands

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/completion"
)

// Complete is the shell builtin `complete` for programmable completion specs.
type Complete struct {
	specs *completion.Specs
}

func NewComplete(specs *completion.Specs) *Complete {
	return &Complete{specs: specs}
}

func (c *Complete) Name() string {
	return "complete"
}

func (c *Complete) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("complete: usage: complete -C <script> <command> | complete -r <command>")
	}

	switch args[0] {
	case "-C":
		if len(args) < 3 {
			return fmt.Errorf("complete: -C: option requires an argument")
		}
		script := args[1]
		command := args[2]
		c.specs.Set(command, script)
		return nil
	case "-r":
		if len(args) < 2 {
			return fmt.Errorf("complete: -r: option requires an argument")
		}
		c.specs.Remove(args[1])
		return nil
	default:
		return fmt.Errorf("complete: %s: invalid option", args[0])
	}
}
