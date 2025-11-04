package gomdjira

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Pre-compiled regular expressions for performance
var (
	// Header patterns
	headerRegexes = []*regexp.Regexp{
		regexp.MustCompile(`^######\s*(.+)`),
		regexp.MustCompile(`^#####\s*(.+)`),
		regexp.MustCompile(`^####\s*(.+)`),
		regexp.MustCompile(`^###\s*(.+)`),
		regexp.MustCompile(`^##\s*(.+)`),
		regexp.MustCompile(`^#\s*(.+)`),
	}
	headerReplacements = []string{"h6. ${1}", "h5. ${1}", "h4. ${1}", "h3. ${1}", "h2. ${1}", "h1. ${1}"}

	// Text formatting patterns
	boldRegex1   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	boldRegex2   = regexp.MustCompile(`__(.+?)__`)
	italicRegex1 = regexp.MustCompile(`\*([^*]+)\*`)
	italicRegex2 = regexp.MustCompile(`_(.+?)_`)

	// Other formatting patterns
	inlineCodeRegex    = regexp.MustCompile("`([^`]+)`")
	strikethroughRegex = regexp.MustCompile(`~~(.+?)~~`)
	linkRegex          = regexp.MustCompile(`\[(.*?)\]\((.+?)\)`)
	orderedListRegex   = regexp.MustCompile(`^\d+\.\s+`)
	unorderedListRegex = regexp.MustCompile(`^\*\s+`)

	// Nested list patterns
	nestedOrderedListRegex   = regexp.MustCompile(`^( {2,})\d+\.\s+`)
	nestedUnorderedListRegex = regexp.MustCompile(`^( {2,})[*-]\s+`)

	// Fenced code block pattern
	fencedCodeRegex = regexp.MustCompile("(?s)```(\\w+)?\\n(.*?)\\n```")
)

// convertLine converts a single markdown line to Jira markup.
func convertLine(line string) string {
	// Headers are processed first and return early if matched
	for i, regex := range headerRegexes {
		if regex.MatchString(line) {
			return regex.ReplaceAllString(line, headerReplacements[i])
		}
	}

	// Protect inline code from other formatting by using placeholders
	const inlineCodePlaceholder = "§INLINECODE§"
	var inlineCodeBlocks []string
	line = inlineCodeRegex.ReplaceAllStringFunc(line, func(match string) string {
		codeContent := inlineCodeRegex.FindStringSubmatch(match)[1]
		inlineCodeBlocks = append(inlineCodeBlocks, "{{"+codeContent+"}}")
		return inlineCodePlaceholder + fmt.Sprintf("%d", len(inlineCodeBlocks)-1) + inlineCodePlaceholder
	})

	// Bold formatting with placeholder to avoid conflicts with italic
	const boldPlaceholder = "§BOLD§"
	line = boldRegex1.ReplaceAllString(line, boldPlaceholder+"${1}"+boldPlaceholder)
	line = boldRegex2.ReplaceAllString(line, boldPlaceholder+"${1}"+boldPlaceholder)

	// Italic formatting
	line = italicRegex1.ReplaceAllString(line, "_${1}_")
	line = italicRegex2.ReplaceAllString(line, "_${1}_")

	// Restore bold formatting
	line = strings.ReplaceAll(line, boldPlaceholder, "*")

	// Other formatting
	line = strikethroughRegex.ReplaceAllString(line, "-${1}-")
	line = linkRegex.ReplaceAllString(line, "[${1}|${2}]")

	// List processing (nested lists first, then top-level)
	line = nestedOrderedListRegex.ReplaceAllStringFunc(line, func(match string) string {
		spaces := nestedOrderedListRegex.FindStringSubmatch(match)[1]
		level := len(spaces)/4 + 1 // 4 spaces per level, +1 for base level
		if level > 6 {
			level = 6
		}
		return strings.Repeat("#", level) + " "
	})
	line = nestedUnorderedListRegex.ReplaceAllStringFunc(line, func(match string) string {
		spaces := nestedUnorderedListRegex.FindStringSubmatch(match)[1]
		level := len(spaces)/4 + 1 // 4 spaces per level, +1 for base level
		if level > 6 {
			level = 6
		}
		return strings.Repeat("-", level) + " "
	})

	line = orderedListRegex.ReplaceAllString(line, "# ")
	line = unorderedListRegex.ReplaceAllString(line, "- ")

	// Restore protected inline code
	for i, codeBlock := range inlineCodeBlocks {
		placeholder := inlineCodePlaceholder + fmt.Sprintf("%d", i) + inlineCodePlaceholder
		line = strings.ReplaceAll(line, placeholder, codeBlock)
	}

	return line
}

