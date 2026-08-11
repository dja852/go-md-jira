package gomdjira

import (
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
	bareLinkRegex      = regexp.MustCompile(`<(.+?)>`)
	orderedListRegex   = regexp.MustCompile(`^\d+\.\s+`)
	unorderedListRegex = regexp.MustCompile(`^\*\s+`)

	// Nested list patterns
	nestedOrderedListRegex   = regexp.MustCompile(`^( {2,})\d+\.\s+`)
	nestedUnorderedListRegex = regexp.MustCompile(`^( {2,})[*-]\s+`)

	// Fenced code block pattern
	fencedCodeRegex = regexp.MustCompile("(?s)```(\\w+)?\\n(.*?)\\n```")

	// Jira -> Markdown patterns
	jiraCodeBlockRegex     = regexp.MustCompile(`(?s)\{code(?::([^}\n]+))?\}\n(.*?)\n\{code\}`)
	jiraHeaderRegex        = regexp.MustCompile(`^h([1-6])\.\s+(.+)$`)
	jiraOrderedListRegex   = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	jiraUnorderedListRegex = regexp.MustCompile(`^(-{1,6})\s+(.+)$`)
	jiraInlineCodeRegex    = regexp.MustCompile(`\{\{([^{}]+)\}\}`)
	jiraLinkRegex          = regexp.MustCompile(`\[(.*?)\|(.+?)\]`)
	jiraBoldRegex          = regexp.MustCompile(`(^|[^[:alnum:]])\*([^*\n]+)\*([^[:alnum:]]|$)`)
	jiraItalicRegex        = regexp.MustCompile(`(^|[^[:alnum:]])_([^_\n]+)_([^[:alnum:]]|$)`)
	jiraStrikethroughRegex = regexp.MustCompile(`(^|[^[:alnum:]])-([[:alnum:]][^-\n]*[[:alnum:]])-([^[:alnum:]]|$)`)
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
	line = bareLinkRegex.ReplaceAllString(line, "[${1}|${1}]")

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

type convertedMarkdownLine struct {
	text       string
	isListItem bool
}

func isJiraListItem(line string) bool {
	return jiraOrderedListRegex.MatchString(line) || jiraUnorderedListRegex.MatchString(line)
}

func removeSingleBlankLinesBetweenListItems(lines []convertedMarkdownLine) []string {
	normalizedLines := make([]string, 0, len(lines))
	for i, line := range lines {
		isSingleListBlank := strings.TrimSpace(line.text) == "" &&
			i > 0 && i+1 < len(lines) &&
			lines[i-1].isListItem && lines[i+1].isListItem
		if isSingleListBlank {
			continue
		}

		normalizedLines = append(normalizedLines, line.text)
	}

	return normalizedLines
}

// ConvertMarkdownString converts a markdown string to Jira markup.
func ConvertMarkdownString(markdown string) string {
	convertedContent := convertMultilineElements(markdown)
	lines := strings.Split(convertedContent, "\n")
	jiraLines := make([]convertedMarkdownLine, 0, len(lines))

	inCodeBlock := false
	for _, line := range lines {
		// Track code block boundaries to avoid processing content inside them
		if strings.HasPrefix(line, "{code") && (strings.HasSuffix(line, "}") || line == "{code}") {
			inCodeBlock = !inCodeBlock
			jiraLines = append(jiraLines, convertedMarkdownLine{text: line})
			continue
		}

		// Only apply line conversions if we're not in a code block
		listItem := false
		if !inCodeBlock {
			line = convertLine(line)
			listItem = isJiraListItem(line)
		}
		jiraLines = append(jiraLines, convertedMarkdownLine{text: line, isListItem: listItem})
	}

	return strings.Join(removeSingleBlankLinesBetweenListItems(jiraLines), "\n")
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

// convertJiraMultilineElements converts Jira code blocks to fenced Markdown code blocks.
func convertJiraMultilineElements(content string) string {
	return jiraCodeBlockRegex.ReplaceAllStringFunc(content, processJiraCodeBlock)
}

// processJiraCodeBlock converts a Jira code block to fenced Markdown format.
func processJiraCodeBlock(match string) string {
	matches := jiraCodeBlockRegex.FindStringSubmatch(match)
	if len(matches) < 3 {
		return match // Safety fallback
	}

	lang := matches[1]
	code := matches[2]

	if lang != "" {
		return fmt.Sprintf("```%s\n%s\n```", lang, code)
	}
	return fmt.Sprintf("```\n%s\n```", code)
}

// convertJiraInline converts Jira inline markup to Markdown inline markup.
func convertJiraInline(line string) string {
	const inlineCodePlaceholder = "__INLINE_CODE__"
	var inlineCodeBlocks []string

	line = jiraInlineCodeRegex.ReplaceAllStringFunc(line, func(match string) string {
		codeContent := jiraInlineCodeRegex.FindStringSubmatch(match)[1]
		inlineCodeBlocks = append(inlineCodeBlocks, "`"+codeContent+"`")
		return inlineCodePlaceholder + fmt.Sprintf("%d", len(inlineCodeBlocks)-1) + inlineCodePlaceholder
	})

	line = jiraLinkRegex.ReplaceAllString(line, "[${1}](${2})")
	line = jiraBoldRegex.ReplaceAllString(line, "${1}**${2}**${3}")
	line = jiraItalicRegex.ReplaceAllString(line, "${1}_${2}_${3}")
	line = jiraStrikethroughRegex.ReplaceAllString(line, "${1}~~${2}~~${3}")

	for i, codeBlock := range inlineCodeBlocks {
		placeholder := inlineCodePlaceholder + fmt.Sprintf("%d", i) + inlineCodePlaceholder
		line = strings.ReplaceAll(line, placeholder, codeBlock)
	}

	return line
}

// convertJiraLine converts a single Jira wiki line to Markdown.
func convertJiraLine(line string) string {
	if matches := jiraOrderedListRegex.FindStringSubmatch(line); len(matches) == 3 {
		depth := len(matches[1]) - 1
		if depth < 0 {
			depth = 0
		}
		indent := strings.Repeat("    ", depth)
		return indent + "1. " + convertJiraInline(matches[2])
	}

	if matches := jiraUnorderedListRegex.FindStringSubmatch(line); len(matches) == 3 {
		depth := len(matches[1]) - 1
		if depth < 0 {
			depth = 0
		}
		indent := strings.Repeat("    ", depth)
		return indent + "- " + convertJiraInline(matches[2])
	}

	if matches := jiraHeaderRegex.FindStringSubmatch(line); len(matches) == 3 {
		level := matches[1]
		text := matches[2]
		return strings.Repeat("#", int(level[0]-'0')) + " " + convertJiraInline(text)
	}

	return convertJiraInline(line)
}

// ConvertJiraString converts Jira wiki markup string to Markdown.
func ConvertJiraString(jira string) string {
	convertedContent := convertJiraMultilineElements(jira)
	lines := strings.Split(convertedContent, "\n")
	markdownLines := make([]string, 0, len(lines))

	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			markdownLines = append(markdownLines, line)
			continue
		}

		if !inCodeBlock {
			line = convertJiraLine(line)
		}

		markdownLines = append(markdownLines, line)
	}

	return strings.Join(markdownLines, "\n")
}

// ConvertJiraFile reads a Jira wiki file and returns converted Markdown as a string.
func ConvertJiraFile(filePath string) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}

	return ConvertJiraString(string(contentBytes)), nil
}

// JiraToMarkdown reads a Jira wiki file, converts content to Markdown, and prints it.
func JiraToMarkdown(filePath string) error {
	return JiraToMarkdownWriter(filePath, os.Stdout)
}

// JiraToMarkdownWriter converts a Jira wiki file to Markdown and writes to the specified writer.
func JiraToMarkdownWriter(filePath string, writer io.Writer) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	result := ConvertJiraString(string(contentBytes))
	_, err = fmt.Fprintln(writer, result)
	if err != nil {
		return fmt.Errorf("could not write output: %w", err)
	}

	return nil
}
