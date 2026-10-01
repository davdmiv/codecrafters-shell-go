package main

import (
	"bufio"
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
		tokens := splitArgs(command)

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
			fmt.Println(strings.Join(splitArgs(command[5:]), " "))
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

func splitArgs(cmdline string) []string {
	var args []string

	var arg strings.Builder
	var parenthesis = '\x00'

	for _, r := range cmdline {

		if (parenthesis == '\x00' || parenthesis == r) && (r == '\'' || r == '"') {
			if parenthesis == '\x00' {
				parenthesis = r
			} else {
				parenthesis = '\x00'
			}
			continue
		}

		if parenthesis == '\x00' && unicode.IsSpace(r) {
			if arg.Len() != 0 {
				args = append(args, arg.String())
				arg.Reset()
			}
			continue
		}

		arg.WriteRune(r)
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	return args
}
