package commands

import (
	"io"
)

// Context описывает окружение, в котором выполняется команда.
// Это позволяет командам не зависеть напрямую от os.Stdout/os.Stdin,
// что критически важно для написания юнит-тестов.
type Context struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	// Сюда позже можно добавить состояние шелла:
	// переменные окружения, историю команд, алиасы и т.д.
}

// Command - интерфейс, который должна реализовать любая встроенная команда.
type Command interface {
	// Name возвращает имя команды (например, "echo", "cd")
	Name() string
	// Execute выполняет команду.
	// ВАЖНО: args здесь НЕ включают само имя команды (только аргументы).
	Execute(ctx *Context, args []string) error
}

// Registry хранит все зарегистрированные встроенные команды.
type Registry struct {
	commands map[string]Command
}

// NewRegistry создает новый пустой реестр.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register добавляет команду в реестр.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get ищет команду по имени.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}
