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

func TestCompleteOnlyCompletesCommandWord(t *testing.T) {
	t.Setenv("PATH", "")

	if got := Complete("", 0, []string{"echo", "exit"}); !reflect.DeepEqual(got, []string{"echo", "exit"}) {
		t.Fatalf("Complete() for empty command = %#v, want [echo exit]", got)
	}
	if got := Complete("echo he", len("echo he"), []string{"echo", "exit"}); got != nil {
		t.Fatalf("Complete() in argument = %#v, want nil", got)
	}
	if got := Complete("   ec", len("   ec"), []string{"echo", "exit"}); !reflect.DeepEqual(got, []string{"echo"}) {
		t.Fatalf("Complete() with leading spaces = %#v, want [echo]", got)
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

func writeFile(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}
