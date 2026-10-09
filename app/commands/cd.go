package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

type Cd struct{}

func NewCd() *Cd {
	return &Cd{}
}

func (e *Cd) Name() string {
	return "cd"
}

func (e *Cd) Execute(ctx *Context, args []string) error {
	var target string
	if len(args) > 0 {
		target = args[0]
	}

	if target == "~" || target == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not find home directory: %w", err)
		}
		target = home
	}

	path := filepath.Clean(target)

	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("cd: %s: %w\n", path, err)
	}

	return nil
}
