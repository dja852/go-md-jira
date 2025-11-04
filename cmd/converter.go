package main

import (
	"flag"
	"fmt"
	"os"

	gomdjira "github.com/dja852/go-md-jira"
)

func main() {
	var format string
	var showHelp bool

	flag.StringVar(&format, "format", "jira", "Output format: 'jira' for Jira markup or 'adf' for Atlassian Document Format JSON")
	flag.StringVar(&format, "f", "jira", "Output format (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <file_path>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert Markdown files to Jira markup or Atlassian Document Format (ADF).\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s input.md                    # Convert to Jira markup (default)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format=jira input.md       # Convert to Jira markup\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format=adf input.md        # Convert to ADF JSON\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f adf input.md             # Convert to ADF JSON (shorthand)\n", os.Args[0])
	}

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: Missing required argument <file_path>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	// Validate format
	if format != "jira" && format != "adf" {
		fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'. Must be 'jira' or 'adf'\n", format)
		os.Exit(1)
	}

	// Convert based on format
	switch format {
	case "jira":
		if err := gomdjira.MarkdownToJira(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "adf":
		if err := gomdjira.MarkdownToADFWriter(filePath, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
