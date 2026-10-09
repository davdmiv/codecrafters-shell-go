package commands

import (
	"fmt"
	"strings"
)

// Echo реализует встроенную команду echo.
type Echo struct{}

// NewEcho - конструктор (опционально, но это хороший тон в Go).
func NewEcho() *Echo {
	return &Echo{}
}

func (e *Echo) Name() string {
	return "echo"
}

// Execute выводит аргументы в стандартный поток вывода.
// Обратите внимание: нам больше не нужно делать args[1:],
// так как вызывающий код (парсер/шелл) уже отрезал имя команды.
func (e *Echo) Execute(ctx *Context, args []string) error {
	output := strings.Join(args, " ")
	_, err := fmt.Fprintln(ctx.Stdout, output)
	return err
}
