package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

		cmd, err := parser.Parse(command)
		if err != nil {
			fmt.Fprintf(sh.stderr, "Parse error: %v\n", err)
			continue
		}
		if code, ok := sh.execute(cmd); ok {
			return code
		}
	}
}

func (sh *Shell) execute(command *parser.Command) (code int, ok bool) {
	baseCmd := command.Name
	args := command.Args

	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	for _, redir := range command.Redirections {
		switch redir.Type {
		case parser.TokenRedirectOut:
			file, err := os.Create(redir.FilePath)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			defer file.Close()
			ctx.Stdout = file
		case parser.TokenRedirectAppend:
			file, err := os.OpenFile(redir.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			defer file.Close()
			ctx.Stdout = file
		case parser.TokenRedirectIn:
			file, err := os.Open(redir.FilePath)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			defer file.Close()
			ctx.Stdin = file
		}
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
