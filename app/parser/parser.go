package parser

import (
	"errors"
	"strings"
	"unicode"
)

// Parse парсит строку и возвращает структуру Command
func Parse(input string) (*Command, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	return buildCommand(tokens)
}

func tokenize(input string) ([]Token, error) {
	var tokens []Token
	var arg strings.Builder
	isSingleQuotes := false
	isDoubleQuotes := false
	isBackslash := false
	inArg := false
	argPlain := true

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
		case r == '"' && !isSingleQuotes:
			isDoubleQuotes = !isDoubleQuotes
			inArg = true
			argPlain = false
		case r == '\\' && !isSingleQuotes:
			isBackslash = true
			argPlain = false
		case !isSingleQuotes && !isDoubleQuotes && unicode.IsSpace(r):
			if inArg {
				tokens = append(tokens, Token{Type: TokenArg, Value: arg.String()})
				arg.Reset()
				inArg = false
				argPlain = true
			}
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
				case "2":
					redirectType = TokenRedirectErr
					arg.Reset()
					inArg = false
				}
			}
			if r == '>' && i+1 < len(inputRunes) && inputRunes[i+1] == '>' && redirectType == TokenRedirectErr {
				redirectType = TokenRedirectAppendErr
			}
			if inArg {
				tokens = append(tokens, Token{Type: TokenArg, Value: arg.String()})
				arg.Reset()
				inArg = false
			}

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

	if inArg {
		tokens = append(tokens, Token{Type: TokenArg, Value: arg.String()})
	}

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

// buildCommand собирает токены в структуру Command
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
