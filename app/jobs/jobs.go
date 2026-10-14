// Package jobs manages background job state for the shell.
package jobs

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

// Manager tracks background jobs, assigns recycled IDs, and reaps finished ones.
type Manager struct {
	mu       sync.Mutex
	byID     map[int]*Job
	current  int // most recent job (+)
	previous int // previous current (-)
}

// Job is a single background process.
type Job struct {
	ID      int
	PID     int
	CmdLine string
	cmd     *exec.Cmd
	done    chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		byID: make(map[int]*Job),
	}
}

// Add registers a started command as a background job and returns it.
func (m *Manager) Add(cmd *exec.Cmd, cmdLine string) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextIDLocked()
	j := &Job{
		ID:      id,
		PID:     cmd.Process.Pid,
		CmdLine: cmdLine,
		cmd:     cmd,
		done:    make(chan struct{}),
	}
	go func() {
		_ = cmd.Wait()
		close(j.done)
	}()

	m.byID[id] = j
	if m.current != 0 {
		m.previous = m.current
	}
	m.current = id
	return j
}

func (m *Manager) nextIDLocked() int {
	id := 1
	for {
		if _, exists := m.byID[id]; !exists {
			return id
		}
		id++
	}
}

func (j *Job) finished() bool {
	select {
	case <-j.done:
		return true
	default:
		return false
	}
}

// Listing is a snapshot row for the jobs builtin.
type Listing struct {
	ID      int
	Marker  string // "+", "-", or " "
	Status  string // "Running" or "Done"
	CmdLine string
}

// List returns jobs sorted by ID. Finished jobs are included as Done so the
// caller can display them; use RemoveFinished afterwards (or NotifyFinished).
func (m *Manager) List() []Listing {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]int, 0, len(m.byID))
	for id := range m.byID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	out := make([]Listing, 0, len(ids))
	for _, id := range ids {
		j := m.byID[id]
		status := "Running"
		if j.finished() {
			status = "Done"
		}
		out = append(out, Listing{
			ID:      id,
			Marker:  m.markerLocked(id),
			Status:  status,
			CmdLine: j.CmdLine,
		})
	}
	return out
}

func (m *Manager) markerLocked(id int) string {
	switch id {
	case m.current:
		return "+"
	case m.previous:
		return "-"
	default:
		return " "
	}
}

// FormatListing formats a jobs(1)-style line matching bash/CodeCrafters:
//
//	[1]+  Running                 sleep 10 &
//	[1]+  Done                 cat /tmp/x
//
// Finished entries do not keep a trailing " &".
func FormatListing(l Listing) string {
	cmdline := l.CmdLine
	if l.Status == "Done" {
		cmdline = strings.TrimSuffix(cmdline, " &")
		cmdline = strings.TrimSuffix(cmdline, "&")
	}
	// Two spaces after the marker, then status, then exactly 17 spaces before the command.
	return fmt.Sprintf("[%d]%s  %s%s%s", l.ID, l.Marker, l.Status, strings.Repeat(" ", 17), cmdline)
}

// NotifyFinished writes Done notifications for finished jobs to w and removes them.
// Returns true if anything was printed.
func (m *Manager) NotifyFinished(w io.Writer) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]int, 0, len(m.byID))
	for id, j := range m.byID {
		if j.finished() {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	if len(ids) == 0 {
		return false
	}

	for _, id := range ids {
		j := m.byID[id]
		fmt.Fprintln(w, FormatListing(Listing{
			ID:      id,
			Marker:  m.markerLocked(id),
			Status:  "Done",
			CmdLine: j.CmdLine,
		}))
		m.removeLocked(id)
	}
	return true
}

// RemoveFinished drops jobs that have already exited (e.g. after `jobs` printed Done).
func (m *Manager) RemoveFinished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, j := range m.byID {
		if j.finished() {
			m.removeLocked(id)
		}
	}
}

func (m *Manager) removeLocked(id int) {
	delete(m.byID, id)
	if m.current == id {
		// Drop "+" without promoting "-". Remaining previous job keeps "-".
		m.current = 0
	}
	if m.previous == id {
		// Elect a new previous among remaining non-current jobs.
		m.previous = m.electPreviousLocked()
	}
}

// electPreviousLocked returns the highest remaining job ID that is not current.
func (m *Manager) electPreviousLocked() int {
	best := 0
	for id := range m.byID {
		if id == m.current {
			continue
		}
		if id > best {
			best = id
		}
	}
	return best
}
