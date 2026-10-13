// Package completion provides command-name completion for the shell.
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

// Complete returns sorted command names that start with the command prefix at
// cursor. Builtins and executable files found in PATH are both considered.
// Cursor is a byte offset into line. Completion is limited to the first word;
// argument and path completion are intentionally left to a separate feature.
func Complete(line string, cursor int, builtins []string) []string {
	word, _, ok := CommandWord(line, cursor)
	if !ok {
		return nil
	}

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
