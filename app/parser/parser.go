package parser

import (
	"strings"
	"unicode"
)

func SplitArgs(input string) ([]string, error) {
	var args []string
	var arg strings.Builder
	isSingleQuotes := false
	isDoubleQuotes := false
	isBackslash := false
	inArg := false

	for _, r := range input {
		switch {
		case isBackslash:
			arg.WriteRune(r)
			inArg = true
			isBackslash = false
		case r == '\'' && !isDoubleQuotes:
			isSingleQuotes = !isSingleQuotes
			inArg = true
		case r == '"' && !isSingleQuotes:
			isDoubleQuotes = !isDoubleQuotes
			inArg = true
		case r == '\\' && !isSingleQuotes:
			isBackslash = true
		case !isSingleQuotes && !isDoubleQuotes && unicode.IsSpace(r):
			if inArg {
				args = append(args, arg.String())
				arg.Reset()
				inArg = false
			}
		default:
			arg.WriteRune(r)
			inArg = true
		}
	}

	if inArg {
		args = append(args, arg.String())
	}

	return args, nil
}
