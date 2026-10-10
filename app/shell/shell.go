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

type Shell struct {
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
	registry *commands.Registry 
}

func New(stdout, stderr io.Writer, stdin io.Reader) *Shell {
	sh := &Shell{
		stdout:   stdout,
		stderr:   stderr,
		stdin:    stdin,
		registry: commands.NewRegistry(),
	}

	sh.registry.Register(commands.NewEcho())
	sh.registry.Register(commands.NewExit())
	sh.registry.Register(commands.NewPwd())
	sh.registry.Register(commands.NewCd())
	sh.registry.Register(commands.NewType(sh.registry))

	return sh
}

func (sh *Shell) Run() int {
	reader := bufio.NewReader(sh.stdin)

	for {
		fmt.Fprint(sh.stdout, "$ ")

		command, err := reader.ReadString('\n')
		if err != nil {
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

func (sh *Shell) execute(tokens []string) (code int, ok bool) {
	baseCmd := tokens[0]
	args := tokens[1:]

	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	var cmd commands.Command

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

		var statusErr *commands.StatusError
		if errors.As(err, &statusErr) {
			return 0, false
		}
		fmt.Fprintf(ctx.Stderr, "%v\n", err)
	}
	return 0, false
}
