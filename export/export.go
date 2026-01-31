package export

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"fastmail-to-text/jmap"
)

// ExportOptions controls export behavior
type ExportOptions struct {
	StripQuotes     bool
	StripSignatures bool
}

// DefaultLLMOptions returns options optimized for LLM consumption
func DefaultLLMOptions() ExportOptions {
	return ExportOptions{
		StripQuotes:     true,
		StripSignatures: true,
	}
}

// FormatThread formats a thread as readable text
func FormatThread(emails []jmap.Email) string {
	var sb strings.Builder

	for i, email := range emails {
		sb.WriteString(fmt.Sprintf("=== Email %d of %d ===\n", i+1, len(emails)))
		sb.WriteString(fmt.Sprintf("From: %s\n", formatAddresses(email.From)))
		sb.WriteString(fmt.Sprintf("To: %s\n", formatAddresses(email.To)))
		if len(email.CC) > 0 {
			sb.WriteString(fmt.Sprintf("CC: %s\n", formatAddresses(email.CC)))
		}
		sb.WriteString(fmt.Sprintf("Date: %s\n", formatDate(email.ReceivedAt)))
		sb.WriteString(fmt.Sprintf("Subject: %s\n", email.Subject))
		sb.WriteString("\n")

		body := email.GetBodyText()
		// Convert HTML to text if needed
		if strings.Contains(body, "<") && strings.Contains(body, ">") {
			body = HTMLToText(body)
		}
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// FormatThreadForLLM formats a thread in LLM-optimized format with quote/signature stripping
func FormatThreadForLLM(emails []jmap.Email, opts ExportOptions) string {
	var sb strings.Builder

	for i, email := range emails {
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("MessageIdx: %d\n", i+1))
		sb.WriteString(fmt.Sprintf("From: %s\n", formatEmailsOnly(email.From)))
		sb.WriteString(fmt.Sprintf("To: %s\n", formatEmailsOnly(email.To)))
		if len(email.CC) > 0 {
			sb.WriteString(fmt.Sprintf("CC: %s\n", formatEmailsOnly(email.CC)))
		}
		sb.WriteString(fmt.Sprintf("Date: %s\n", formatDate(email.ReceivedAt)))
		sb.WriteString("\n")

		body := getCleanBody(email, opts)
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// getCleanBody extracts and cleans email body based on options
func getCleanBody(email jmap.Email, opts ExportOptions) string {
	body := email.GetBodyText()
	return CleanBody(body, opts)
}

// CleanBody cleans a body string based on options
func CleanBody(body string, opts ExportOptions) string {
	// Handle HTML content
	if strings.Contains(body, "<") && strings.Contains(body, ">") {
		if opts.StripQuotes {
			body = HTMLToTextStripQuotes(body)
		} else {
			body = HTMLToText(body)
		}
	}

	// Strip quotes from plain text
	if opts.StripQuotes {
		body = StripQuotes(body)
	}

	// Strip signature
	if opts.StripSignatures {
		body = StripSignature(body)
	}

	return body
}

// formatEmailsOnly formats addresses as just email addresses (no display names)
func formatEmailsOnly(addrs []jmap.EmailAddress) string {
	emails := make([]string, len(addrs))
	for i, addr := range addrs {
		emails[i] = addr.Email
	}
	return strings.Join(emails, ", ")
}

// CopyToClipboard copies the formatted thread to clipboard (LLM format)
func CopyToClipboard(emails []jmap.Email) error {
	opts := DefaultLLMOptions()
	text := FormatThreadForLLM(emails, opts)
	return clipboard.WriteAll(text)
}

// CopyAttachmentInfo copies attachment metadata to clipboard
func CopyAttachmentInfo(emails []jmap.Email) error {
	text := FormatAttachmentInfo(emails)
	return clipboard.WriteAll(text)
}

// FormatAttachmentInfo formats attachment metadata as text
func FormatAttachmentInfo(emails []jmap.Email) string {
	var sb strings.Builder
	sb.WriteString("---\nAttachments:\n\n")

	idx := 1
	for _, email := range emails {
		for _, att := range email.Attachments {
			if att.IsInline {
				continue // Skip inline images
			}
			sb.WriteString(fmt.Sprintf("[%d] %s\n", idx, att.Name))
			sb.WriteString(fmt.Sprintf("    Type: %s\n", att.Type))
			sb.WriteString(fmt.Sprintf("    Size: %s\n\n", formatSize(att.Size)))
			idx++
		}
	}

	if idx == 1 {
		sb.WriteString("(No attachments)\n")
	}

	return sb.String()
}

// CopyFullThread copies thread content + attachment metadata to clipboard
func CopyFullThread(emails []jmap.Email, opts ExportOptions) error {
	content := FormatThreadForLLM(emails, opts)
	attachInfo := FormatAttachmentInfo(emails)
	return clipboard.WriteAll(content + "\n" + attachInfo)
}

// ExportToFolder creates a folder with thread.txt and downloaded attachments
func ExportToFolder(emails []jmap.Email, client *jmap.Client, opts ExportOptions) (string, error) {
	if len(emails) == 0 {
		return "", fmt.Errorf("no emails to export")
	}

	// Create export directory
	subject := sanitizeFilename(emails[0].Subject)
	timestamp := time.Now().Format("2006-01-02_150405")
	dirName := fmt.Sprintf("%s_%s", subject, timestamp)

	if err := os.MkdirAll(dirName, 0755); err != nil {
		return "", err
	}

	// Write thread.txt
	threadContent := FormatThreadForLLM(emails, opts)
	threadPath := filepath.Join(dirName, "thread.txt")
	if err := os.WriteFile(threadPath, []byte(threadContent), 0644); err != nil {
		return "", err
	}

	// Download and save attachments
	attachDir := filepath.Join(dirName, "attachments")
	hasAttachments := false
	usedNames := make(map[string]int)

	for _, email := range emails {
		for _, att := range email.Attachments {
			if att.IsInline {
				continue
			}

			if !hasAttachments {
				if err := os.MkdirAll(attachDir, 0755); err != nil {
					return "", err
				}
				hasAttachments = true
			}

			data, err := client.DownloadBlob(att.BlobID, att.Name, att.Type)
			if err != nil {
				return "", fmt.Errorf("failed to download %s: %w", att.Name, err)
			}

			// Handle filename conflicts
			filename := deduplicateFilename(att.Name, usedNames)
			attPath := filepath.Join(attachDir, filename)
			if err := os.WriteFile(attPath, data, 0644); err != nil {
				return "", err
			}
		}
	}

	return dirName, nil
}

// deduplicateFilename ensures unique filenames by appending _1, _2, etc.
func deduplicateFilename(name string, usedNames map[string]int) string {
	count, exists := usedNames[name]
	if !exists {
		usedNames[name] = 1
		return name
	}

	usedNames[name] = count + 1
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s_%d%s", base, count, ext)
}

// formatSize formats bytes as human-readable size
func formatSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// ExportToFile exports the formatted thread to a file
func ExportToFile(emails []jmap.Email, filename string) error {
	if filename == "" {
		// Auto-generate filename from subject and date
		subject := "email"
		if len(emails) > 0 {
			subject = sanitizeFilename(emails[0].Subject)
		}
		filename = fmt.Sprintf("%s_%s.txt", subject, time.Now().Format("2006-01-02_150405"))
	}

	text := FormatThread(emails)
	return os.WriteFile(filename, []byte(text), 0644)
}

// formatAddresses formats a list of email addresses
func formatAddresses(addrs []jmap.EmailAddress) string {
	parts := make([]string, len(addrs))
	for i, addr := range addrs {
		parts[i] = addr.String()
	}
	return strings.Join(parts, ", ")
}

// formatDate formats a JMAP date string
func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Local().Format("2006-01-02 3:04 PM")
}

// sanitizeFilename removes invalid characters from a filename
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = re.ReplaceAllString(name, "_")

	// Trim spaces and limit length
	name = strings.TrimSpace(name)
	if len(name) > 50 {
		name = name[:50]
	}
	if name == "" {
		name = "email"
	}
	return name
}
