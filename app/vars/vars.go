// Package vars stores shell variables set via declare.
package vars

import (
	"sync"
	"unicode"
)

// Store holds shell variable name→value bindings.
type Store struct {
	mu   sync.Mutex
	vars map[string]string
}

func New() *Store {
	return &Store{vars: make(map[string]string)}
}

// Set assigns value to name (overwrites if present).
func (s *Store) Set(name, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[name] = value
}

// Get returns the value and whether the variable is set.
func (s *Store) Get(name string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.vars[name]
	return v, ok
}

// Lookup returns the value or "" if the variable is unset (for $VAR expansion).
func (s *Store) Lookup(name string) string {
	v, _ := s.Get(name)
	return v
}

// IsValidIdentifier reports whether name is a valid shell variable name:
// [A-Za-z_][A-Za-z0-9_]*.
func IsValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
