package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"fastmail-to-text/export"
	"fastmail-to-text/jmap"
)

type threadViewModel struct {
	emails   []jmap.Email
	viewport viewport.Model
	ready    bool
}

func newThreadViewModel() threadViewModel {
	return threadViewModel{}
}

func (m *threadViewModel) SetEmails(emails []jmap.Email) {
	m.emails = emails
	m.viewport.SetContent(m.formatContent())
	m.viewport.GotoTop()
}

func (m *threadViewModel) SetSize(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(width, height-4)
		m.viewport.HighPerformanceRendering = false
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = height - 4
	}

	if len(m.emails) > 0 {
		m.viewport.SetContent(m.formatContent())
	}
}

func (m *threadViewModel) formatContent() string {
	var sb strings.Builder

	for i, email := range m.emails {
		// Email header
		sb.WriteString(fmt.Sprintf("═══ Email %d of %d ═══\n", i+1, len(m.emails)))

		// From
		from := formatAddresses(email.From)
		sb.WriteString(fmt.Sprintf("From: %s\n", from))

		// To
		to := formatAddresses(email.To)
		sb.WriteString(fmt.Sprintf("To: %s\n", to))

		// CC if present
		if len(email.CC) > 0 {
			cc := formatAddresses(email.CC)
			sb.WriteString(fmt.Sprintf("CC: %s\n", cc))
		}

		// Date
		date := formatDate(email.ReceivedAt)
		sb.WriteString(fmt.Sprintf("Date: %s\n", date))

		// Subject
		sb.WriteString(fmt.Sprintf("Subject: %s\n", email.Subject))
		sb.WriteString("\n")

		// Body
		body := email.GetBodyText()
		if strings.Contains(body, "<") && strings.Contains(body, ">") {
			body = export.HTMLToText(body)
		}
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (m threadViewModel) Update(msg tea.Msg) (threadViewModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m threadViewModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.viewport.View()
}

func (m *threadViewModel) ScrollDown() {
	m.viewport.LineDown(1)
}

func (m *threadViewModel) ScrollUp() {
	m.viewport.LineUp(1)
}

func (m *threadViewModel) PageDown() {
	m.viewport.HalfViewDown()
}

func (m *threadViewModel) PageUp() {
	m.viewport.HalfViewUp()
}

func (m *threadViewModel) Emails() []jmap.Email {
	return m.emails
}

func formatAddresses(addrs []jmap.EmailAddress) string {
	parts := make([]string, len(addrs))
	for i, addr := range addrs {
		parts[i] = addr.String()
	}
	return strings.Join(parts, ", ")
}

func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Local().Format("2006-01-02 3:04 PM")
}
