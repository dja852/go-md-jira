# go-md-jira

A Go library and command-line tool for converting between Markdown and Jira wiki markup. Inspired by https://github.com/eshack94/md-to-jira

## Features

- **Fast Performance**: Pre-compiled regex patterns for optimal conversion speed
- **Comprehensive Conversion**: Headers, lists, text formatting, code blocks, and links
- **Bidirectional Conversion**: Markdown -> Jira wiki and Jira wiki -> Markdown
- **Smart Protection**: Placeholder-based system prevents formatting conflicts
- **Nested Lists**: Proper handling of multi-level ordered and unordered lists
- **Code Block Support**: Fenced code blocks with optional language specification
- **CLI and Library**: Use as standalone tool or import into your Go projects

## Quick Start

### Command Line Usage

```bash
# Clone the repository
git clone https://github.com/dja852/go-md-jira.git
cd go-md-jira

# Convert a markdown file to Jira wiki (default direction)
go run ./cmd input.md

# Convert Jira wiki to Markdown
go run ./cmd -d jira-to-md input.jira

# Or build and run
go build -o md2jira ./cmd
./md2jira input.md
./md2jira -d jira-to-md input.jira
```

### Library Usage

```go
package main

import (
    "fmt"
    "log"
    
    gomdjira "github.com/dja852/go-md-jira"
)

func main() {
    // Convert markdown string directly
    markdown := `# Header
    
This is **bold** and _italic_ text with {{inline code}}.

- List item 1
- List item 2`

    jira := gomdjira.ConvertMarkdownString(markdown)
    fmt.Println(jira)
    
    // Convert markdown file and get result as string
    result, err := gomdjira.ConvertMarkdownFile("input.md")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
    
    // Convert markdown file and print to stdout
    if err := gomdjira.MarkdownToJira("input.md"); err != nil {
        log.Fatal(err)
    }

    // Convert Jira wiki string directly
    jiraInput := "h1. Header\n\n# Ordered item"
    md := gomdjira.ConvertJiraString(jiraInput)
    fmt.Println(md)

    // Convert Jira wiki file and print Markdown to stdout
    if err := gomdjira.JiraToMarkdown("input.jira"); err != nil {
        log.Fatal(err)
    }
}
```

### Installation as Dependency

```bash
go get github.com/dja852/go-md-jira
```

## Conversion Examples

| Markdown | Jira Markup | Description |
|----------|-------------|-------------|
| `# Header 1` | `h1. Header 1` | Headers (h1-h6) |
| `**bold**` | `*bold*` | Bold text |
| `_italic_` | `_italic_` | Italic text |
| `` `code` `` | `{{code}}` | Inline code |
| `~~strike~~` | `-strike-` | Strikethrough |
| `[link](url)` | `[link\|url]` | Links |
| `1. Item` | `# Item` | Ordered lists |
| `- Item` | `- Item` | Unordered lists |
| ` - Nested` | `-- Nested` | Nested lists (4-space indentation) |

A single blank line between Markdown list items is omitted from Jira output so the list remains continuous. Two or more blank lines are preserved as an intentional list break.

### Complex Example

**Markdown Input:**
````markdown
# Project Documentation

This project uses **Go** and supports _multiple formats_.

## Features
1. Fast conversion with `pre-compiled` regex
2. Support for nested lists:
    - Feature A
    - Feature B
        1. Sub-feature 1
        2. Sub-feature 2

```go
func main() {
    fmt.Println("Hello, World!")
}
```

