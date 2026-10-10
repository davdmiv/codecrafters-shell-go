package commands

import (
	"fmt"
	"strconv"
)

// ExitError сигнализирует шеллу, что нужно завершить REPL с данным кодом.
// Сама команда exit не вызывает os.Exit — это делает main/shell.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

type Exit struct{}

func NewExit() *Exit {
	return &Exit{}
}

func (e *Exit) Name() string {
	return "exit"
}

func (e *Exit) Execute(ctx *Context, args []string) error {
	code := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("exit: %s: numeric argument required", args[0])
		}
		code = n
	}
	return &ExitError{Code: code}
}
