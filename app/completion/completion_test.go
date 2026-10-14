package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCompleteCombinesBuiltinsAndExecutables(t *testing.T) {
	pathDir := t.TempDir()
	t.Setenv("PATH", pathDir)

	writeFile(t, filepath.Join(pathDir, "echo-external"), 0755)
	writeFile(t, filepath.Join(pathDir, "echo-disabled"), 0644)
	if err := os.Mkdir(filepath.Join(pathDir, "echo-directory"), 0755); err != nil {
		t.Fatal(err)
	}

	got := Complete("ec", 2, []string{"echo", "exit", "echo"})
	want := []string{"echo", "echo-external"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Complete() = %#v, want %#v", got, want)
	}
}

func TestCompleteCommandWord(t *testing.T) {
	t.Setenv("PATH", "")

	if got := Complete("", 0, []string{"echo", "exit"}); !reflect.DeepEqual(got, []string{"echo", "exit"}) {
		t.Fatalf("Complete() for empty command = %#v, want [echo exit]", got)
	}
	if got := Complete("   ec", len("   ec"), []string{"echo", "exit"}); !reflect.DeepEqual(got, []string{"echo"}) {
		t.Fatalf("Complete() with leading spaces = %#v, want [echo]", got)
	}
}

func TestCompletePathInArgument(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "hello.txt", 0644)
	writeFile(t, "help.md", 0644)
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}

	got := Complete("echo hel", len("echo hel"), nil)
	want := []string{"hello.txt", "help.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Complete() path arg = %#v, want %#v", got, want)
	}

	got = Complete("echo do", len("echo do"), nil)
	want = []string{"docs/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Complete() directory = %#v, want %#v", got, want)
	}
}

func TestCompletePathNestedAndAbsolute(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "alpha.txt"), 0644)

	token := filepath.Join(dir, "sub", "al")
	got := CompletePath(token)
	want := []string{filepath.Join(dir, "sub", "alpha.txt")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CompletePath(absolute) = %#v, want %#v", got, want)
	}

	t.Chdir(dir)
	got = CompletePath("sub/al")
	want = []string{"sub/alpha.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CompletePath(nested) = %#v, want %#v", got, want)
	}
}

func TestCompletePathTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(home, "notes.txt"), 0644)

	got := CompletePath("~")
	if !reflect.DeepEqual(got, []string{"~/"}) {
		t.Fatalf("CompletePath(~) = %#v, want [~/]", got)
	}

	got = CompletePath("~/no")
	want := []string{"~/notes.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CompletePath(~/no) = %#v, want %#v", got, want)
	}
}

func TestCompletePathSkipsHidden(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "visible.txt", 0644)
	writeFile(t, ".secret", 0644)

	got := CompletePath("")
	if !reflect.DeepEqual(got, []string{"visible.txt"}) {
		t.Fatalf("CompletePath() = %#v, want [visible.txt]", got)
	}

	got = CompletePath(".")
	if !reflect.DeepEqual(got, []string{".secret"}) {
		t.Fatalf("CompletePath(.) = %#v, want [.secret]", got)
	}
}

func TestCompleteRejectsInvalidCursor(t *testing.T) {
	for _, cursor := range []int{-1, 4} {
		if got := Complete("pwd", cursor, []string{"pwd"}); got != nil {
			t.Errorf("Complete(cursor=%d) = %#v, want nil", cursor, got)
		}
	}
}

func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"echo"}, "echo"},
		{[]string{"echo", "exit"}, "e"},
		{[]string{"echo", "echoes"}, "echo"},
		{[]string{"cat", "ls"}, ""},
	}
	for _, tt := range tests {
		if got := CommonPrefix(tt.in); got != tt.want {
			t.Errorf("CommonPrefix(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCommandWord(t *testing.T) {
	word, start, ok := CommandWord("  ech", 5)
	if !ok || word != "ech" || start != 2 {
		t.Fatalf("CommandWord() = %q, %d, %v; want ech, 2, true", word, start, ok)
	}
	if _, _, ok := CommandWord("echo hi", 7); ok {
		t.Fatal("CommandWord in argument: want ok=false")
	}
}

func TestTokenAtCursor(t *testing.T) {
	token, start, ok := TokenAtCursor("echo hel", 8)
	if !ok || token != "hel" || start != 5 {
		t.Fatalf("TokenAtCursor() = %q, %d, %v; want hel, 5, true", token, start, ok)
	}
	token, start, ok = TokenAtCursor("echo ", 5)
	if !ok || token != "" || start != 5 {
		t.Fatalf("TokenAtCursor(trailing space) = %q, %d, %v; want \"\", 5, true", token, start, ok)
	}
}

func writeFile(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}
