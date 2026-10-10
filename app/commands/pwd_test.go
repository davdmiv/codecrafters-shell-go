package commands

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPwd_Execute(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var out bytes.Buffer
	ctx := &Context{
		Stdout: &out,
		Stderr: io.Discard,
	}

	err := NewPwd().Execute(ctx, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	want += "\n"

	if got := out.String(); got != want {
		t.Fatalf("stdout: got %q, want %q", got, want)
	}
}
