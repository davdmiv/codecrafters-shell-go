package commands

import (
	"bytes"
	"errors"
	"testing"
)

func TestExit(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	testRegistry := NewRegistry()
	testRegistry.Register(NewType(testRegistry))

	ctx := &Context{Stdout: &out, Stderr: &errBuf}
	err := NewExit().Execute(ctx, []string{"2"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("TestExit() return error: %q", err)
	}
	if got, want := out.String(), ""; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}
