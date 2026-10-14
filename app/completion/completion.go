// Package completion provides command-name and path completion for the shell.
package completion

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CommandWord returns the command-name prefix under cursor and its start offset.
// ok is false when the cursor is outside the first word (e.g. in arguments).
func CommandWord(line string, cursor int) (word string, start int, ok bool) {
	if cursor < 0 || cursor > len(line) {
		return "", 0, false
	}

	beforeCursor := line[:cursor]
	start = 0
	for start < len(beforeCursor) {
		r, size := utf8.DecodeRuneInString(beforeCursor[start:])
		if !unicode.IsSpace(r) {
			break
		}
		start += size
	}

	word = beforeCursor[start:]
	if strings.IndexFunc(word, unicode.IsSpace) >= 0 {
		return "", 0, false
	}
	return word, start, true
}

// TokenAtCursor returns the whitespace-delimited token under cursor.
func TokenAtCursor(line string, cursor int) (token string, start int, ok bool) {
	if cursor < 0 || cursor > len(line) {
		return "", 0, false
	}

	before := line[:cursor]
	start = 0
	for i := 0; i < len(before); {
		r, size := utf8.DecodeRuneInString(before[i:])
		if unicode.IsSpace(r) {
			start = i + size
		}
		i += size
	}
	return before[start:], start, true
}

// Complete returns sorted completion candidates for the cursor position.
// The first word is completed as a command (builtins + PATH); later tokens as paths.
// Directory matches end with "/".
func Complete(line string, cursor int, builtins []string) []string {
	if word, _, ok := CommandWord(line, cursor); ok {
		return completeCommands(word, builtins)
	}

	token, _, ok := TokenAtCursor(line, cursor)
	if !ok {
		return nil
	}
	return CompletePath(token)
}

func completeCommands(word string, builtins []string) []string {
	candidates := make(map[string]struct{})
	for _, name := range builtins {
		if strings.HasPrefix(name, word) {
			candidates[name] = struct{}{}
		}
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, word) {
				continue
			}
			info, err := os.Stat(filepath.Join(dir, name))
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
				continue
			}
			candidates[name] = struct{}{}
		}
	}

	result := make([]string, 0, len(candidates))
	for name := range candidates {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// CompletePath returns path completions for token. Directories end with "/".
// Supports relative paths, absolute paths, and "~" / "~/...".
func CompletePath(token string) []string {
	if token == "~" {
		return []string{"~/"}
	}

	var (
		displayDir string // prefix kept in the returned candidates
		readDir    string // directory to list
		prefix     string
	)

	switch {
	case strings.HasPrefix(token, "~/"):
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		rest := token[2:]
		subDir, name := filepath.Split(rest)
		displayDir = "~/" + subDir
		readDir = filepath.Join(home, subDir)
		prefix = name
	default:
		subDir, name := filepath.Split(token)
		displayDir = subDir
		if subDir == "" {
			readDir = "."
		} else {
			readDir = subDir
		}
		prefix = name
	}

	entries, err := os.ReadDir(readDir)
	if err != nil {
		return nil
	}

	var result []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		candidate := displayDir + name
		if entry.IsDir() {
			candidate += "/"
		}
		result = append(result, candidate)
	}
	sort.Strings(result)
	return result
}

// CommonPrefix returns the longest shared prefix of strs.
func CommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
