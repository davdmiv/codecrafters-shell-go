package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/codecrafters-io/shell-starter-go/app/commands"
	"github.com/codecrafters-io/shell-starter-go/app/completion"
	"github.com/codecrafters-io/shell-starter-go/app/history"
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
	history     *history.History
	interactive bool // raw TTY: echo keystrokes ourselves
	tabPending  bool // first Tab on ambiguous match already rang the bell
}

func New(stdout, stderr io.Writer, stdin io.Reader) *Shell {
	specs := completion.NewSpecs()
	jobMgr := jobs.NewManager()
	hist := history.New()
	sh := &Shell{
		stdout:   newSyncWriter(stdout),
		stderr:   newSyncWriter(stderr),
		stdin:    stdin,
		registry: commands.NewRegistry(),
		specs:    specs,
		jobs:     jobMgr,
		history:  hist,
	}

	sh.registry.Register(commands.NewEcho())
	sh.registry.Register(commands.NewExit())
	sh.registry.Register(commands.NewPwd())
	sh.registry.Register(commands.NewCd())
	sh.registry.Register(commands.NewType(sh.registry))
	sh.registry.Register(commands.NewComplete(specs))
	sh.registry.Register(commands.NewJobs(jobMgr))
	sh.registry.Register(commands.NewHistory(hist))

	return sh
}

func (sh *Shell) Run() int {
	sh.loadHistFile()
	defer sh.saveHistFile()

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

		sh.history.Add(command)

		pipeline, err := parser.Parse(command)
		if err != nil {
			fmt.Fprintf(sh.stderr, "Parse error: %v\n", err)
			continue
		}
		if code, ok := sh.execute(pipeline, command); ok {
			return code
		}
	}
}

// loadHistFile loads HISTFILE into memory on startup (missing file is OK).
func (sh *Shell) loadHistFile() {
	path := os.Getenv("HISTFILE")
	if path == "" {
		return
	}
	if err := sh.history.ReadFile(path); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(sh.stderr, "history: %v\n", err)
		}
		return
	}
	// Already on disk — don't re-append them on exit.
	sh.history.MarkSaved()
}

// saveHistFile appends new in-memory entries to HISTFILE on exit.
func (sh *Shell) saveHistFile() {
	path := os.Getenv("HISTFILE")
	if path == "" {
		return
	}
	if err := sh.history.AppendFile(path); err != nil {
		fmt.Fprintf(sh.stderr, "history: %v\n", err)
	}
}

func (sh *Shell) execute(pipeline *parser.Pipeline, rawLine string) (code int, ok bool) {
	if len(pipeline.Commands) == 0 {
		return 0, false
	}
	if len(pipeline.Commands) == 1 {
		return sh.executeSingle(pipeline.Commands[0], pipeline.Background, rawLine)
	}
	if pipeline.Background {
		fmt.Fprintln(sh.stderr, "background pipelines are not supported")
		return 0, false
	}
	return sh.executePipeline(pipeline)
}

