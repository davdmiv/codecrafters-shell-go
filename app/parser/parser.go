package parser

import (
	"errors"
	"strings"
	"unicode"
)

// Lookup resolves a variable name to its value ("" if unset).
type Lookup func(name string) string

// Parse парсит строку и возвращает Pipeline (одна команда — тоже Pipeline).
// Если lookup != nil, раскрывает $VAR / ${VAR} вне одинарных кавычек.
func Parse(input string, lookup Lookup) (*Pipeline, error) {
	tokens, err := tokenize(input, lookup)
	if err != nil {
		return nil, err
	}
	return buildPipeline(tokens)
}

func tokenize(input string, lookup Lookup) ([]Token, error) {
	var tokens []Token
	var arg strings.Builder
	isSingleQuotes := false
	isDoubleQuotes := false
	isBackslash := false
	inArg := false
	argPlain := true
	argQuoted := false

	flushArg := func() {
		if !inArg {
			return
		}
		// Пустой неquoted аргумент после раскрытия $UNSET отбрасываем (как bash).
		if arg.Len() > 0 || argQuoted {
			tokens = append(tokens, Token{Type: TokenArg, Value: arg.String()})
		}
		arg.Reset()
		inArg = false
		argPlain = true
		argQuoted = false
	}

	inputRunes := []rune(input)
	for i := 0; i < len(inputRunes); i++ {
		r := inputRunes[i]
		switch {
		case isBackslash:
			arg.WriteRune(r)
			inArg = true
			argPlain = false
			isBackslash = false
		case r == '\'' && !isDoubleQuotes:
			isSingleQuotes = !isSingleQuotes
			inArg = true
			argPlain = false
			argQuoted = true
		case r == '"' && !isSingleQuotes:
			isDoubleQuotes = !isDoubleQuotes
			inArg = true
			argPlain = false
			argQuoted = true
		case r == '\\' && !isSingleQuotes:
			isBackslash = true
			argPlain = false
		case r == '$' && !isSingleQuotes && lookup != nil:
			name, end := readParam(inputRunes, i)
			if name == "" {
				arg.WriteRune('$')
				inArg = true
			} else {
				arg.WriteString(lookup(name))
				inArg = true
				argPlain = false
				i = end
			}
		case !isSingleQuotes && !isDoubleQuotes && unicode.IsSpace(r):
			flushArg()
		case !isSingleQuotes && !isDoubleQuotes && r == '|':
			flushArg()
			tokens = append(tokens, Token{Type: TokenPipe, Value: "|"})
		case !isSingleQuotes && !isDoubleQuotes && (r == '<' || r == '>'):
			redirectType := TokenRedirectOut
			// A directly adjacent, unquoted `1` or `2` is an explicit file
			// descriptor, not a command argument (for example, `echo 2>file`).
			if r == '>' && inArg && argPlain {
				switch arg.String() {
				case "1":
					redirectType = TokenRedirectOut
					arg.Reset()
					inArg = false
					argQuoted = false
				case "2":
					redirectType = TokenRedirectErr
					arg.Reset()
					inArg = false
					argQuoted = false
				}
			}
			if r == '>' && i+1 < len(inputRunes) && inputRunes[i+1] == '>' && redirectType == TokenRedirectErr {
				redirectType = TokenRedirectAppendErr
			}
			flushArg()

			switch r {
			case '<':
				tokens = append(tokens, Token{Type: TokenRedirectIn, Value: "<"})
			case '>':
				if i+1 < len(inputRunes) && inputRunes[i+1] == '>' {
					if redirectType == TokenRedirectAppendErr {
						tokens = append(tokens, Token{Type: TokenRedirectAppendErr, Value: "2>>"})
					} else {
						tokens = append(tokens, Token{Type: TokenRedirectAppend, Value: ">>"})
					}
					i++
				} else {
					tokens = append(tokens, Token{Type: redirectType, Value: ">"})
				}
			}
		default:
			arg.WriteRune(r)
			inArg = true
		}
	}

	flushArg()

	if isBackslash {
		return nil, errors.New("unterminated escape")
	}

	if isSingleQuotes {
		return nil, errors.New("unterminated single quote")
	}

	if isDoubleQuotes {
		return nil, errors.New("unterminated double quote")
	}

	return tokens, nil
}

// readParam читает $NAME или ${NAME}, начиная с '$' в runes[i].
// Возвращает имя и индекс последнего потреблённого руна.
// Если это не параметр — ("", i).
func readParam(runes []rune, i int) (name string, end int) {
	if i+1 >= len(runes) {
		return "", i
	}
	if runes[i+1] == '{' {
		j := i + 2
		for j < len(runes) && runes[j] != '}' {
			j++
		}
		if j >= len(runes) {
			return "", i
		}
		return string(runes[i+2 : j]), j
	}
	j := i + 1
	if !isParamStart(runes[j]) {
		return "", i
	}
	j++
	for j < len(runes) && isParamCont(runes[j]) {
		j++
	}
	return string(runes[i+1 : j]), j - 1
}

func isParamStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isParamCont(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func buildPipeline(tokens []Token) (*Pipeline, error) {
	if len(tokens) == 0 {
		return nil, errors.New("empty command")
	}

	background := false
	if n := len(tokens); tokens[n-1].Type == TokenArg && tokens[n-1].Value == "&" {
		if n == 1 {
			return nil, errors.New("empty command")
		}
		background = true
		tokens = tokens[:n-1]
	}

	var segments [][]Token
	var current []Token
	for _, token := range tokens {
		if token.Type == TokenPipe {
			if len(current) == 0 {
				return nil, errors.New("empty command in pipeline")
			}
			segments = append(segments, current)
			current = nil
			continue
		}
		current = append(current, token)
	}
	if len(current) == 0 {
		return nil, errors.New("empty command in pipeline")
	}
	segments = append(segments, current)

	commands := make([]*Command, 0, len(segments))
	for _, seg := range segments {
		cmd, err := buildCommand(seg)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}

	return &Pipeline{Commands: commands, Background: background}, nil
}

// buildCommand собирает токены одной команды (без `|` и завершающего `&`).
func buildCommand(tokens []Token) (*Command, error) {
	if len(tokens) == 0 {
		return nil, errors.New("empty command")
	}
	if tokens[0].Type != TokenArg {
		return nil, errors.New("command name must come before redirections")
	}

	cmd := &Command{
		Name: tokens[0].Value,
	}

	for i := 1; i < len(tokens); i++ {
		token := tokens[i]

		if token.Type == TokenArg {
			cmd.Args = append(cmd.Args, token.Value)
			continue
		}

		message := ""
		switch token.Type {
		case TokenRedirectOut:
			message = "missing file for redirection >"
		case TokenRedirectAppend:
			message = "missing file for redirection >>"
		case TokenRedirectIn:
			message = "missing file for redirection <"
		case TokenRedirectErr:
			message = "missing file for redirection 2>"
		case TokenRedirectAppendErr:
			message = "missing file for redirection 2>>"
		default:
			return nil, errors.New("unknown token type")
		}
		if i+1 >= len(tokens) || tokens[i+1].Type != TokenArg {
			return nil, errors.New(message)
		}

		i++
		cmd.Redirections = append(cmd.Redirections, Redirection{
			Type:     token.Type,
			FilePath: tokens[i].Value,
		})
	}

	return cmd, nil
}
