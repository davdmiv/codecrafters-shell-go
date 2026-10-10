package main

import (
	"os"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	sh := shell.New(os.Stdout, os.Stderr, os.Stdin)

	os.Exit(sh.Run())
}
