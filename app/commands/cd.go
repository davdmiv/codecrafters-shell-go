package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Cd struct{}

func NewCd() *Cd {
	return &Cd{}
}

func (c *Cd) Name() string {
	return "cd"
}

func (c *Cd) Execute(ctx *Context, args []string) error {
	var target string
	if len(args) > 0 {
		target = args[0]
	}

	if target == "" || target == "~" || strings.HasPrefix(target, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not find home directory: %w", err)
		}
		if target == "" || target == "~" {
			target = home
		} else {
			target = filepath.Join(home, target[2:])
		}
	}

	path := filepath.Clean(target)

	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("%s: No such file or directory", path)
	}

	return nil
}
