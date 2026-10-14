package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/codecrafters-io/shell-starter-go/app/completion"
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
		case 127, '\b':
			if len(line) == 0 {
				continue
			}
			_, size := utf8.DecodeLastRune(line)
			line = line[:len(line)-size]
			sh.tabPending = false
			if sh.interactive {
				fmt.Fprint(sh.stdout, "\b \b")
			}
		default:
			if r < 32 {
				continue
			}
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			line = append(line, buf[:n]...)
			sh.tabPending = false
			if sh.interactive {
				fmt.Fprint(sh.stdout, string(r))
			}
		}
	}
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

	matches := completion.Complete(lineStr, cursor, sh.registry.Names())
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
