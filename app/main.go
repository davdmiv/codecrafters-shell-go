package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)
	builtins := []string{"echo", "exit", "type"}

	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error while reading: ", err.Error())
			os.Exit(1)
		}

		command = strings.TrimSpace(command)
		tokens := strings.Split(command, " ")

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
			fmt.Println(command[5:])
		} else if path, err := exec.LookPath(baseCmd); err == nil {
			var cmd *exec.Cmd
			cmd = exec.Command(path, tokens[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
			fmt.Println(string(output))
		} else {
			fmt.Println(command + ": command not found")
		}
	}
}
