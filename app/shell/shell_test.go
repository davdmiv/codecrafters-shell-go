package shell

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

func runShell(input string) (stdout, stderr string, code int) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	code = New(&out, &errBuf, strings.NewReader(input)).Run()
	return out.String(), errBuf.String(), code
}

func TestRun_JobsBuiltinEmpty(t *testing.T) {
	stdout, stderr, code := runShell("jobs\nexit\n")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_BackgroundJobStartAndDone(t *testing.T) {
	stdout, stderr, code := runShell("true &\nsleep 0.1\nexit\n")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !regexp.MustCompile(`\[1\] \d+`).MatchString(stdout) {
		t.Errorf("stdout = %q, want job start [1] <pid>", stdout)
	}
	if !strings.Contains(stderr, "Done") || !strings.Contains(stderr, "true") || strings.Contains(stderr, "true &") {
		t.Errorf("stderr = %q, want Done notification without trailing &", stderr)
	}
}

func TestRun_BackgroundJobRunningListed(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}

	unblocked := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		w, err := os.OpenFile(fifo, os.O_WRONLY, 0)
		if err == nil {
			_ = w.Close()
		}
		close(unblocked)
	}()

	stdout, stderr, code := runShell("cat " + fifo + " &\njobs\nexit\n")
	<-unblocked

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !regexp.MustCompile(`\[1\] \d+`).MatchString(stdout) {
		t.Errorf("stdout = %q, want [1] <pid>", stdout)
	}
	if !strings.Contains(stdout, "Running") || !strings.Contains(stdout, "cat "+fifo+" &") {
		t.Errorf("stdout = %q, want Running listing with trailing &", stdout)
	}
	_ = stderr
}

