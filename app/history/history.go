// Package history stores in-memory command history for the shell.
package history

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Entry is one numbered history line.
type Entry struct {
	Number  int
	Command string
}

// History is an in-memory list of past commands (no persistence).
type History struct {
	mu      sync.Mutex
	entries []string
	saved   int // entries[:saved] already written via -w / -a
}

func New() *History {
	return &History{}
}

// Add appends a non-empty command. Consecutive duplicates are ignored.
func (h *History) Add(line string) {
	if line == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.addLocked(line, true)
}

// ReadFile appends every non-empty line from path to in-memory history
// (bash `history -r`). Consecutive-duplicate skipping is not applied —
// file contents are appended as-is.
func (h *History) ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h.mu.Lock()
	defer h.mu.Unlock()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		h.addLocked(line, false)
	}
	return scanner.Err()
}

// MarkSaved treats the current list as already persisted (e.g. after
// loading HISTFILE on startup), so a later AppendFile only writes new lines.
func (h *History) MarkSaved() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.saved = len(h.entries)
}

// WriteFile overwrites path with the full in-memory history
// (bash `history -w`), one command per line.
func (h *History) WriteFile(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := writeLines(f, h.entries); err != nil {
		return err
	}
	h.saved = len(h.entries)
	return nil
}

// AppendFile appends history lines not yet written by -w/-a
// (bash `history -a`).
func (h *History) AppendFile(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.saved > len(h.entries) {
		h.saved = len(h.entries)
	}
	pending := h.entries[h.saved:]
	if len(pending) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := writeLines(f, pending); err != nil {
		return err
	}
	h.saved = len(h.entries)
	return nil
}

func writeLines(f *os.File, lines []string) error {
	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := w.WriteString(line); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (h *History) addLocked(line string, skipDup bool) {
	if skipDup {
		if n := len(h.entries); n > 0 && h.entries[n-1] == line {
			return
		}
	}
	h.entries = append(h.entries, line)
}

// Len returns the number of stored commands.
func (h *History) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.entries)
}

// All returns every entry with 1-based numbers.
func (h *History) All() []Entry {
	return h.Last(0)
}

// Last returns the last n entries (with original numbers).
// If n <= 0 or n >= Len, returns all entries.
func (h *History) Last(n int) []Entry {
	h.mu.Lock()
	defer h.mu.Unlock()

	start := 0
	if n > 0 && n < len(h.entries) {
		start = len(h.entries) - n
	}
	out := make([]Entry, 0, len(h.entries)-start)
	for i := start; i < len(h.entries); i++ {
		out = append(out, Entry{Number: i + 1, Command: h.entries[i]})
	}
	return out
}

// Get returns the command at 0-based index, or "" if out of range.
func (h *History) Get(index int) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if index < 0 || index >= len(h.entries) {
		return ""
	}
	return h.entries[index]
}

// FormatEntry formats a history line like bash: "    1  echo hi".
func FormatEntry(e Entry) string {
	return fmt.Sprintf("%5d  %s", e.Number, e.Command)
}

// Navigator walks history with ↑/↓ during a single readline session.
type Navigator struct {
	h     *History
	draft string // line being edited when not browsing history
	index int    // len(entries) == at draft; otherwise 0..len-1
}

// BeginNav starts navigation; draft is the current input buffer.
func (h *History) BeginNav(draft string) *Navigator {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &Navigator{h: h, draft: draft, index: len(h.entries)}
}

// Up moves to an older entry. ok is false at the oldest entry (or if empty).
func (n *Navigator) Up() (line string, ok bool) {
	n.h.mu.Lock()
	defer n.h.mu.Unlock()
	if len(n.h.entries) == 0 {
		return n.draft, false
	}
	if n.index == 0 {
		return n.h.entries[0], false
	}
	if n.index > len(n.h.entries) {
		n.index = len(n.h.entries)
	}
	n.index--
	return n.h.entries[n.index], true
}

// Down moves to a newer entry / back to the draft.
// ok is false when already on the draft.
func (n *Navigator) Down() (line string, ok bool) {
	n.h.mu.Lock()
	defer n.h.mu.Unlock()
	if n.index >= len(n.h.entries) {
		return n.draft, false
	}
	n.index++
	if n.index >= len(n.h.entries) {
		return n.draft, true
	}
	return n.h.entries[n.index], true
}

// SyncDraft records the current buffer when the user edits it (leaves browse mode).
func (n *Navigator) SyncDraft(line string) {
	n.h.mu.Lock()
	defer n.h.mu.Unlock()
	n.draft = line
	n.index = len(n.h.entries)
}
