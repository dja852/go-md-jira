# go-md-jira

A high-performance Go library and command-line tool for converting Markdown to Jira markup format and Atlassian Document Format (ADF). 

## Features

- **Fast Performance**: Pre-compiled regex patterns for optimal conversion speed
- **Multiple Output Formats**: Jira markup and Atlassian Document Format (ADF) JSON
- **Comprehensive Conversion**: Headers, lists, text formatting, code blocks, and links
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

# Convert to Jira markup (default)
go run ./cmd input.md
go run ./cmd -format=jira input.md

# Convert to Atlassian Document Format (ADF) JSON
go run ./cmd -format=adf input.md
go run ./cmd -f adf input.md

# Show help
go run ./cmd --help

# Or build and run
go build -o md2jira ./cmd
./md2jira input.md                    # Jira format (default)
./md2jira -f adf input.md             # ADF format
```

#### CLI Options

- `-format` or `-f`: Output format (`jira` or `adf`, default: `jira`)
- `-help` or `-h`: Show help message

#### CLI Examples

```bash
# Convert to Jira markup
./md2jira document.md
./md2jira -format=jira document.md

# Convert to ADF JSON
./md2jira -format=adf document.md  
./md2jira -f adf document.md

# Redirect output to file
./md2jira document.md > output.jira
./md2jira -f adf document.md > output.json
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
    
    // Convert to Atlassian Document Format (ADF)
    adfJSON, err := gomdjira.ConvertMarkdownStringToADFJSON(markdown)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("ADF JSON:", adfJSON)
    
    // Convert file to ADF structure
    adf, err := gomdjira.ConvertMarkdownFileToADF("input.md")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("ADF Document: %+v\n", adf)
}
```

### Installation as Dependency

```bash
go get github.com/dja852/go-md-jira
```

## Conversion Examples

### Jira Markup Output

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

### Atlassian Document Format (ADF) Output

The same markdown also converts to ADF JSON format:

**ADF JSON Output (excerpt):**
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "heading",
      "attrs": { "level": 1 },
      "content": [
        { "type": "text", "text": "Project Documentation" }
      ]
    },
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "This project uses " },
        { "type": "text", "text": "Go", "marks": [{"type": "strong"}] },
        { "type": "text", "text": " and supports " },
        { "type": "text", "text": "multiple formats", "marks": [{"type": "em"}] }
      ]
    },
    {
      "type": "codeBlock",
      "content": [
        { "type": "text", "text": "func main() {\n    fmt.Println(\"Hello, World!\")\n}" }
      ]
    }
  ]
}
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

### Atlassian Document Format (ADF) Functions

#### `ConvertMarkdownStringToADF(markdown string) (*ADFDocument, error)`
Converts a markdown string to Atlassian Document Format (ADF) structure.

**Parameters:**
- `markdown`: Input markdown text

**Returns:**
- ADF document structure and error if conversion fails

#### `ConvertMarkdownStringToADFJSON(markdown string) (string, error)`
Converts a markdown string to ADF JSON format.

**Parameters:**
- `markdown`: Input markdown text

**Returns:**
- JSON string in ADF format and error if conversion fails

#### `ConvertMarkdownFileToADF(filePath string) (*ADFDocument, error)`
Reads a markdown file and converts it to ADF structure.

**Parameters:**
- `filePath`: Path to the markdown file

**Returns:**
- ADF document structure and error if file operations fail

#### `ConvertMarkdownFileToADFJSON(filePath string) (string, error)`
Reads a markdown file and converts it to ADF JSON format.

**Parameters:**
- `filePath`: Path to the markdown file

**Returns:**
- JSON string in ADF format and error if file operations fail

#### `MarkdownToADFWriter(filePath string, writer io.Writer) error`
Converts a markdown file to ADF JSON and writes to the specified writer.

**Parameters:**
- `filePath`: Path to the markdown file
- `writer`: io.Writer to receive the ADF JSON output

**Returns:**
- Error if file operations fail, nil on success

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