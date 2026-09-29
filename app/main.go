package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

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
		if command == "exit" {
			break
		}
		fmt.Println(command[:len(command)-1] + ": command not found")
	}
}
