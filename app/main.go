package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)
	builtins := []string{"echo", "exit", "type", "pwd", "cd"}

	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error while reading: ", err.Error())
			os.Exit(1)
		}

		command = strings.TrimSpace(command)

		tokens, err := splitArgs(command)
		if err != nil {
			fmt.Println(err)
			continue
		}

		baseCmd := tokens[0]

		var target string
		if len(tokens) > 1 {
			target = tokens[1]
		}

		if baseCmd == "type" {
			if slices.Contains(builtins, target) {
				fmt.Println(target + " is a shell builtin")
			} else if path, err := exec.LookPath(target); err == nil {
				fmt.Printf("%s is %s\n", target, path)
			} else {
				fmt.Println(target + ": not found")
			}
		} else if baseCmd == "exit" {
			break
		} else if baseCmd == "echo" {
			tokens, err = splitArgs(command[5:])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(strings.Join(tokens, " "))
		} else if baseCmd == "pwd" {
			dir, _ := os.Getwd()
			fmt.Println(dir)
		} else if baseCmd == "cd" {
			if target == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					return
				}
				target = home
			}

			path := filepath.Clean(target)

			if err := os.Chdir(path); err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", path)
			}
		} else if _, err := exec.LookPath(baseCmd); err == nil {
			cmd := exec.Command(baseCmd, tokens[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Run()
		} else {
			fmt.Println(command + ": command not found")
		}
	}
}

func splitArgs(input string) ([]string, error) {
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
		case r == '\\' && !isDoubleQuotes && !isSingleQuotes:
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

	if isBackslash {
		return nil, errors.New("unterminated escape")
	}

	if isSingleQuotes {
		return nil, errors.New("unterminated single quote")
	}

	if isDoubleQuotes {
		return nil, errors.New("unterminated double quote")
	}

	if inArg {
		args = append(args, arg.String())
	}

	return args, nil
}
