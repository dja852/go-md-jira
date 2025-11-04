package gomdjira

import (
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
