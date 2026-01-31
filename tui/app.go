package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"fastmail-to-text/export"
	"fastmail-to-text/jmap"
)

var debugFile *os.File

func init() {
	debugFile, _ = os.Create("debug.log")
}

func debugLog(format string, args ...interface{}) {
	if debugFile != nil {
		fmt.Fprintf(debugFile, format+"\n", args...)
		debugFile.Sync()
	}
}

type view int

const (
	viewSearch view = iota
	viewList
	viewThread
)

type Model struct {
	client     *jmap.Client
	view       view
	search     searchModel
	threadList threadListModel
	threadView threadViewModel
	width      int
	height     int
	loading    bool
	status     string
	err        error
}

// Messages
type searchResultMsg struct {
	emails []jmap.Email
	err    error
}

type threadLoadedMsg struct {
	emails []jmap.Email
	err    error
}

type statusMsg string

type errMsg error

type exportFolderMsg struct {
	dirName string
	err     error
}

func New(client *jmap.Client) Model {
	return Model{
		client:     client,
		view:       viewSearch,
		search:     newSearchModel(),
		threadList: newThreadListModel(),
		threadView: newThreadViewModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit globally
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		// Handle view-specific keys
		switch m.view {
		case viewSearch:
			return m.updateSearch(msg)
		case viewList:
			return m.updateList(msg)
		case viewThread:
			return m.updateThread(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.threadList.SetHeight(m.height - 6)
		m.threadView.SetSize(m.width, m.height-4)
		return m, nil

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = "Error: " + msg.err.Error()
			return m, nil
		}

		// Debug: log results
		for _, email := range msg.emails {
			debugLog("Search result: Subject=%q ThreadID=%s", email.Subject, email.ThreadID)
		}

		// Group emails by normalized subject
		items := GroupEmailsBySubject(msg.emails)

		debugLog("Grouped into %d conversations", len(items))
		for _, item := range items {
			debugLog("  %q: %d emails", item.Subject, item.EmailCount)
		}

		m.threadList.SetItems(items)
		m.view = viewList
		m.status = fmt.Sprintf("Found %d conversations", len(items))
		return m, nil

	case threadLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = "Error: " + msg.err.Error()
			return m, nil
		}

		// Debug: log loaded thread emails
		debugLog("Thread loaded with %d emails", len(msg.emails))
		for i, email := range msg.emails {
			debugLog("  [%d] Subject=%q From=%v", i, email.Subject, email.From)
		}

		m.threadView.SetEmails(msg.emails)
		m.view = viewThread
		m.status = fmt.Sprintf("%d emails in thread", len(msg.emails))
		return m, nil

	case exportFolderMsg:
		m.loading = false
		if msg.err != nil {
			m.status = "Export failed: " + msg.err.Error()
		} else {
			m.status = fmt.Sprintf("Exported to %s/", msg.dirName)
		}
		return m, nil

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case errMsg:
		m.err = msg
		m.status = "Error: " + msg.Error()
		return m, nil
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		query := m.search.Value()
		m.loading = true
		m.status = "Searching..."
		return m, m.doSearch(query)

	case key.Matches(msg, keys.Back):
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		m.threadList.MoveUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.threadList.MoveDown()
		return m, nil

	case key.Matches(msg, keys.Enter):
		selected := m.threadList.Selected()
		if selected != nil && len(selected.Emails) > 0 {
			m.loading = true
			m.status = "Loading emails..."
			// Get email IDs to fetch full bodies
			ids := make([]string, len(selected.Emails))
			for i, e := range selected.Emails {
				ids[i] = e.ID
			}
			return m, m.loadEmails(ids)
		}
		return m, nil

	case key.Matches(msg, keys.Search):
		m.view = viewSearch
		return m, m.search.Focus()

	case key.Matches(msg, keys.Back):
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) updateThread(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		m.threadView.ScrollUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.threadView.ScrollDown()
		return m, nil

	case key.Matches(msg, keys.PageUp):
		m.threadView.PageUp()
		return m, nil

	case key.Matches(msg, keys.PageDown):
		m.threadView.PageDown()
		return m, nil

	case key.Matches(msg, keys.Copy):
		emails := m.threadView.Emails()
		if err := export.CopyToClipboard(emails); err != nil {
			m.status = "Copy failed: " + err.Error()
		} else {
			m.status = "Copied to clipboard (LLM format)!"
		}
		return m, nil

	case key.Matches(msg, keys.CopyAttachments):
		emails := m.threadView.Emails()
		if err := export.CopyAttachmentInfo(emails); err != nil {
			m.status = "Copy failed: " + err.Error()
		} else {
			m.status = "Attachment info copied!"
		}
		return m, nil

	case key.Matches(msg, keys.CopyFull):
		emails := m.threadView.Emails()
		opts := export.DefaultLLMOptions()
		if err := export.CopyFullThread(emails, opts); err != nil {
			m.status = "Copy failed: " + err.Error()
		} else {
			m.status = "Full thread copied (with attachments)!"
		}
		return m, nil

	case key.Matches(msg, keys.ExportFolder):
		emails := m.threadView.Emails()
		m.loading = true
		m.status = "Exporting to folder..."
		return m, m.doExportFolder(emails)

	case key.Matches(msg, keys.Export):
		emails := m.threadView.Emails()
		if err := export.ExportToFile(emails, ""); err != nil {
			m.status = "Export failed: " + err.Error()
		} else {
			m.status = "Exported to file!"
		}
		return m, nil

	case key.Matches(msg, keys.Back):
		m.view = viewList
		return m, nil
	}

	// Pass other keys to viewport
	var cmd tea.Cmd
	m.threadView, cmd = m.threadView.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.view {
	case viewSearch:
		content = m.viewSearch()
	case viewList:
		content = m.viewList()
	case viewThread:
		content = m.viewThread()
	}

	// Add status bar
	status := m.status
	if m.loading {
		status = "Loading..."
	}

	statusBar := statusBarStyle.Render(status)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m Model) viewSearch() string {
	title := titleStyle.Render("Fastmail Search")
	label := searchLabelStyle.Render("Enter search query:")
	input := m.search.View()
	help := helpStyle.Render("\nPress Enter to search, Esc to quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, label, input, help)
}

func (m Model) viewList() string {
	title := titleStyle.Render("Threads")
	list := m.threadList.View(m.width)
	help := helpStyle.Render("\n↑/↓ navigate • Enter open • / search • q quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, list, help)
}

func (m Model) viewThread() string {
	selected := m.threadList.Selected()
	title := ""
	if selected != nil {
		title = titleStyle.Render(selected.Subject)
	}

	content := m.threadView.View()
	help := helpStyle.Render("↑/↓ scroll • c copy • a attachments • f full • j folder • e export • q back")

	return lipgloss.JoinVertical(lipgloss.Left, title, content, help)
}

func (m Model) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		emails, err := m.client.SearchEmails(query, 50)
		return searchResultMsg{emails: emails, err: err}
	}
}

func (m Model) loadEmails(ids []string) tea.Cmd {
	return func() tea.Msg {
		emails, err := m.client.GetEmails(ids)
		return threadLoadedMsg{emails: emails, err: err}
	}
}

func (m Model) doExportFolder(emails []jmap.Email) tea.Cmd {
	return func() tea.Msg {
		opts := export.DefaultLLMOptions()
		dirName, err := export.ExportToFolder(emails, m.client, opts)
		return exportFolderMsg{dirName: dirName, err: err}
	}
}
