package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Enter           key.Binding
	Back            key.Binding
	Quit            key.Binding
	Search          key.Binding
	Export          key.Binding
	Copy            key.Binding
	CopyAttachments key.Binding
	CopyFull        key.Binding
	ExportFolder    key.Binding
	PageUp          key.Binding
	PageDown        key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	CopyAttachments: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "copy attachments"),
	),
	CopyFull: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "copy full"),
	),
	ExportFolder: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "export folder"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdn", "page down"),
	),
}
