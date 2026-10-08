package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

type handler func([]string)

var builtins = map[string]handler{
	"echo": handleEcho,
	"exit": handleExit,
	"type": handleType,
	"pwd":  handlePwd,
	"cd":   handleCd,
}

func main() {
	reader := bufio.NewReader(os.Stdin)

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

		if _, ok := builtins[baseCmd]; ok {
			builtins[baseCmd](tokens)
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

func handleExit(args []string) {
	os.Exit(0)
}

func handleEcho(args []string) {
	fmt.Println(strings.Join(args[1:], " "))
}

func handlePwd(args []string) {
	dir, _ := os.Getwd()
	fmt.Println(dir)
}

func handleType(args []string) {
	var target string
	if len(args) > 1 {
		target = args[1]
	}

	if _, ok := builtins[target]; ok {
		fmt.Println(target + " is a shell builtin")
	} else if path, err := exec.LookPath(target); err == nil {
		fmt.Printf("%s is %s\n", target, path)
	} else {
		fmt.Println(target + ": not found")
	}
}

func handleCd(args []string) {
	var target string
	if len(args) > 1 {
		target = args[1]
	}

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
