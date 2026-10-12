package shell

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
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
