package commands

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestType_Name(t *testing.T) {
	reg := NewRegistry()
	if got, want := NewType(reg).Name(), "type"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestType_Execute_Builtin(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewEcho())
	reg.Register(NewType(reg))

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	// args без имени команды — только то, что ищем
	err := NewType(reg).Execute(ctx, []string{"echo"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "echo is a shell builtin\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestType_Execute_External(t *testing.T) {
	path, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not found in PATH")
	}

	reg := NewRegistry()
	reg.Register(NewType(reg))

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err = NewType(reg).Execute(ctx, []string{"true"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "true is "+path+"\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestType_Execute_MissingArgument(t *testing.T) {
	reg := NewRegistry()
	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewType(reg).Execute(ctx, nil)
	if err == nil {
		t.Fatal("Execute: want error, got nil")
	}
	if got, want := err.Error(), "type: missing argument"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestType_Execute_NotFound(t *testing.T) {
	reg := NewRegistry()
	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}

	err := NewType(reg).Execute(ctx, []string{"nosuchcommand_xyz"})
	if err == nil {
		t.Fatal("Execute: want error, got nil")
	}
	if got, want := err.Error(), "nosuchcommand_xyz: not found"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}
