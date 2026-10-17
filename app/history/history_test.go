package history

import (
	"os"
	"testing"
)

func TestHistory_AddAndAll(t *testing.T) {
	h := New()
	h.Add("echo a")
	h.Add("echo b")
	h.Add("") // ignored
	h.Add("echo b") // consecutive duplicate ignored
	h.Add("echo c")

	got := h.All()
	want := []Entry{
		{1, "echo a"},
		{2, "echo b"},
		{3, "echo c"},
	}
	if len(got) != len(want) {
		t.Fatalf("All() len = %d, want %d (%#v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("All()[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestHistory_Last(t *testing.T) {
	h := New()
	h.Add("a")
	h.Add("b")
	h.Add("c")

	got := h.Last(2)
	want := []Entry{{2, "b"}, {3, "c"}}
	if len(got) != len(want) {
		t.Fatalf("Last(2) = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Last(2)[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}

	if all := h.Last(0); len(all) != 3 {
		t.Errorf("Last(0) len = %d, want 3", len(all))
	}
	if all := h.Last(99); len(all) != 3 {
		t.Errorf("Last(99) len = %d, want 3", len(all))
	}
}

func TestFormatEntry(t *testing.T) {
	if got, want := FormatEntry(Entry{1, "ls dist/"}), "    1  ls dist/"; got != want {
		t.Errorf("FormatEntry = %q, want %q", got, want)
	}
}

func TestHistory_AppendFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/hist"

	h := New()
	h.Add("echo a")
	h.Add("echo b")
	if err := h.AppendFile(path); err != nil {
		t.Fatalf("AppendFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "echo a\necho b\n"; got != want {
		t.Fatalf("after first -a: %q, want %q", got, want)
	}

	h.Add("echo c")
	if err := h.AppendFile(path); err != nil {
		t.Fatalf("AppendFile 2: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "echo a\necho b\necho c\n"; got != want {
		t.Fatalf("after second -a: %q, want %q", got, want)
	}

	// Nothing new — file unchanged.
	if err := h.AppendFile(path); err != nil {
		t.Fatalf("AppendFile 3: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "echo a\necho b\necho c\n"; got != want {
		t.Fatalf("after empty -a: %q, want %q", got, want)
	}
}

func TestHistory_WriteFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/hist"

	h := New()
	h.Add("echo a")
	h.Add("echo b")
	// Pre-existing content must be overwritten.
	if err := os.WriteFile(path, []byte("old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := h.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "echo a\necho b\n"; got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestHistory_ReadFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/hist"
	content := "echo hello\necho world\n\necho hello\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	h := New()
	h.Add("history -r " + path)
	if err := h.ReadFile(path); err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	got := h.All()
	want := []Entry{
		{1, "history -r " + path},
		{2, "echo hello"},
		{3, "echo world"},
		{4, "echo hello"},
	}
	if len(got) != len(want) {
		t.Fatalf("All() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("All()[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestNavigator_UpDown(t *testing.T) {
	h := New()
	h.Add("first")
	h.Add("second")

	nav := h.BeginNav("draft")

	line, ok := nav.Up()
	if !ok || line != "second" {
		t.Fatalf("Up() = %q %v, want second true", line, ok)
	}
	line, ok = nav.Up()
	if !ok || line != "first" {
		t.Fatalf("Up() = %q %v, want first true", line, ok)
	}
	line, ok = nav.Up()
	if ok || line != "first" {
		t.Fatalf("Up() at top = %q %v, want first false", line, ok)
	}

	line, ok = nav.Down()
	if !ok || line != "second" {
		t.Fatalf("Down() = %q %v, want second true", line, ok)
	}
	line, ok = nav.Down()
	if !ok || line != "draft" {
		t.Fatalf("Down() = %q %v, want draft true", line, ok)
	}
	line, ok = nav.Down()
	if ok || line != "draft" {
		t.Fatalf("Down() at bottom = %q %v, want draft false", line, ok)
	}
}
