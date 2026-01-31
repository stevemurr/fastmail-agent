package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type searchModel struct {
	input textinput.Model
}

func newSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search emails..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return searchModel{
		input: ti,
	}
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m searchModel) View() string {
	return searchInputStyle.Render(m.input.View())
}

func (m searchModel) Value() string {
	return m.input.Value()
}

func (m *searchModel) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *searchModel) Blur() {
	m.input.Blur()
}

func (m *searchModel) SetValue(s string) {
	m.input.SetValue(s)
}
