package main

import (
	"fmt"
	"os"

	"example.com/myapp/internal"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Ожидаемый ввод: %s <file_path> <name> <role> <pasd56rresword>\n", os.Args[0])
		os.Exit(1)
	}
	file_path, name, role, password := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	internal.WritePassword(name, role, password, file_path)
}
