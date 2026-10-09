package commands

import (
	"fmt"
	"os/exec"
)

// External реализует интерфейс Command для внешних программ.
type External struct {
	command string // Абсолютный путь к исполняемому файлу (например, "/bin/ls")
}

// NewExternal создает новую внешнюю команду.
func NewExternal(command string) *External {
	return &External{command: command}
}

func (e *External) Name() string {
	return e.command
}

// Execute запускает внешний процесс.
func (e *External) Execute(ctx *Context, args []string) error {
	// Создаем команду, передавая ей путь и аргументы
	cmd := exec.Command(e.command, args...)

	// ВАЖНО: пробрасываем потоки ввода-вывода из нашего контекста
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin

	// Запускаем и ждем завершения
	if err := cmd.Run(); err != nil {
		// Внешние команды обычно сами пишут ошибки в свой stderr,
		// но если процесс вообще не удалось запустить, вернем ошибку.
		return fmt.Errorf("failed to execute %s: %w", e.command, err)
	}

	return nil
}
