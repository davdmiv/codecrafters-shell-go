package commands

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/completion"
)

func TestComplete_Name(t *testing.T) {
	if got := NewComplete(completion.NewSpecs()).Name(); got != "complete" {
		t.Fatalf("Name() = %q, want complete", got)
	}
}

func TestComplete_RegisterAndRemove(t *testing.T) {
	specs := completion.NewSpecs()
	ctx := &Context{Stdout: ioDiscard{}, Stderr: ioDiscard{}}

	err := NewComplete(specs).Execute(ctx, []string{"-C", "/tmp/completer", "git"})
	if err != nil {
		t.Fatalf("Execute -C: %v", err)
	}
	script, ok := specs.Get("git")
	if !ok || script != "/tmp/completer" {
		t.Fatalf("Get(git) = %q, %v; want /tmp/completer, true", script, ok)
	}

	err = NewComplete(specs).Execute(ctx, []string{"-r", "git"})
	if err != nil {
		t.Fatalf("Execute -r: %v", err)
	}
	if _, ok := specs.Get("git"); ok {
		t.Fatal("Get(git) after -r: want missing")
	}
}

func TestComplete_UsageErrors(t *testing.T) {
	specs := completion.NewSpecs()
	ctx := &Context{Stdout: ioDiscard{}, Stderr: ioDiscard{}}
	cmd := NewComplete(specs)

	if err := cmd.Execute(ctx, nil); err == nil {
		t.Fatal("want usage error")
	}
	if err := cmd.Execute(ctx, []string{"-C", "/tmp/x"}); err == nil {
		t.Fatal("want -C arity error")
	}
	if err := cmd.Execute(ctx, []string{"-r"}); err == nil {
		t.Fatal("want -r arity error")
	}
}

func TestType_CompleteIsBuiltin(t *testing.T) {
	reg := NewRegistry()
	specs := completion.NewSpecs()
	reg.Register(NewComplete(specs))
	reg.Register(NewType(reg))

	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: ioDiscard{}}
	if err := NewType(reg).Execute(ctx, []string{"complete"}); err != nil {
		t.Fatalf("type complete: %v", err)
	}
	if got, want := out.String(), "complete is a shell builtin\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