func (sh *Shell) executeSingle(command *parser.Command, background bool, rawLine string) (code int, ok bool) {
	baseCmd := command.Name
	args := command.Args

	ctx := &commands.Context{
		Stdout: sh.stdout,
		Stderr: sh.stderr,
		Stdin:  sh.stdin,
	}

	opened, err := sh.applyRedirections(ctx, command.Redirections)
	if err != nil {
		fmt.Fprintf(sh.stderr, "Error: %v\n", err)
		return 0, false
	}
	defer func() {
		for _, c := range opened {
			_ = c.Close()
		}
	}()

	if background {
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

func (sh *Shell) executePipeline(pipeline *parser.Pipeline) (code int, ok bool) {
	cmds := pipeline.Commands
	n := len(cmds)

	type pipeEnds struct{ r, w *os.File }
	pipes := make([]pipeEnds, n-1)
	for i := range pipes {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(sh.stderr, "Error: %v\n", err)
			for j := 0; j < i; j++ {
				_ = pipes[j].r.Close()
				_ = pipes[j].w.Close()
			}
			return 0, false
		}
		pipes[i] = pipeEnds{r: r, w: w}
	}

	var (
		wg       sync.WaitGroup
		procs    []*exec.Cmd
		lastProc *exec.Cmd
		lastErr  error
		lastMu   sync.Mutex
	)

	closePipeEnds := func(ends []io.Closer) {
		for _, c := range ends {
			_ = c.Close()
		}
	}

	for i, command := range cmds {
		ctx := &commands.Context{
			Stdout: sh.stdout,
			Stderr: sh.stderr,
			Stdin:  sh.stdin,
		}

		var stagePipeEnds []io.Closer
		if i > 0 {
			ctx.Stdin = pipes[i-1].r
			stagePipeEnds = append(stagePipeEnds, pipes[i-1].r)
		}
		if i < n-1 {
			ctx.Stdout = pipes[i].w
			stagePipeEnds = append(stagePipeEnds, pipes[i].w)
		}

		opened, err := sh.applyRedirections(ctx, command.Redirections)
		if err != nil {
			fmt.Fprintf(sh.stderr, "Error: %v\n", err)
			closePipeEnds(stagePipeEnds)
			for _, c := range opened {
				_ = c.Close()
			}
			// Unstarted pipe ends still open — close everything remaining.
			for _, p := range pipes {
				_ = p.r.Close()
				_ = p.w.Close()
			}
			for _, p := range procs {
				_ = p.Wait()
			}
			wg.Wait()
			return 0, false
		}

		isLast := i == n-1
		allClosers := append([]io.Closer(nil), stagePipeEnds...)
		allClosers = append(allClosers, opened...)

		if builtin, found := sh.registry.Get(command.Name); found {
			wg.Add(1)
			go func(b commands.Command, ctx *commands.Context, args []string, closers []io.Closer, isLast bool) {
				defer wg.Done()
				defer closePipeEnds(closers)
				err := b.Execute(ctx, args)
				if isLast {
					lastMu.Lock()
					lastErr = err
					lastMu.Unlock()
				}
			}(builtin, ctx, append([]string(nil), command.Args...), allClosers, isLast)
			continue
		}

		if _, err := exec.LookPath(command.Name); err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: command not found\n", command.Name)
			closePipeEnds(allClosers)
			continue
		}

		proc, err := commands.StartExternal(command.Name, command.Args, ctx, false)
		closePipeEnds(allClosers)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%v\n", err)
			if isLast {
				lastErr = err
			}
			continue
		}
		procs = append(procs, proc)
		if isLast {
			lastProc = proc
		}
	}

	for _, proc := range procs {
		err := proc.Wait()
		if proc == lastProc {
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					lastErr = &commands.StatusError{Code: exitErr.ExitCode()}
				} else {
					lastErr = err
				}
			} else {
				lastErr = nil
			}
		}
	}
	wg.Wait()

	if lastErr != nil {
		var statusErr *commands.StatusError
		if errors.As(lastErr, &statusErr) {
			return 0, false
		}
		// ExitError from a pipeline builtin must not terminate the shell.
		var exitErr *commands.ExitError
		if errors.As(lastErr, &exitErr) {
			return 0, false
		}
		fmt.Fprintf(sh.stderr, "%v\n", lastErr)
	}
	return 0, false
}

func (sh *Shell) applyRedirections(ctx *commands.Context, redirections []parser.Redirection) ([]io.Closer, error) {
	opened := make([]io.Closer, 0, len(redirections))
	closeOpened := func() {
		for _, c := range opened {
			_ = c.Close()
		}
	}

	for _, redir := range redirections {
		switch redir.Type {
		case parser.TokenRedirectOut:
			file, err := os.Create(redir.FilePath)
			if err != nil {
				closeOpened()
				return nil, err
			}
			opened = append(opened, file)
			ctx.Stdout = file
		case parser.TokenRedirectAppend:
			file, err := os.OpenFile(redir.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				closeOpened()
				return nil, err
			}
			opened = append(opened, file)
			ctx.Stdout = file
		case parser.TokenRedirectErr:
			file, err := os.Create(redir.FilePath)
			if err != nil {
				closeOpened()
				return nil, err
			}
			opened = append(opened, file)
			ctx.Stderr = file
		case parser.TokenRedirectAppendErr:
			file, err := os.OpenFile(redir.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				closeOpened()
				return nil, err
			}
			opened = append(opened, file)
			ctx.Stderr = file
		case parser.TokenRedirectIn:
			file, err := os.Open(redir.FilePath)
			if err != nil {
				closeOpened()
				return nil, err
			}
			opened = append(opened, file)
			ctx.Stdin = file
		}
	}
	return opened, nil
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
