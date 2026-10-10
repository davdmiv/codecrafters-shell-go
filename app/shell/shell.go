package shell

import (
	"bufio"
	"errors"
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

// Run запускает бесконечный цикл REPL и возвращает код выхода процесса.
func (sh *Shell) Run() int {
	reader := bufio.NewReader(sh.stdin)

	for {
		// Используем Fprint, чтобы писать в наш io.Writer, а не жестко в os.Stdout
		fmt.Fprint(sh.stdout, "$ ")

		command, err := reader.ReadString('\n')
		if err != nil {
			// Обработка EOF (например, Ctrl+D)
			if err == io.EOF {
				fmt.Fprintln(sh.stdout, "\nexit")
				return 0
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

		if code, ok := sh.execute(tokens); ok {
			return code
		}
	}
}

// execute запускает команду. ok=true означает, что шелл должен завершиться.
func (sh *Shell) execute(tokens []string) (code int, ok bool) {
	baseCmd := tokens[0]
	args := tokens[1:]

	// Создаем контекст выполнения для текущей команды
	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	var cmd commands.Command

	// 1. Проверяем встроенные команды
	if builtin, found := sh.registry.Get(baseCmd); found {
		cmd = builtin
	} else {
		if _, err := exec.LookPath(baseCmd); err == nil {
			cmd = commands.NewExternal(baseCmd)
		} else {
			fmt.Fprintf(ctx.Stderr, "%s: command not found\n", baseCmd)
			return 0, false
		}
	}

	if err := cmd.Execute(ctx, args); err != nil {
		var exitErr *commands.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code, true
		}
		fmt.Fprintf(ctx.Stderr, "%v\n", err)
	}
	return 0, false
}
