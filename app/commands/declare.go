package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/vars"
)

// Declare is the shell builtin `declare`.
type Declare struct {
	store *vars.Store
}

func NewDeclare(store *vars.Store) *Declare {
	return &Declare{store: store}
}

func (d *Declare) Name() string {
	return "declare"
}

func (d *Declare) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return nil
	}

	if args[0] == "-p" {
		if len(args) < 2 {
			return nil
		}
		name := args[1]
		value, ok := d.store.Get(name)
		if !ok {
			return fmt.Errorf("declare: %s: not found", name)
		}
		_, err := fmt.Fprintf(ctx.Stdout, "declare -- %s=%q\n", name, value)
		return err
	}

	for _, arg := range args {
		name, value, hasValue := strings.Cut(arg, "=")
		if !hasValue {
			name = arg
			value = ""
		}
		if !vars.IsValidIdentifier(name) {
			return fmt.Errorf("declare: `%s': not a valid identifier", arg)
		}
		d.store.Set(name, value)
	}
	return nil
}
