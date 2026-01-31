package export

import (
	"regexp"
	"strings"
)

// Quote detection patterns
var (
	// "On Mon, Jan 1, 2024 at 10:00 AM John Doe <john@example.com> wrote:"
	attributionPattern = regexp.MustCompile(`(?im)^On .+wrote:\s*$`)

	// Lines starting with > (quoted text)
	quotedLinePattern = regexp.MustCompile(`(?m)^>+.*$`)

	// Gmail-style: "---------- Forwarded message ---------"
	forwardedPattern = regexp.MustCompile(`(?im)^-{5,}\s*Forwarded message\s*-{5,}`)

	// Outlook-style: "----- Original Message -----"
	originalMsgPattern = regexp.MustCompile(`(?im)^-{5,}\s*Original Message\s*-{5,}`)

	// "From: ... Sent: ... To: ... Subject:" block header pattern
	outlookHeaderPattern = regexp.MustCompile(`(?im)^From:\s*.+\n(Sent|Date):\s*.+\nTo:\s*.+\n(Cc:\s*.+\n)?Subject:\s*.+$`)
)

// Signature detection patterns
var (
	// Standard signature delimiter: "-- " (dash dash space, per RFC)
	sigDelimiterPattern = regexp.MustCompile(`(?m)^--\s*$`)

	// Common mobile/client signature phrases
	sigPhrasePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?im)^sent from my (iphone|ipad|android|mobile|galaxy|pixel).*$`),
		regexp.MustCompile(`(?im)^get outlook for .*$`),
		regexp.MustCompile(`(?im)^sent from mail for windows.*$`),
		regexp.MustCompile(`(?im)^sent from yahoo mail.*$`),
		regexp.MustCompile(`(?im)^sent via .*$`),
	}
)

// StripQuotes removes quoted portions from email body text
func StripQuotes(body string) string {
	// Step 1: Find attribution line ("On X wrote:") and truncate everything after
	if loc := attributionPattern.FindStringIndex(body); loc != nil {
		body = strings.TrimSpace(body[:loc[0]])
	}

	// Step 2: Find forwarded/original message markers and truncate
	for _, pattern := range []*regexp.Regexp{forwardedPattern, originalMsgPattern} {
		if loc := pattern.FindStringIndex(body); loc != nil {
			body = strings.TrimSpace(body[:loc[0]])
		}
	}

	// Step 3: Find Outlook-style header blocks and truncate
	if loc := outlookHeaderPattern.FindStringIndex(body); loc != nil {
		body = strings.TrimSpace(body[:loc[0]])
	}

	// Step 4: Remove lines starting with > (inline quotes)
	body = quotedLinePattern.ReplaceAllString(body, "")

	// Step 5: Clean up excessive blank lines
	body = regexp.MustCompile(`\n{3,}`).ReplaceAllString(body, "\n\n")

	return strings.TrimSpace(body)
}

// StripSignature removes email signature from body text
func StripSignature(body string) string {
	// Method 1: Standard "-- " delimiter
	if loc := sigDelimiterPattern.FindStringIndex(body); loc != nil {
		body = strings.TrimSpace(body[:loc[0]])
	}

	// Method 2: Check for common mobile/client signature phrases at end of email
	lines := strings.Split(body, "\n")
	for i := len(lines) - 1; i >= 0 && i >= len(lines)-5; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		for _, pattern := range sigPhrasePatterns {
			if pattern.MatchString(line) {
				body = strings.TrimSpace(strings.Join(lines[:i], "\n"))
				return body
			}
		}
	}

	return body
}

// StripQuotesFromHTML removes blockquote elements from HTML before text extraction
func StripQuotesFromHTML(html string) string {
	// Remove <blockquote>...</blockquote> and contents (non-greedy, handles nesting poorly but works for most cases)
	blockquotePattern := regexp.MustCompile(`(?is)<blockquote[^>]*>.*?</blockquote>`)
	html = blockquotePattern.ReplaceAllString(html, "")

	// Remove Gmail's quoted div: <div class="gmail_quote">...</div>
	gmailQuotePattern := regexp.MustCompile(`(?is)<div[^>]*class="[^"]*gmail_quote[^"]*"[^>]*>.*?</div>`)
	html = gmailQuotePattern.ReplaceAllString(html, "")

	// Remove Outlook's quoted div: <div id="appendonsend">...</div> or similar quote markers
	outlookQuotePattern := regexp.MustCompile(`(?is)<div[^>]*id="(appendonsend|divRplyFwdMsg)"[^>]*>.*?</div>`)
	html = outlookQuotePattern.ReplaceAllString(html, "")

	return html
}