Visit [GitHub](https://github.com/example/repo) for more info.
````

**Jira Output:**
```
h1. Project Documentation

This project uses *Go* and supports _multiple formats_.

h2. Features
# Fast conversion with {{pre-compiled}} regex
# Support for nested lists:
-- Feature A
-- Feature B
### Sub-feature 1
### Sub-feature 2

{code:go}
func main() {
    fmt.Println("Hello, World!")
}
{code}

Visit [GitHub|https://github.com/example/repo] for more info.
```

## API Reference

### Core Functions

#### `ConvertMarkdownString(markdown string) string`
Converts a markdown string directly to Jira markup.

**Parameters:**
- `markdown`: Input markdown text

**Returns:**
- Converted Jira markup string

#### `ConvertMarkdownFile(filePath string) (string, error)`
Reads a markdown file and returns the converted Jira markup as a string.

**Parameters:**
- `filePath`: Path to the markdown file

**Returns:**
- Converted Jira markup string and error if file operations fail

#### `MarkdownToJira(filePath string) error`
Reads a markdown file and prints the converted Jira markup to stdout.

**Parameters:**
- `filePath`: Path to the markdown file

**Returns:**
- Error if file operations fail, nil on success

#### `MarkdownToJiraWriter(inputPath string, writer io.Writer) error`
Converts a markdown file and writes output to a custom writer.

**Parameters:**
- `inputPath`: Path to the markdown file
- `writer`: io.Writer to receive the output

**Returns:**
- Error if file operations fail, nil on success

#### `ConvertJiraString(jira string) string`
Converts a Jira wiki string directly to Markdown.

#### `ConvertJiraFile(filePath string) (string, error)`
Reads a Jira wiki file and returns converted Markdown as a string.

#### `JiraToMarkdown(filePath string) error`
Reads a Jira wiki file and prints converted Markdown to stdout.

#### `JiraToMarkdownWriter(inputPath string, writer io.Writer) error`
Converts a Jira wiki file and writes Markdown to a custom writer.

## Regular Expression Patterns

The converter uses pre-compiled regex patterns for optimal performance:

### Header Patterns
```go
// Matches markdown headers and converts to Jira format
^######\s*(.+)  → h6. ${1}  // # Header 6
^#####\s*(.+)   → h5. ${1}  // # Header 5
^####\s*(.+)    → h4. ${1}  // # Header 4
^###\s*(.+)     → h3. ${1}  // # Header 3
^##\s*(.+)      → h2. ${1}  // # Header 2
^#\s*(.+)       → h1. ${1}  // # Header 1
```

### Text Formatting Patterns
```go
\*\*(.+?)\*\*   // **bold** → *bold*
__(.+?)__       // __bold__ → *bold*
\*([^*]+)\*     // *italic* → _italic_
_(.+?)_         // _italic_ → _italic_
`([^`]+)`       // `code` → {{code}}
~~(.+?)~~       // ~~strike~~ → -strike-
```

### Link Pattern
```go
\[(.*?)\]\((.+?)\)  // [text](url) → [text|url]
```

### List Patterns
```go
^\d+\.\s+                    // "1. Item" → "# Item"
^\*\s+                       // "* Item" → "- Item"
^( {2,})\d+\.\s+            // Nested ordered lists
^( {2,})[*-]\s+             // Nested unordered lists
```

### Code Block Pattern
```go
(?s)```(\w+)?\n(.*?)\n```   // Fenced code blocks with optional language
```

## Conversion Logic

### Processing Flow

1. **Placeholder Protection**: Inline code and bold text are temporarily replaced with placeholders (`§INLINECODE§`, `§BOLD§`) to prevent interference with other patterns

2. **Header Processing**: Headers are processed first as they take precedence over other formatting

3. **List Processing**: 
   - Nested lists are calculated based on 4-space indentation levels
   - Ordered lists use `#`, `##`, `###` for nesting
   - Unordered lists use `-`, `--`, `---` for nesting

4. **Text Formatting**: Bold, italic, strikethrough, and links are processed

5. **Code Block Processing**: Fenced code blocks are converted to Jira `{code:lang}` format

6. **Placeholder Restoration**: Original content is restored from placeholders

### Nested List Level Calculation

The converter uses 4-space indentation to determine nesting levels:

```go
level := (len(spaces) / 4) + 1  // 4 spaces = 1 level deeper
```

Examples:
- `    - Item` (4 spaces) → `-- Item` (level 2)
- `        1. Item` (8 spaces) → `### Item` (level 3)
- `            - Item` (12 spaces) → `---- Item` (level 4)

### Performance Optimizations

1. **Pre-compiled Regex**: All patterns are compiled once at package initialization
2. **Placeholder System**: Prevents multiple passes over the same content
3. **Single-pass Processing**: Most conversions happen in a single iteration
4. **Efficient String Operations**: Uses `strings.Builder` for concatenation

## Testing

Run the test suite:

```bash
go test -v
```

Run performance benchmarks:

```bash
go test -bench=.
```

Current benchmark results:
```
BenchmarkConvertMarkdownString-8    50000    ~30000 ns/op
```

## Architecture

### Package Structure
```
go-md-jira/
├── go-md-jira.go        # Core conversion logic
├── benchmark_test.go    # Tests and benchmarks
├── cmd/
│   └── converter.go     # Command-line interface
├── go.mod              # Go module definition
└── README.md           # This documentation
```

### Design Decisions

1. **No Indented Code Blocks**: Disabled to avoid conflicts with nested lists (architectural choice for clarity)
2. **Placeholder Protection**: Prevents nested formatting issues with complex markdown
3. **Pre-compiled Regex**: Performance optimization for repeated conversions
4. **Single Package**: Simple, focused API surface

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details.

## Performance Notes

- Pre-compiled regex patterns provide ~10x performance improvement over on-demand compilation
- Placeholder protection system adds minimal overhead while preventing conversion errors
- Memory efficient with string builders and single-pass processing
- Suitable for high-throughput conversion scenarios