// convertMultilineElements converts fenced code blocks.
func convertMultilineElements(content string) string {
	content = fencedCodeRegex.ReplaceAllStringFunc(content, processCodeBlock)
	return content
}

// processCodeBlock converts fenced code blocks to Jira format.
func processCodeBlock(match string) string {
	matches := fencedCodeRegex.FindStringSubmatch(match)
	if len(matches) < 3 {
		return match // Safety fallback
	}

	lang := matches[1] // Optional language
	code := matches[2] // Code content

	if lang != "" {
		return fmt.Sprintf("{code:%s}\n%s\n{code}", lang, code)
	}
	return fmt.Sprintf("{code}\n%s\n{code}", code)
}

// ConvertMarkdownString converts a markdown string to Jira markup.
func ConvertMarkdownString(markdown string) string {
	convertedContent := convertMultilineElements(markdown)
	lines := strings.Split(convertedContent, "\n")
	jiraLines := make([]string, 0, len(lines))

	inCodeBlock := false
	for _, line := range lines {
		// Track code block boundaries to avoid processing content inside them
		if strings.HasPrefix(line, "{code") && (strings.HasSuffix(line, "}") || line == "{code}") {
			if line == "{code}" {
				inCodeBlock = false
			} else {
				inCodeBlock = true
			}
			jiraLines = append(jiraLines, line)
			continue
		}

		// Only apply line conversions if we're not in a code block
		if !inCodeBlock {
			line = convertLine(line)
		}
		jiraLines = append(jiraLines, line)
	}

	return strings.Join(jiraLines, "\n")
}

// ConvertMarkdownFile reads a markdown file and returns the converted Jira markup as a string.
func ConvertMarkdownFile(filePath string) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}

	return ConvertMarkdownString(string(contentBytes)), nil
}

// MarkdownToJira reads a markdown file, converts its content to Jira markup, and prints it.
func MarkdownToJira(filePath string) error {
	return MarkdownToJiraWriter(filePath, os.Stdout)
}

// MarkdownToJiraWriter converts a markdown file to Jira markup and writes to the specified writer.
func MarkdownToJiraWriter(filePath string, writer io.Writer) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	result := ConvertMarkdownString(string(contentBytes))
	_, err = fmt.Fprintln(writer, result)
	if err != nil {
		return fmt.Errorf("could not write output: %w", err)
	}

	return nil
}

// ADF (Atlassian Document Format) structures
type ADFDocument struct {
	Version int       `json:"version"`
	Type    string    `json:"type"`
	Content []ADFNode `json:"content"`
}

type ADFNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []ADFNode              `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Marks   []ADFMark              `json:"marks,omitempty"`
}

type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// convertFencedCodeBlocksToADF processes fenced code blocks and adds them to ADF document.
func convertFencedCodeBlocksToADF(content string, doc *ADFDocument) string {
	return content // Pass through without modification - we'll handle in main loop
}

// ConvertMarkdownStringToADF converts a markdown string to Atlassian Document Format (ADF).
func ConvertMarkdownStringToADF(markdown string) (*ADFDocument, error) {
	doc := &ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []ADFNode{},
	}

	// Split content into lines for processing
	lines := strings.Split(markdown, "\n")

	var currentParagraph *ADFNode
	var inCodeBlock bool
	var codeBlockLines []string
	var codeBlockLang string

	for _, line := range lines {
		// Check for fenced code block start
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockLang = strings.TrimSpace(line[3:]) // Extract language
				codeBlockLines = []string{}

				// Finish any current paragraph
				if currentParagraph != nil && len(currentParagraph.Content) > 0 {
					doc.Content = append(doc.Content, *currentParagraph)
					currentParagraph = nil
				}
				continue
			} else {
				// Ending a code block
				inCodeBlock = false

				// Create ADF code block
				codeBlock := ADFNode{
					Type: "codeBlock",
					Content: []ADFNode{
						{
							Type: "text",
							Text: strings.Join(codeBlockLines, "\n"),
						},
					},
				}

				// Add language attribute if present
				if codeBlockLang != "" {
					codeBlock.Attrs = map[string]interface{}{
						"language": codeBlockLang,
					}
				}

				doc.Content = append(doc.Content, codeBlock)
				codeBlockLines = nil
				codeBlockLang = ""
				continue
			}
		}

		if inCodeBlock {
			// Collect code block content
			codeBlockLines = append(codeBlockLines, line)
			continue
		}

		// Process regular content
		line = strings.TrimSpace(line)

		// Handle empty lines
		if line == "" {
			if currentParagraph != nil && len(currentParagraph.Content) > 0 {
				doc.Content = append(doc.Content, *currentParagraph)
				currentParagraph = nil
			}
			continue
		}

		// Convert line to ADF nodes
		node := convertLineToADF(line)
		if node != nil {
			if node.Type == "paragraph" {
				if currentParagraph != nil && len(currentParagraph.Content) > 0 {
					doc.Content = append(doc.Content, *currentParagraph)
				}
				doc.Content = append(doc.Content, *node)
				currentParagraph = nil
			} else if node.Type == "heading" || node.Type == "bulletList" || node.Type == "orderedList" {
				if currentParagraph != nil && len(currentParagraph.Content) > 0 {
					doc.Content = append(doc.Content, *currentParagraph)
					currentParagraph = nil
				}
				doc.Content = append(doc.Content, *node)
			} else {
				// Text content - add to current paragraph
				if currentParagraph == nil {
					currentParagraph = &ADFNode{
						Type:    "paragraph",
						Content: []ADFNode{},
					}
				}
				currentParagraph.Content = append(currentParagraph.Content, *node)
			}
		}
	}

	// Add any remaining paragraph
	if currentParagraph != nil && len(currentParagraph.Content) > 0 {
		doc.Content = append(doc.Content, *currentParagraph)
	}

	return doc, nil
}

// convertLineToADF converts a single line to ADF nodes.
func convertLineToADF(line string) *ADFNode {
	// Check for headers
	for i := 6; i >= 1; i-- {
		prefix := strings.Repeat("#", i) + " "
		if strings.HasPrefix(line, prefix) {
			text := strings.TrimSpace(line[len(prefix):])
			return &ADFNode{
				Type: "heading",
				Attrs: map[string]interface{}{
					"level": i,
				},
				Content: []ADFNode{
					{
						Type: "text",
						Text: text,
					},
				},
			}
		}
	}

	// Check for lists
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		text := strings.TrimSpace(line[2:])
		return &ADFNode{
			Type: "bulletList",
			Content: []ADFNode{
				{
					Type: "listItem",
					Content: []ADFNode{
						{
							Type:    "paragraph",
							Content: parseInlineFormatting(text),
						},
					},
				},
			},
		}
	}

	// Check for ordered lists
	if orderedListRegex.MatchString(line) {
		text := orderedListRegex.ReplaceAllString(line, "")
		return &ADFNode{
			Type: "orderedList",
			Content: []ADFNode{
				{
					Type: "listItem",
					Content: []ADFNode{
						{
							Type:    "paragraph",
							Content: parseInlineFormatting(text),
						},
					},
				},
			},
		}
	}

	// Regular paragraph
	if line != "" {
		return &ADFNode{
			Type:    "paragraph",
			Content: parseInlineFormatting(line),
		}
	}

	return nil
}

// parseInlineFormatting parses inline formatting (bold, italic, code, links) and returns ADF nodes.
func parseInlineFormatting(text string) []ADFNode {
	nodes := []ADFNode{}

	// For simplicity, we'll handle basic text for now
	// This could be expanded to handle complex inline formatting
	if text == "" {
		return nodes
	}

	// Handle inline code
	if inlineCodeRegex.MatchString(text) {
		parts := inlineCodeRegex.Split(text, -1)
		matches := inlineCodeRegex.FindAllStringSubmatch(text, -1)

		for i, part := range parts {
			if part != "" {
				nodes = append(nodes, ADFNode{
					Type: "text",
					Text: part,
				})
			}
			if i < len(matches) {
				nodes = append(nodes, ADFNode{
					Type: "text",
					Text: matches[i][1],
					Marks: []ADFMark{
						{Type: "code"},
					},
				})
			}
		}
		return nodes
	}

	// Handle bold text
	if boldRegex1.MatchString(text) || boldRegex2.MatchString(text) {
		// Simple bold handling - could be expanded
		text = boldRegex1.ReplaceAllString(text, "${1}")
		text = boldRegex2.ReplaceAllString(text, "${1}")
		return []ADFNode{
			{
				Type: "text",
				Text: text,
				Marks: []ADFMark{
					{Type: "strong"},
				},
			},
		}
	}

	// Handle italic text
	if italicRegex1.MatchString(text) || italicRegex2.MatchString(text) {
		text = italicRegex1.ReplaceAllString(text, "${1}")
		text = italicRegex2.ReplaceAllString(text, "${1}")
		return []ADFNode{
			{
				Type: "text",
				Text: text,
				Marks: []ADFMark{
					{Type: "em"},
				},
			},
		}
	}

	// Handle links
	if linkRegex.MatchString(text) {
		matches := linkRegex.FindAllStringSubmatch(text, -1)
		if len(matches) > 0 {
			linkText := matches[0][1]
			linkURL := matches[0][2]
			return []ADFNode{
				{
					Type: "text",
					Text: linkText,
					Marks: []ADFMark{
						{
							Type: "link",
							Attrs: map[string]interface{}{
								"href": linkURL,
							},
						},
					},
				},
			}
		}
	}

	// Plain text
	return []ADFNode{
		{
			Type: "text",
			Text: text,
		},
	}
}

// ConvertMarkdownFileToADF reads a markdown file and converts it to ADF format.
func ConvertMarkdownFileToADF(filePath string) (*ADFDocument, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	return ConvertMarkdownStringToADF(string(contentBytes))
}

// ConvertMarkdownStringToADFJSON converts a markdown string to ADF JSON format.
func ConvertMarkdownStringToADFJSON(markdown string) (string, error) {
	adf, err := ConvertMarkdownStringToADF(markdown)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(adf, "", "  ")
	if err != nil {
		return "", fmt.Errorf("could not marshal ADF to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// ConvertMarkdownFileToADFJSON reads a markdown file and converts it to ADF JSON format.
func ConvertMarkdownFileToADFJSON(filePath string) (string, error) {
	adf, err := ConvertMarkdownFileToADF(filePath)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(adf, "", "  ")
	if err != nil {
		return "", fmt.Errorf("could not marshal ADF to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// MarkdownToADFWriter converts a markdown file to ADF JSON and writes to the specified writer.
func MarkdownToADFWriter(filePath string, writer io.Writer) error {
	jsonResult, err := ConvertMarkdownFileToADFJSON(filePath)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, jsonResult)
	if err != nil {
		return fmt.Errorf("could not write output: %w", err)
	}

	return nil
}
