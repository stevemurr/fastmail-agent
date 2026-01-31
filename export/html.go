package export

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// HTMLToText converts HTML to plain text
func HTMLToText(htmlStr string) string {
	return htmlToTextWithOptions(htmlStr, false)
}

// HTMLToTextStripQuotes converts HTML to plain text, removing quoted content
func HTMLToTextStripQuotes(htmlStr string) string {
	return htmlToTextWithOptions(htmlStr, true)
}

func htmlToTextWithOptions(htmlStr string, stripQuotes bool) string {
	if stripQuotes {
		htmlStr = StripQuotesFromHTML(htmlStr)
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		// Fall back to regex-based stripping
		return stripHTMLRegex(htmlStr)
	}

	var sb strings.Builder
	extractText(doc, &sb)
	text := sb.String()

	// Clean up whitespace
	text = cleanWhitespace(text)

	return text
}

func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
	}

	// Add newlines for block elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "br":
			sb.WriteString("\n")
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr":
			sb.WriteString("\n")
		case "script", "style", "head":
			// Skip these elements entirely
			return
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}

	// Add trailing newline for block elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "table":
			sb.WriteString("\n")
		}
	}
}

func stripHTMLRegex(s string) string {
	// Remove script and style blocks
	reScript := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	s = reScript.ReplaceAllString(s, "")
	reStyle := regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	s = reStyle.ReplaceAllString(s, "")

	// Replace br and block elements with newlines
	reBr := regexp.MustCompile(`(?i)<br\s*/?>`)
	s = reBr.ReplaceAllString(s, "\n")
	reBlock := regexp.MustCompile(`(?i)</(p|div|h[1-6]|li|tr)>`)
	s = reBlock.ReplaceAllString(s, "\n")

	// Remove remaining tags
	reTags := regexp.MustCompile(`<[^>]*>`)
	s = reTags.ReplaceAllString(s, "")

	// Decode common HTML entities
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")

	return cleanWhitespace(s)
}

func cleanWhitespace(s string) string {
	// Replace multiple spaces with single space
	reSpaces := regexp.MustCompile(`[ \t]+`)
	s = reSpaces.ReplaceAllString(s, " ")

	// Replace multiple newlines with double newline
	reNewlines := regexp.MustCompile(`\n{3,}`)
	s = reNewlines.ReplaceAllString(s, "\n\n")

	// Trim each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	s = strings.Join(lines, "\n")

	return strings.TrimSpace(s)
}
