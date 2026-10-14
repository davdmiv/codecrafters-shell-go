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
	"github.com/codecrafters-io/shell-starter-go/app/completion"
	"github.com/codecrafters-io/shell-starter-go/app/jobs"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

type Shell struct {
	stdout      io.Writer
	stderr      io.Writer
	stdin       io.Reader
	registry    *commands.Registry
	specs       *completion.Specs
	jobs        *jobs.Manager
	interactive bool // raw TTY: echo keystrokes ourselves
	tabPending  bool // first Tab on ambiguous match already rang the bell
}

func New(stdout, stderr io.Writer, stdin io.Reader) *Shell {
	specs := completion.NewSpecs()
	jobMgr := jobs.NewManager()
	sh := &Shell{
		stdout:   newSyncWriter(stdout),
		stderr:   newSyncWriter(stderr),
		stdin:    stdin,
		registry: commands.NewRegistry(),
		specs:    specs,
		jobs:     jobMgr,
	}

	sh.registry.Register(commands.NewEcho())
	sh.registry.Register(commands.NewExit())
	sh.registry.Register(commands.NewPwd())
	sh.registry.Register(commands.NewCd())
	sh.registry.Register(commands.NewType(sh.registry))
	sh.registry.Register(commands.NewComplete(specs))
	sh.registry.Register(commands.NewJobs(jobMgr))

	return sh
}

func (sh *Shell) Run() int {
	reader := bufio.NewReader(sh.stdin)

	for {
		// Notify about finished background jobs before the next prompt.
		sh.jobs.NotifyFinished(sh.stderr)

		// raw → readline → restore → execute:
		// во время команды терминал снова «cooked», иначе \n от процессов
		// не возвращает курсор в колонку 0 и ломает следующий prompt.
		restore := sh.enableRawMode()
		fmt.Fprint(sh.stdout, "$ ")
		command, err := sh.readLine(reader)
		restore()

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
		if code, ok := sh.execute(cmd, command); ok {
			return code
		}
	}
}

func (sh *Shell) execute(command *parser.Command, rawLine string) (code int, ok bool) {
	baseCmd := command.Name
	args := command.Args

	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	opened := make([]io.Closer, 0, len(command.Redirections))
	defer func() {
		for _, c := range opened {
			_ = c.Close()
		}
	}()

	for _, redir := range command.Redirections {
		switch redir.Type {
		case parser.TokenRedirectOut:
			file, err := os.Create(redir.FilePath)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			opened = append(opened, file)
			ctx.Stdout = file
		case parser.TokenRedirectAppend:
			file, err := os.OpenFile(redir.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			opened = append(opened, file)
			ctx.Stdout = file
		case parser.TokenRedirectErr:
			file, err := os.Create(redir.FilePath)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			opened = append(opened, file)
			ctx.Stderr = file
		case parser.TokenRedirectAppendErr:
			file, err := os.OpenFile(redir.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			opened = append(opened, file)
			ctx.Stderr = file
		case parser.TokenRedirectIn:
			file, err := os.Open(redir.FilePath)
			if err != nil {
				fmt.Fprintf(sh.stderr, "Error: %v\n", err)
				return 0, false
			}
			opened = append(opened, file)
			ctx.Stdin = file
		}
	}

	if command.Background {
		return sh.executeBackground(baseCmd, args, ctx, rawLine)
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

func (sh *Shell) executeBackground(baseCmd string, args []string, ctx *commands.Context, rawLine string) (code int, ok bool) {
	// Builtins in the background are not supported; run only external programs.
	if _, isBuiltin := sh.registry.Get(baseCmd); isBuiltin {
		fmt.Fprintf(ctx.Stderr, "%s: background jobs for builtins are not supported\n", baseCmd)
		return 0, false
	}
	if _, err := exec.LookPath(baseCmd); err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: command not found\n", baseCmd)
		return 0, false
	}

	proc, err := commands.StartExternal(baseCmd, args, ctx, true)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%v\n", err)
		return 0, false
	}

	job := sh.jobs.Add(proc, rawLine)
	fmt.Fprintf(sh.stdout, "[%d] %d\n", job.ID, job.PID)
	return 0, false
}
