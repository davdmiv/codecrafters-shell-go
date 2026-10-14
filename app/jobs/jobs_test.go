package jobs

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func TestManager_AddListReap(t *testing.T) {
	m := NewManager()
	cmd := exec.Command("sleep", "0.2")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	job := m.Add(cmd, "sleep 0.2 &")
	if job.ID != 1 || job.PID != cmd.Process.Pid {
		t.Fatalf("job = %+v", job)
	}

	list := m.List()
	if len(list) != 1 || list[0].Status != "Running" || list[0].Marker != "+" {
		t.Fatalf("List() = %+v", list)
	}

	time.Sleep(300 * time.Millisecond)

	var buf bytes.Buffer
	if !m.NotifyFinished(&buf) {
		t.Fatal("expected NotifyFinished to report Done")
	}
	want := fmt.Sprintf("[1]+  %-21s%s\n", "Done", "sleep 0.2")
	if got := buf.String(); got != want {
		t.Fatalf("NotifyFinished = %q, want %q", got, want)
	}
	if len(m.List()) != 0 {
		t.Fatalf("jobs remain after notify: %+v", m.List())
	}
}

func TestManager_RecycleIDs(t *testing.T) {
	m := NewManager()

	start := func(arg string) *exec.Cmd {
		cmd := exec.Command("true")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		return cmd
	}

	j1 := m.Add(start("a"), "true &")
	j2 := m.Add(start("b"), "true &")
	if j1.ID != 1 || j2.ID != 2 {
		t.Fatalf("ids = %d, %d", j1.ID, j2.ID)
	}

	time.Sleep(50 * time.Millisecond)
	m.NotifyFinished(ioDiscard{})

	j3 := m.Add(start("c"), "true &")
	if j3.ID != 1 {
		t.Fatalf("recycled id = %d, want 1", j3.ID)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func TestManager_JobsBuiltinRemovesDone(t *testing.T) {
	m := NewManager()
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	m.Add(cmd, "true &")
	time.Sleep(50 * time.Millisecond)

	list := m.List()
	if len(list) != 1 || list[0].Status != "Done" {
		t.Fatalf("List() = %+v", list)
	}
	m.RemoveFinished()
	if len(m.List()) != 0 {
		t.Fatal("expected empty after RemoveFinished")
	}
}

func TestFormatListing(t *testing.T) {
	got := FormatListing(Listing{ID: 1, Marker: "+", Status: "Running", CmdLine: "sleep 1 &"})
	want := "[1]+  Running              sleep 1 &"
	if got != want {
		t.Fatalf("FormatListing Running = %q, want %q", got, want)
	}

	got = FormatListing(Listing{ID: 1, Marker: "+", Status: "Done", CmdLine: "cat /tmp/pear-94 &"})
	want = "[1]+  Done                 cat /tmp/pear-94"
	if got != want {
		t.Fatalf("FormatListing Done = %q, want %q", got, want)
	}
}
