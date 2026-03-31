package main

import (
	"flag"
	"fmt"
	"os"

	gomdjira "github.com/dja852/go-md-jira"
)

func main() {
	var direction string

	flag.StringVar(&direction, "direction", "md-to-jira", "Conversion direction: 'md-to-jira' or 'jira-to-md'")
	flag.StringVar(&direction, "d", "md-to-jira", "Conversion direction (shorthand)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run ./cmd [OPTIONS] <file_path>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	switch direction {
	case "md-to-jira":
		if err := gomdjira.MarkdownToJira(filePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "jira-to-md":
		if err := gomdjira.JiraToMarkdown(filePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Error: invalid direction '%s'. Must be 'md-to-jira' or 'jira-to-md'\n", direction)
		os.Exit(1)
	}
}
