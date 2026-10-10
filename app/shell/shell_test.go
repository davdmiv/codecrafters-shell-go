package shell

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func runShell(input string) (stdout, stderr string, code int) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	code = New(&out, &errBuf, strings.NewReader(input)).Run()
	return out.String(), errBuf.String(), code
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
