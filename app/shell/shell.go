package shell

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/commands"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

// Shell инкапсулирует состояние и логику работы оболочки.
type Shell struct {
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
	registry *commands.Registry // Реестр команд из прошлого примера
}

// New - конструктор шелла.
func New(stdout, stderr io.Writer, stdin io.Reader) *Shell {
	sh := &Shell{
		stdout:   stdout,
		stderr:   stderr,
		stdin:    stdin,
		registry: commands.NewRegistry(),
	}

	// Регистрируем встроенные команды (используем реализации из echo.go, cd.go и т.д.)
	sh.registry.Register(commands.NewEcho())
	sh.registry.Register(commands.NewExit())
	sh.registry.Register(commands.NewPwd())
	sh.registry.Register(commands.NewCd())
	sh.registry.Register(commands.NewType(sh.registry))

	return sh
}

// Run запускает бесконечный цикл REPL (Read-Eval-Print Loop).
func (sh *Shell) Run() {
	reader := bufio.NewReader(sh.stdin)

	for {
		// Используем Fprint, чтобы писать в наш io.Writer, а не жестко в os.Stdout
		fmt.Fprint(sh.stdout, "$ ")

		command, err := reader.ReadString('\n')
		if err != nil {
			// Обработка EOF (например, Ctrl+D)
			if err == io.EOF {
				fmt.Fprintln(sh.stdout, "\nexit")
				return
			}
			fmt.Fprintf(sh.stderr, "Error while reading: %v\n", err)
			continue
		}

		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		tokens, err := parser.SplitArgs(command)
		if err != nil {
			fmt.Fprintf(sh.stderr, "Parse error: %v\n", err)
			continue
		}
		if len(tokens) == 0 {
			continue
		}

		// Запускаем выполнение команды
		sh.execute(tokens)
	}
}

// execute определяет, как запустить команду (встроенная, внешняя или не найдена).
func (sh *Shell) execute(tokens []string) {
	baseCmd := tokens[0]
	args := tokens[1:]

	// Создаем контекст выполнения для текущей команды
	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	// 1. Проверяем встроенные команды
	if cmd, ok := sh.registry.Get(baseCmd); ok {
		if err := cmd.Execute(ctx, args); err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %v\n", baseCmd, err)
		}
		return
	}

	// 2. Проверяем внешние команды
	if _, err := exec.LookPath(baseCmd); err == nil {
		extCmd := exec.Command(baseCmd, args...)
		extCmd.Stdout = ctx.Stdout
		extCmd.Stderr = ctx.Stderr
		extCmd.Stdin = ctx.Stdin

		if err := extCmd.Run(); err != nil {
			// Внешние команды сами печатают свои ошибки в stderr,
			// но если процесс не запустился, сообщим об этом.
			fmt.Fprintf(ctx.Stderr, "%v\n", err)
		}
		return
	}

	// 3. Команда не найдена
	fmt.Fprintf(ctx.Stderr, "%s: command not found\n", baseCmd)
}
