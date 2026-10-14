package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCompletionArgs(t *testing.T) {
	tests := []struct {
		line            string
		wantCmd, wantCur, wantPrev string
	}{
		{"git cl", "git", "cl", "git"},
		{"git ", "git", "", "git"},
		{"git clone ", "git", "", "clone"},
		{"", "", "", ""},
	}
	for _, tt := range tests {
		cmd, cur, prev := CompletionArgs(tt.line, len(tt.line))
		if cmd != tt.wantCmd || cur != tt.wantCur || prev != tt.wantPrev {
			t.Errorf("CompletionArgs(%q) = (%q,%q,%q), want (%q,%q,%q)",
				tt.line, cmd, cur, prev, tt.wantCmd, tt.wantCur, tt.wantPrev)
		}
	}
}

func TestRunCompleter(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "completer")
	content := "#!/bin/sh\n" +
		"case \"$2\" in\n" +
		"  cl*) printf '%s\\n' clone;;\n" +
		"esac\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := RunCompleter(script, "git cl", len("git cl"), nil)
	if err != nil {
		t.Fatalf("RunCompleter: %v", err)
	}
	want := []string{"clone"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunCompleter() = %#v, want %#v", got, want)
	}
}

func TestFirstCommand(t *testing.T) {
	if got := FirstCommand("  git cl"); got != "git" {
		t.Fatalf("FirstCommand = %q, want git", got)
	}
}
