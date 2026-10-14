package completion

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CompletionArgs returns (command, currentWord, previousWord) for a -C completer,
// matching bash's argument convention.
func CompletionArgs(line string, cursor int) (command, current, previous string) {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(line) {
		cursor = len(line)
	}
	before := line[:cursor]

	endsWithSpace := false
	if len(before) > 0 {
		r, _ := utf8.DecodeLastRuneInString(before)
		endsWithSpace = unicode.IsSpace(r)
	}

	fields := strings.Fields(before)
	if len(fields) == 0 {
		return "", "", ""
	}
	command = fields[0]

	if endsWithSpace {
		current = ""
		previous = fields[len(fields)-1]
		return command, current, previous
	}

	current = fields[len(fields)-1]
	if len(fields) >= 2 {
		previous = fields[len(fields)-2]
	}
	return command, current, previous
}

// FirstCommand returns the first whitespace-separated token of line, or "".
func FirstCommand(line string) string {
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	if line == "" {
		return ""
	}
	end := strings.IndexFunc(line, unicode.IsSpace)
	if end < 0 {
		return line
	}
	return line[:end]
}

// RunCompleter invokes a bash-style complete -C script.
// Env: COMP_LINE, COMP_POINT. Args: command, currentWord, previousWord.
// Candidates are read from stdout, one per line. stderr is streamed to stderrOut.
func RunCompleter(script, line string, cursor int, stderrOut io.Writer) ([]string, error) {
	command, current, previous := CompletionArgs(line, cursor)

	cmd := exec.Command(script, command, current, previous)
	cmd.Env = append(os.Environ(),
		"COMP_LINE="+line,
		fmt.Sprintf("COMP_POINT=%d", cursor),
	)
	if stderrOut == nil {
		stderrOut = io.Discard
	}
	cmd.Stderr = stderrOut

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// Non-zero exit still may have produced candidates; keep what we read.
		// Only fail hard if the process could not start.
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}

	var matches []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Completer may return broader lists; keep prefix matches for the token.
		if current == "" || strings.HasPrefix(line, current) {
			matches = append(matches, line)
		}
	}
	return matches, scanner.Err()
}
