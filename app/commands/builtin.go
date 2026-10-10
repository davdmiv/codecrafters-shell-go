package commands

import (
	"io"
)

type Context struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	// Сюда позже можно добавить состояние шелла:
	// переменные окружения, историю команд, алиасы и т.д.
}

type Command interface {
	Name() string

	// ВАЖНО: args здесь НЕ включают само имя команды (только аргументы).
	Execute(ctx *Context, args []string) error
}

type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}
