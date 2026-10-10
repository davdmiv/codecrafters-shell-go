package commands

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestExternal_Name(t *testing.T) {
	cmd := NewExternal("ls")
	if got, want := cmd.Name(), "ls"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestExternal_Execute_Stdout(t *testing.T) {
	echoPath := lookPathOrSkip(t, "echo")

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewExternal(echoPath).Execute(ctx, []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "hello world\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestExternal_Execute_Stderr(t *testing.T) {
	shPath := lookPathOrSkip(t, "sh")

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewExternal(shPath).Execute(ctx, []string{"-c", "echo fail >&2"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := errBuf.String(), "fail\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestExternal_Execute_Stdin(t *testing.T) {
	catPath := lookPathOrSkip(t, "cat")

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{
		Stdout: &out,
		Stderr: &errBuf,
		Stdin:  strings.NewReader("from stdin"),
	}

	err := NewExternal(catPath).Execute(ctx, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "from stdin"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestExternal_Execute_NonZeroExit(t *testing.T) {
	falsePath := lookPathOrSkip(t, "false")

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewExternal(falsePath).Execute(ctx, nil)
	if err == nil {
		t.Fatal("Execute: want StatusError for non-zero exit, got nil")
	}

	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("Execute: want StatusError, got %T: %v", err, err)
	}
	if statusErr.Code == 0 {
		t.Errorf("StatusError.Code = 0, want non-zero")
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestExternal_Execute_NotFound(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewExternal("command-that-does-not-exist-xyz").Execute(ctx, nil)
	if err == nil {
		t.Fatal("Execute: want error for missing command, got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute") {
		t.Errorf("error = %q, want substring %q", err.Error(), "failed to execute")
	}
}

func lookPathOrSkip(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not found in PATH: %v", name, err)
	}
	return path
}
