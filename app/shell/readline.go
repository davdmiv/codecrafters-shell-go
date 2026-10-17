package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/app/completion"
	"github.com/codecrafters-io/shell-starter-go/app/history"
	"golang.org/x/term"
)

func (sh *Shell) enableRawMode() (restore func()) {
	f, ok := sh.stdin.(*os.File)
	if !ok {
		return func() {}
	}
	fd := int(f.Fd())
	if !term.IsTerminal(fd) {
		return func() {}
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}
	}
	sh.interactive = true
	return func() {
		_ = term.Restore(fd, oldState)
		sh.interactive = false
	}
}

func (sh *Shell) readLine(reader *bufio.Reader) (string, error) {
	var line []byte
	sh.tabPending = false
	nav := sh.history.BeginNav("")

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					return "", io.EOF
				}
				return string(line), nil
			}
			return "", err
		}

		switch r {
		case '\r':
			if b, err := reader.Peek(1); err == nil && len(b) > 0 && b[0] == '\n' {
				_, _ = reader.ReadByte()
			}
			sh.finishLine()
			return string(line), nil
		case '\n':
			sh.finishLine()
			return string(line), nil
		case '\t':
			line = sh.handleTab(line)
			nav.SyncDraft(string(line))
		case 127, '\b':
			if len(line) == 0 {
				continue
			}
			_, size := utf8.DecodeLastRune(line)
			line = line[:len(line)-size]
			sh.tabPending = false
			nav.SyncDraft(string(line))
			if sh.interactive {
				fmt.Fprint(sh.stdout, "\b \b")
			}
		case 0x1b: // ESC — стрелки ↑/↓ (CSI или SS3)
			line = sh.handleEscape(reader, line, nav)
		default:
			if r < 32 {
				continue
			}
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			line = append(line, buf[:n]...)
			sh.tabPending = false
			nav.SyncDraft(string(line))
			if sh.interactive {
				fmt.Fprint(sh.stdout, string(r))
			}
		}
	}
}

func (sh *Shell) handleEscape(reader *bufio.Reader, line []byte, nav *history.Navigator) []byte {
	b, err := reader.ReadByte()
	if err != nil {
		return line
	}

	var key byte
	switch b {
	case '[': // CSI: ESC [ A/B
		key, err = reader.ReadByte()
		if err != nil {
			return line
		}
	case 'O': // SS3: ESC O A/B (application cursor keys)
		key, err = reader.ReadByte()
		if err != nil {
			return line
		}
	default:
		return line
	}

	var next string
	var moved bool
	switch key {
	case 'A': // up
		next, moved = nav.Up()
	case 'B': // down
		next, moved = nav.Down()
	default:
		return line
	}
	if !moved {
		if sh.interactive {
			fmt.Fprint(sh.stdout, "\a")
		}
		return line
	}
	return sh.replaceInputLine(line, next)
}

func (sh *Shell) replaceInputLine(old []byte, next string) []byte {
	if sh.interactive {
		// Стереть текущую строку ввода и напечатать выбранную из history.
		fmt.Fprint(sh.stdout, "\r\x1b[2K$ ")
		fmt.Fprint(sh.stdout, next)
	}
	sh.tabPending = false
	return []byte(next)
}

func (sh *Shell) finishLine() {
	if sh.interactive {
		fmt.Fprint(sh.stdout, "\r\n")
	}
}

func (sh *Shell) handleTab(line []byte) []byte {
	lineStr := string(line)
	cursor := len(lineStr)

	token, start, ok := completion.TokenAtCursor(lineStr, cursor)
	if !ok {
		fmt.Fprint(sh.stdout, "\a")
		sh.tabPending = false
		return line
	}

	matches := sh.completionMatches(lineStr, cursor)
	switch len(matches) {
	case 0:
		fmt.Fprint(sh.stdout, "\a")
		sh.tabPending = false
		return line
	case 1:
		// Директории уже с "/"; для файлов и команд добавляем пробел.
		completed := matches[0]
		if !strings.HasSuffix(completed, "/") {
			completed += " "
		}
		suffix := completed[len(token):]
		newLine := lineStr[:start] + completed
		fmt.Fprint(sh.stdout, suffix)
		sh.tabPending = false
		return []byte(newLine)
	default:
		lcp := completion.CommonPrefix(matches)
		if len(lcp) > len(token) {
			suffix := lcp[len(token):]
			newLine := lineStr[:start] + lcp
			fmt.Fprint(sh.stdout, suffix)
			sh.tabPending = false
			return []byte(newLine)
		}

		if !sh.tabPending {
			fmt.Fprint(sh.stdout, "\a")
			sh.tabPending = true
			return line
		}

		nl := "\n"
		if sh.interactive {
			nl = "\r\n"
		}
		fmt.Fprintf(sh.stdout, "%s%s%s$ %s", nl, strings.Join(matches, " "), nl, lineStr)
		sh.tabPending = false
		return line
	}
}

// completionMatches предпочитает programmable -C completer для зарегистрированной
// команды (первое слово строки), иначе — обычное command/path completion.
func (sh *Shell) completionMatches(line string, cursor int) []string {
	cmdName := completion.FirstCommand(line)
	if script, ok := sh.specs.Get(cmdName); ok {
		matches, err := completion.RunCompleter(script, line, cursor, sh.stderr)
		if err != nil {
			return nil
		}
		return matches
	}
	return completion.Complete(line, cursor, sh.registry.Names())
}
