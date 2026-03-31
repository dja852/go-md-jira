package gomdjira

import (
	"os"
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

func TestConvertJiraString(t *testing.T) {
	jira := `h1. Header 1
h2. Header 2

This is *bold* text and _italic_ text.

Here is [link|https://example.com] and {{inline code}}.

- Unordered item
-- Nested unordered item
# Ordered item
## Nested ordered item

-strikethrough-

{code:go}
func main() {
    println("hello")
}
{code}`

	result := ConvertJiraString(jira)

	if !strings.Contains(result, "# Header 1") {
		t.Error("Header conversion failed")
	}
	if !strings.Contains(result, "## Header 2") {
		t.Error("Header level conversion failed")
	}
	if !strings.Contains(result, "**bold**") {
		t.Error("Bold conversion failed")
	}
	if !strings.Contains(result, "_italic_") {
		t.Error("Italic conversion failed")
	}
	if !strings.Contains(result, "[link](https://example.com)") {
		t.Error("Link conversion failed")
	}
	if !strings.Contains(result, "`inline code`") {
		t.Error("Inline code conversion failed")
	}
	if !strings.Contains(result, "- Unordered item") {
		t.Error("Unordered list conversion failed")
	}
	if !strings.Contains(result, "    - Nested unordered item") {
		t.Error("Nested unordered list conversion failed")
	}
	if !strings.Contains(result, "1. Ordered item") {
		t.Error("Ordered list conversion failed")
	}
	if !strings.Contains(result, "    1. Nested ordered item") {
		t.Error("Nested ordered list conversion failed")
	}
	if !strings.Contains(result, "~~strikethrough~~") {
		t.Error("Strikethrough conversion failed")
	}
	if !strings.Contains(result, "```go") {
		t.Error("Code block language conversion failed")
	}
}

func TestConvertJiraStringListBeforeHeaderParsing(t *testing.T) {
	jira := `# First ordered
## Nested ordered
h2. Actual heading`

	result := ConvertJiraString(jira)
	lines := strings.Split(result, "\n")

	if lines[0] != "1. First ordered" {
		t.Fatalf("expected first line to be ordered list, got %q", lines[0])
	}
	if lines[1] != "    1. Nested ordered" {
		t.Fatalf("expected second line to be nested ordered list, got %q", lines[1])
	}
	if lines[2] != "## Actual heading" {
		t.Fatalf("expected third line to be heading, got %q", lines[2])
	}
}

func TestConvertJiraFile(t *testing.T) {
	content := `h1. Jira Title

# Item one
## Item two

{code}
echo hi
{code}`

	tmp, err := os.CreateTemp("", "jira-input-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	result, err := ConvertJiraFile(tmp.Name())
	if err != nil {
		t.Fatalf("ConvertJiraFile failed: %v", err)
	}

	if !strings.Contains(result, "# Jira Title") {
		t.Error("File reverse conversion header failed")
	}
	if !strings.Contains(result, "1. Item one") {
		t.Error("File reverse conversion ordered list failed")
	}
	if !strings.Contains(result, "```") {
		t.Error("File reverse conversion code block failed")
	}

	_, err = ConvertJiraFile("nonexistent-jira-file.txt")
	if err == nil {
		t.Error("Expected error for non-existent Jira file")
	}
}

func TestConvertJiraStringInlineBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts bounded bold",
			input:    "This is *bold* text.",
			expected: "This is **bold** text.",
		},
		{
			name:     "converts bounded italic",
			input:    "This is _italic_ text.",
			expected: "This is _italic_ text.",
		},
		{
			name:     "converts bounded strike",
			input:    "This is -removed- text.",
			expected: "This is ~~removed~~ text.",
		},
		{
			name:     "does not convert hyphenated word",
			input:    "This is state-of-the-art output.",
			expected: "This is state-of-the-art output.",
		},
		{
			name:     "does not convert embedded bold markers",
			input:    "abc*not-bold*def",
			expected: "abc*not-bold*def",
		},
		{
			name:     "does not convert strike in numeric token",
			input:    "Build 2024-10-31 is stable.",
			expected: "Build 2024-10-31 is stable.",
		},
		{
			name:     "does not alter inline code payload",
			input:    "Use {{state-of-the-art}} and {{*literal*}}.",
			expected: "Use `state-of-the-art` and `*literal*`.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ConvertJiraString(tc.input)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
