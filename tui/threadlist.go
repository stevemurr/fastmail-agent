package tui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"fastmail-to-text/jmap"
)

// ThreadItem represents a thread/conversation in the list
type ThreadItem struct {
	NormalizedSubject string
	Subject           string
	From              string
	Date              time.Time
	Preview           string
	EmailCount        int
	Emails            []jmap.Email // All emails in this conversation
}

// NormalizeSubject strips Re:/Fwd:/Fw: prefixes and normalizes whitespace
func NormalizeSubject(subject string) string {
	// Remove Re:/Fwd:/Fw: prefixes (case insensitive, can be repeated)
	re := regexp.MustCompile(`(?i)^(re|fwd|fw):\s*`)
	normalized := subject
	for {
		stripped := re.ReplaceAllString(normalized, "")
		if stripped == normalized {
			break
		}
		normalized = stripped
	}
	// Normalize whitespace
	normalized = strings.TrimSpace(normalized)
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	return strings.ToLower(normalized)
}

// GroupEmailsBySubject groups emails by normalized subject
func GroupEmailsBySubject(emails []jmap.Email) []ThreadItem {
	// Group by normalized subject
	groups := make(map[string][]jmap.Email)
	var order []string

	for _, email := range emails {
		key := NormalizeSubject(email.Subject)
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], email)
	}

	// Build thread items
	items := make([]ThreadItem, 0, len(order))
	for _, key := range order {
		groupEmails := groups[key]

		// Sort emails by date (oldest first for reading, but show newest date in list)
		sort.Slice(groupEmails, func(i, j int) bool {
			ti, _ := time.Parse(time.RFC3339, groupEmails[i].ReceivedAt)
			tj, _ := time.Parse(time.RFC3339, groupEmails[j].ReceivedAt)
			return ti.Before(tj)
		})

		// Use the most recent email for display info
		newest := groupEmails[len(groupEmails)-1]

		from := ""
		if len(newest.From) > 0 {
			if newest.From[0].Name != "" {
				from = newest.From[0].Name
			} else {
				from = newest.From[0].Email
			}
		}

		date, _ := time.Parse(time.RFC3339, newest.ReceivedAt)

		items = append(items, ThreadItem{
			NormalizedSubject: key,
			Subject:           newest.Subject,
			From:              from,
			Date:              date,
			Preview:           newest.Preview,
			EmailCount:        len(groupEmails),
			Emails:            groupEmails,
		})
	}

	return items
}

type threadListModel struct {
	items  []ThreadItem
	cursor int
	height int
	offset int
}

func newThreadListModel() threadListModel {
	return threadListModel{
		items:  []ThreadItem{},
		cursor: 0,
		height: 10,
		offset: 0,
	}
}

func (m *threadListModel) SetItems(items []ThreadItem) {
	m.items = items
	m.cursor = 0
	m.offset = 0
}

func (m *threadListModel) SetHeight(h int) {
	// Reserve space for borders and padding
	m.height = h - 4
	if m.height < 3 {
		m.height = 3
	}
}

func (m *threadListModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
	}
}

func (m *threadListModel) MoveDown() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
		if m.cursor >= m.offset+m.height {
			m.offset = m.cursor - m.height + 1
		}
	}
}

func (m *threadListModel) Selected() *ThreadItem {
	if len(m.items) == 0 {
		return nil
	}
	return &m.items[m.cursor]
}

func (m *threadListModel) View(width int) string {
	if len(m.items) == 0 {
		return helpStyle.Render("No emails found. Press / to search.")
	}

	var sb strings.Builder

	end := m.offset + m.height
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := m.offset; i < end; i++ {
		item := m.items[i]
		selected := i == m.cursor

		// Format the date
		dateStr := item.Date.Format("Jan 02")
		if item.Date.Year() != time.Now().Year() {
			dateStr = item.Date.Format("Jan 02 06")
		}

		// Build the line
		subject := item.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		// Add email count if more than 1
		countStr := ""
		if item.EmailCount > 1 {
			countStr = fmt.Sprintf(" (%d)", item.EmailCount)
		}

		// Truncate subject if needed
		maxSubjectLen := width - 35 - len(countStr)
		if maxSubjectLen < 20 {
			maxSubjectLen = 20
		}
		if len(subject) > maxSubjectLen {
			subject = subject[:maxSubjectLen-3] + "..."
		}

		// Truncate from if needed
		from := item.From
		if len(from) > 20 {
			from = from[:17] + "..."
		}

		line := fmt.Sprintf("%-20s  %s  %s%s",
			from,
			dateStyle.Render(dateStr),
			subject,
			countStr,
		)

		if selected {
			sb.WriteString(selectedItemStyle.Render("> " + line))
		} else {
			sb.WriteString(itemStyle.Render("  " + line))
		}
		sb.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(m.items) > m.height {
		scrollInfo := fmt.Sprintf("\n%d-%d of %d", m.offset+1, end, len(m.items))
		sb.WriteString(helpStyle.Render(scrollInfo))
	}

	return sb.String()
}
