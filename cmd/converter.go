package main

import (
	"fmt"
	"os"

	gomdjira "github.com/dja852/go-md-jira"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file_path>")
		os.Exit(1)
	}
	filePath := os.Args[1]
	if err := gomdjira.MarkdownToJira(filePath); err != nil { // Use the function from the package
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
