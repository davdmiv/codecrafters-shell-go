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

func TestComplete_PrintSpec(t *testing.T) {
	specs := completion.NewSpecs()
	cmd := NewComplete(specs)
	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: ioDiscard{}}

	if err := cmd.Execute(ctx, []string{"-C", "/tmp/completer", "git"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	out.Reset()
	if err := cmd.Execute(ctx, []string{"-p", "git"}); err != nil {
		t.Fatalf("Execute -p: %v", err)
	}
	if got, want := out.String(), "complete -C '/tmp/completer' git\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}

	out.Reset()
	err := cmd.Execute(ctx, []string{"-p", "nosuch"})
	if err == nil {
		t.Fatal("want error for missing spec")
	}
	if got, want := err.Error(), "complete: nosuch: no completion specification"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestComplete_PrintAllSpecs(t *testing.T) {
	specs := completion.NewSpecs()
	cmd := NewComplete(specs)
	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: ioDiscard{}}

	_ = cmd.Execute(ctx, []string{"-C", "/tmp/b", "zeta"})
	_ = cmd.Execute(ctx, []string{"-C", "/tmp/a", "alpha"})

	out.Reset()
	if err := cmd.Execute(ctx, []string{"-p"}); err != nil {
		t.Fatalf("Execute -p: %v", err)
	}
	want := "complete -C '/tmp/a' alpha\ncomplete -C '/tmp/b' zeta\n"
	if got := out.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
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
