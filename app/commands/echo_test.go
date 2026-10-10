package commands

import (
	"bytes"
	"testing"
)

func TestEcho(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &errBuf}
	err := NewEcho().Execute(ctx, []string{"hello", "world"})

	if err != nil {
		t.Errorf("TestEcho() return error: %q", err)
	}
	if got, want := out.String(), "hello world\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}
