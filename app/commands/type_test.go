package commands

import (
	"bytes"
	"testing"
)

func TestType(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	testRegistry := NewRegistry()
	testRegistry.Register(NewType(testRegistry))

	ctx := &Context{Stdout: &out, Stderr: &errBuf}
	err := NewType(testRegistry).Execute(ctx, []string{"type", "type"})

	if err != nil {
		t.Errorf("TestType() return error: %q", err)
	}
	if got, want := out.String(), "type is a shell builtin\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}
