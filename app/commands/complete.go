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
		return fmt.Errorf("complete: usage: complete [-C <script> | -r | -p] <command>")
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
	case "-p":
		return c.printSpecs(ctx, args[1:])
	default:
		return fmt.Errorf("complete: %s: invalid option", args[0])
	}
}

// printSpecs prints completion specs in bash-compatible form:
//
//	complete -C '<script>' <command>
//
// With no names, prints all registered specs (sorted by command).
func (c *Complete) printSpecs(ctx *Context, names []string) error {
	if len(names) == 0 {
		for _, command := range c.specs.Commands() {
			script, _ := c.specs.Get(command)
			fmt.Fprintf(ctx.Stdout, "complete -C %s %s\n", quoteCompleteArg(script), command)
		}
		return nil
	}

	for _, command := range names {
		script, ok := c.specs.Get(command)
		if !ok {
			return fmt.Errorf("complete: %s: no completion specification", command)
		}
		fmt.Fprintf(ctx.Stdout, "complete -C %s %s\n", quoteCompleteArg(script), command)
	}
	return nil
}

// quoteCompleteArg wraps s in single quotes the way bash `complete -p` does.
func quoteCompleteArg(s string) string {
	return "'" + s + "'"
}
