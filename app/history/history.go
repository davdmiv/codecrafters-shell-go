// Package history stores in-memory command history for the shell.
package history

import (
	"fmt"
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
	if n := len(h.entries); n > 0 && h.entries[n-1] == line {
		return
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

// FormatEntry formats a history line like bash: "    1 echo hi".
func FormatEntry(e Entry) string {
	return fmt.Sprintf("%5d %s", e.Number, e.Command)
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
