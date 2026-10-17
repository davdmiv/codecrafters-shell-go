package commands

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/shell-starter-go/app/history"
)

func TestHistory_Name(t *testing.T) {
	if got, want := NewHistory(history.New()).Name(), "history"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestHistory_Execute_ListAndLimit(t *testing.T) {
	store := history.New()
	store.Add("echo a")
	store.Add("echo b")
	store.Add("echo c")

	var out bytes.Buffer
	ctx := &Context{Stdout: &out}
	if err := NewHistory(store).Execute(ctx, nil); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	want := "    1 echo a\n    2 echo b\n    3 echo c\n"
	if got := out.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}

	out.Reset()
	if err := NewHistory(store).Execute(ctx, []string{"2"}); err != nil {
		t.Fatalf("Execute limit: %v", err)
	}
	want = "    2 echo b\n    3 echo c\n"
	if got := out.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestHistory_Execute_BadArg(t *testing.T) {
	err := NewHistory(history.New()).Execute(&Context{Stdout: &bytes.Buffer{}}, []string{"x"})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestType_HistoryIsBuiltin(t *testing.T) {
	reg := NewRegistry()
	store := history.New()
	reg.Register(NewHistory(store))
	reg.Register(NewType(reg))

	var out bytes.Buffer
	ctx := &Context{Stdout: &out, Stderr: &bytes.Buffer{}}
	if err := NewType(reg).Execute(ctx, []string{"history"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got, want := out.String(), "history is a shell builtin\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}
