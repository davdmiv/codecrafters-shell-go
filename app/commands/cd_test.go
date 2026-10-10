package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCd_Execute(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	destination := t.TempDir()
	trueDist, err := filepath.EvalSymlinks(destination)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks: %v", err)
	}

	var out bytes.Buffer
	var errBuf bytes.Buffer
	ctx := &Context{
		Stdout: &out,
		Stderr: &errBuf,
	}

	err = NewCd().Execute(ctx, []string{trueDist})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := errBuf.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}

	got, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	if want := trueDist; got != want {
		t.Errorf("shell cwd = %q, want %q", got, want)
	}

	err = NewPwd().Execute(ctx, nil)
	if err != nil {
		t.Fatalf("NewPwd Execute: %v", err)
	}

	if got, want := out.String(), trueDist+"\n"; got != want {
		t.Errorf("pwd stdout = %q, want %q", got, want)
	}
}
