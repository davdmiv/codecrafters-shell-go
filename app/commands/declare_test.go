package commands

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/vars"
)

func TestDeclare_Name(t *testing.T) {
	if got, want := NewDeclare(vars.New()).Name(), "declare"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestDeclare_AssignAndPrint(t *testing.T) {
	store := vars.New()
	cmd := NewDeclare(store)
	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &bytes.Buffer{}}

	if err := cmd.Execute(ctx, []string{"orange=pear"}); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("assign stdout = %q, want empty", out.String())
	}

	if err := cmd.Execute(ctx, []string{"-p", "orange"}); err != nil {
		t.Fatalf("print: %v", err)
	}
	if got, want := out.String(), "declare -- orange=\"pear\"\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestDeclare_PrintMissing(t *testing.T) {
	err := NewDeclare(vars.New()).Execute(
		&Context{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		[]string{"-p", "missing_var"},
	)
	if err == nil {
		t.Fatal("want error")
	}
	if got, want := err.Error(), "declare: missing_var: not found"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestDeclare_InvalidIdentifier(t *testing.T) {
	err := NewDeclare(vars.New()).Execute(
		&Context{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		[]string{"5raspberry=apple"},
	)
	if err == nil {
		t.Fatal("want error")
	}
	if got, want := err.Error(), "declare: `5raspberry=apple': not a valid identifier"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestType_DeclareIsBuiltin(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewDeclare(vars.New()))
	reg.Register(NewType(reg))

	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &bytes.Buffer{}}
	if err := NewType(reg).Execute(ctx, []string{"declare"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "declare is a shell builtin\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}