func TestRun_EchoAndExit(t *testing.T) {
	stdout, stderr, code := runShell("echo hello world\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ hello world\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_DualPipelineBuiltinToExternal(t *testing.T) {
	stdout, stderr, code := runShell("echo hello | cat\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ hello\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_DualPipelineExternals(t *testing.T) {
	stdout, stderr, code := runShell("printf 'b\\na\\n' | sort\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ a\nb\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_MultiPipeline(t *testing.T) {
	stdout, stderr, code := runShell("echo hello | cat | cat\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ hello\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_PipelineRedirectOut(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	stdout, stderr, code := runShell("echo hello | cat > " + outFile + "\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	if got, want := string(data), "hello\n"; got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestRun_ExitWithCode(t *testing.T) {
	stdout, stderr, code := runShell("exit 3\n")

	if code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
	if got, want := stdout, "$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_EmptyLines(t *testing.T) {
	stdout, stderr, code := runShell("\n\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_CommandNotFound(t *testing.T) {
	stdout, stderr, code := runShell("nosuchcommand_xyz\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr, "nosuchcommand_xyz: command not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRun_ParseError(t *testing.T) {
	stdout, stderr, code := runShell("echo 'unclosed\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr, "Parse error: unterminated single quote\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRun_InvalidExitContinues(t *testing.T) {
	stdout, stderr, code := runShell("exit abc\nexit 0\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr, "exit: abc: numeric argument required\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRun_EOF(t *testing.T) {
	stdout, stderr, code := runShell("")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ \nexit\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_TypeBuiltin(t *testing.T) {
	stdout, stderr, code := runShell("type echo\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ echo is a shell builtin\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_QuotedArgs(t *testing.T) {
	stdout, stderr, code := runShell("echo \"hello world\"\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ hello world\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_ExternalCommand(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true not found in PATH")
	}

	stdout, stderr, code := runShell("true\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_ExternalNonZeroExitIsSilent(t *testing.T) {
	if _, err := exec.LookPath("false"); err != nil {
		t.Skip("false not found in PATH")
	}

	stdout, stderr, code := runShell("false\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty (non-zero exit is not a launch failure)", stderr)
	}
}

func TestRun_OutputRedirectionWithoutArguments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.txt")
	if err := os.WriteFile(path, []byte("old contents\n"), 0600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runShell("echo > " + path + "\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), "\n"; got != want {
		t.Errorf("redirected file = %q, want %q", got, want)
	}
}

func TestRun_AppendRedirection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runShell("echo second>>" + path + "\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), "first\nsecond\n"; got != want {
		t.Errorf("redirected file = %q, want %q", got, want)
	}
}

func TestRun_InputAndOutputRedirection(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat not found in PATH")
	}

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.txt")
	outputPath := filepath.Join(dir, "output.txt")
	input := []byte("content from redirected stdin\n")
	if err := os.WriteFile(inputPath, input, 0600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runShell("cat < " + inputPath + " > " + outputPath + "\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), string(input); got != want {
		t.Errorf("redirected file = %q, want %q", got, want)
	}
}

func TestRun_TabCompletesUniqueBuiltin(t *testing.T) {
	// Non-TTY: typed chars are not echoed; Tab writes only the completion suffix.
	stdout, stderr, code := runShell("ech\t\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// suffix "o " then echo with no args prints "\n"
	if got, want := stdout, "$ o \n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_TabCompletesFilename(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("hello.txt", []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runShell("echo hel\t\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// suffix "lo.txt " then echo prints "hello.txt\n"
	if got, want := stdout, "$ lo.txt hello.txt\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_TypeCompleteIsBuiltin(t *testing.T) {
	stdout, stderr, code := runShell("type complete\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ complete is a shell builtin\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_ProgrammableCompletion(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "git_completer")
	content := "#!/bin/sh\n" +
		"case \"$2\" in\n" +
		"  cl*) printf '%s\\n' clone;;\n" +
		"esac\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	// Register, Tab-complete "git cl" → "git clone ", then exit.
	input := "complete -C " + script + " git\ngit cl\t\nexit\n"
	stdout, _, code := runShell(input)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "one ") {
		t.Errorf("stdout = %q, want Tab to append %q", stdout, "one ")
	}
}

func TestRun_ProgrammableCompletionUnregister(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "git_completer")
	content := "#!/bin/sh\nprintf '%s\\n' clone\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := os.WriteFile("cloneable.txt", []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// After -r, Tab on "git clo" should fall back to path completion (cloneable.txt),
	// not the completer's "clone".
	input := "complete -C " + script + " git\ncomplete -r git\ngit clo\t\nexit\n"
	stdout, _, code := runShell(input)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "neable.txt ") {
		t.Errorf("stdout = %q, want path completion suffix after unregister", stdout)
	}
	if strings.Contains(stdout, "one ") && !strings.Contains(stdout, "neable.txt") {
		t.Errorf("stdout = %q, completer should not run after -r", stdout)
	}
}

func TestRun_TabCompletesDirectoryWithSlash(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runShell("echo do\t\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// directory match ends with "/" and no trailing space
	if got, want := stdout, "$ cs/docs/\n$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_TabNoMatchRingsBell(t *testing.T) {
	stdout, stderr, code := runShell("zzznomatch\t\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "\a") {
		t.Errorf("stdout = %q, want BEL (\\a) on missing completion", stdout)
	}
	if !strings.Contains(stderr, "zzznomatch: command not found") {
		t.Errorf("stderr = %q, want command not found", stderr)
	}
}

func TestRun_TabListsAmbiguousMatchesOnSecondPress(t *testing.T) {
	t.Setenv("PATH", "")

	stdout, stderr, code := runShell("e\t\t\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "\a") {
		t.Errorf("stdout = %q, want BEL on first ambiguous Tab", stdout)
	}
	if !strings.Contains(stdout, "echo") || !strings.Contains(stdout, "exit") {
		t.Errorf("stdout = %q, want listed builtins echo/exit on second Tab", stdout)
	}
	if stderr != "" && !strings.Contains(stderr, "command not found") {
		// After listing, Enter runs incomplete "e" → command not found is ok.
		t.Errorf("stderr = %q", stderr)
	}
}

func TestRun_EOFExecutesPartialLine(t *testing.T) {
	stdout, stderr, code := runShell("exit 7")

	if code != 7 {
		t.Errorf("exit code = %d, want 7", code)
	}
	if got, want := stdout, "$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRun_StderrRedirection(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat not found in PATH")
	}

	dir := t.TempDir()
	missingPath := filepath.Join(dir, "missing.txt")
	errorPath := filepath.Join(dir, "errors.log")

	stdout, stderr, code := runShell("cat " + missingPath + " 2>" + errorPath + "\nexit\n")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if got, want := stdout, "$ $ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty after redirection", stderr)
	}
	contents, err := os.ReadFile(errorPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) == 0 {
		t.Error("stderr redirection file is empty, want cat's error message")
	}
}
