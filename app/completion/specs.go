package completion

import (
	"sort"
	"sync"
)

// Specs хранит programmable completions: имя команды → путь к -C скрипту.
type Specs struct {
	mu    sync.RWMutex
	byCmd map[string]string
}

func NewSpecs() *Specs {
	return &Specs{byCmd: make(map[string]string)}
}

func (s *Specs) Set(command, script string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byCmd[command] = script
}

func (s *Specs) Remove(command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byCmd, command)
}

func (s *Specs) Get(command string) (script string, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	script, ok = s.byCmd[command]
	return script, ok
}

// Commands returns sorted command names that have a completion spec.
func (s *Specs) Commands() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.byCmd))
	for cmd := range s.byCmd {
		names = append(names, cmd)
	}
	sort.Strings(names)
	return names
}
