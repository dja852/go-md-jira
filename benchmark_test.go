package gomdjira

import (
	"encoding/json"
	"strings"
	"testing"
)

// Sample markdown content for benchmarking
const sampleMarkdown = `# Header 1
## Header 2
### Header 3

This is **bold** text and *italic* text.

Here's a [link](https://example.com) and some ` + "`inline code`" + `.

~~strikethrough~~ text.

1. Ordered list item 1
2. Ordered list item 2

* Unordered list item 1
* Unordered list item 2

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

    // This is indented code
    var x = 42

More regular text here.
`

func BenchmarkConvertMarkdownString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ConvertMarkdownString(sampleMarkdown)
	}
}

func BenchmarkConvertMarkdownStringLarge(b *testing.B) {
	// Create a large markdown document by repeating the sample
	largeMarkdown := strings.Repeat(sampleMarkdown, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertMarkdownString(largeMarkdown)
	}
}

func TestConvertMarkdownString(t *testing.T) {
	result := ConvertMarkdownString(sampleMarkdown)

	t.Logf("Result:\n%s", result)

	// Basic smoke test to ensure conversion works
	if !strings.Contains(result, "h1. Header 1") {
		t.Error("Header 1 conversion failed")
	}
	if !strings.Contains(result, "*bold*") {
		t.Errorf("Bold text conversion failed. Result: %s", result)
	}
	if !strings.Contains(result, "_italic_") {
		t.Errorf("Italic text conversion failed. Result: %s", result)
	}
	if !strings.Contains(result, "[link|https://example.com]") {
		t.Error("Link conversion failed")
	}
	if !strings.Contains(result, "{{inline code}}") {
		t.Error("Inline code conversion failed")
	}
}

func TestInlineCodeInLists(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"* Item with `*asterisk*` in code",
			"- Item with {{*asterisk*}} in code",
		},
		{
			"* Another with `_underscore_` text",
			"- Another with {{_underscore_}} text",
		},
		{
			"1. Ordered with `**bold**` code",
			"# Ordered with {{**bold**}} code",
		},
		{
			"- List with `code` and **actual bold**",
			"- List with {{code}} and *actual bold*",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ConvertMarkdownString(test.input)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestConvertMarkdownFile(t *testing.T) {
	// Test with the existing test.md file
	result, err := ConvertMarkdownFile("test.md")
	if err != nil {
		t.Fatalf("ConvertMarkdownFile failed: %v", err)
	}

	// Basic checks to ensure conversion worked
	if !strings.Contains(result, "{{inline code}}") {
		t.Error("Inline code conversion failed in file conversion")
	}
	if !strings.Contains(result, "{code:json}") {
		t.Error("Code block conversion failed in file conversion")
	}
	if !strings.Contains(result, "*If you also have access within this group, please comment to review all other access") {
		t.Error("Bold text conversion failed in file conversion")
	}

	// Test with non-existent file
	_, err = ConvertMarkdownFile("nonexistent.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConvertMarkdownStringToADF(t *testing.T) {
	markdown := `# Header 1
## Header 2

This is **bold** text and *italic* text.

Here's a [link](https://example.com) and some ` + "`inline code`" + `.

- List item 1
- List item 2

1. Ordered item 1
2. Ordered item 2

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```"

	adf, err := ConvertMarkdownStringToADF(markdown)
	if err != nil {
		t.Fatalf("ConvertMarkdownStringToADF failed: %v", err)
	}

	// Basic structure checks
	if adf.Version != 1 {
		t.Error("ADF version should be 1")
	}
	if adf.Type != "doc" {
		t.Error("ADF type should be 'doc'")
	}
	if len(adf.Content) == 0 {
		t.Error("ADF content should not be empty")
	}

	// Check for headers
	hasH1 := false
	hasH2 := false
	for _, node := range adf.Content {
		if node.Type == "heading" {
			if level, ok := node.Attrs["level"].(int); ok {
				if level == 1 {
					hasH1 = true
				}
				if level == 2 {
					hasH2 = true
				}
			}
		}
	}
	if !hasH1 {
		t.Error("Should have H1 header")
	}
	if !hasH2 {
		t.Error("Should have H2 header")
	}
}

func TestConvertMarkdownStringToADFJSON(t *testing.T) {
	markdown := `# Test Header

This is a **test** with ` + "`code`" + `.`

	jsonResult, err := ConvertMarkdownStringToADFJSON(markdown)
	if err != nil {
		t.Fatalf("ConvertMarkdownStringToADFJSON failed: %v", err)
	}

	// Basic JSON structure checks
	if !strings.Contains(jsonResult, `"version": 1`) {
		t.Error("JSON should contain version")
	}
	if !strings.Contains(jsonResult, `"type": "doc"`) {
		t.Error("JSON should contain doc type")
	}
	if !strings.Contains(jsonResult, `"type": "heading"`) {
		t.Error("JSON should contain heading")
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonResult), &parsed); err != nil {
		t.Errorf("Result should be valid JSON: %v", err)
	}
}
