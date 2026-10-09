package main

import (
	"os"

	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	// 1. Создаем экземпляр шелла, внедряя зависимости (потоки ввода-вывода)
	sh := shell.New(os.Stdout, os.Stderr, os.Stdin)

	// 2. Запускаем основной цикл.
	// Весь дальнейший контроль переходит к shell.go
	sh.Run()
}
