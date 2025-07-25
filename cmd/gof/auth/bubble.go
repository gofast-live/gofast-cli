package auth

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gofast-live/gofast-cli/cmd/gof/config"
)

type (
	errMsg struct {
		err error
		msg string
	}
	authMsg struct {
		email  string
		apiKey string
	}
)

type model struct {
	focusIndex    int
	emailInput    textinput.Model
	apiKeyInput   textinput.Model
	spinner       spinner.Model
	err           errMsg
	authenticated bool
}

func initialModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))

	ei := textinput.New()
	ei.Placeholder = "Enter your email address"
	ei.Focus()
	ei.CharLimit = 156
	ei.PromptStyle = config.FocusedStyle
	ei.TextStyle = config.FocusedStyle

	ai := textinput.New()
	ai.Placeholder = "Enter your API key"
	ai.CharLimit = 156

	return model{
		focusIndex:  0,
		spinner:     sp,
		emailInput:  ei,
		apiKeyInput: ai,
		err: errMsg{
			err: nil,
			msg: "",
		},
		authenticated: false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.err = errMsg{}
			email := m.emailInput.Value()
			apiKey := m.apiKeyInput.Value()
			return m, checkConfig(email, apiKey)

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			cmd := m.toggleFocus()
			return m, cmd

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case authMsg:
		m.authenticated = true
		m.err = errMsg{}
		return m, tea.Quit
	case errMsg:
		m.err = msg
		return m, nil
	}

	cmd = m.updateInputs(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	b.WriteRune('\n')
	b.WriteString("Enter your email address and API key\n\n")
	b.WriteString(m.emailInput.View())
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(m.apiKeyInput.View())
	b.WriteRune('\n')
	b.WriteRune('\n')

	if m.err.msg != "" {
		b.WriteString(config.ErrStyle.Render(m.err.msg))
		b.WriteRune('\n')
		if m.err.err != nil {
			b.WriteString(config.ErrStyle.Render(m.err.err.Error()))
			b.WriteRune('\n')
		}
		b.WriteRune('\n')

	}

	b.WriteString(config.HelpStyle.Render("enter: submit • tab: switch input • ctrl+c: quit"))
	return b.String()
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 2)
	m.emailInput, cmds[0] = m.emailInput.Update(msg)
	m.apiKeyInput, cmds[1] = m.apiKeyInput.Update(msg)
	return tea.Batch(cmds...)
}

func (m *model) toggleFocus() tea.Cmd {
	inputs := []*textinput.Model{&m.emailInput, &m.apiKeyInput}
	m.focusIndex++
	if m.focusIndex >= len(inputs) {
		m.focusIndex = 0
	}
	for i := range inputs {
		if i == m.focusIndex {
			inputs[i].Focus()
			inputs[i].PromptStyle = config.FocusedStyle
			inputs[i].TextStyle = config.FocusedStyle
			continue
		}
		inputs[i].Blur()
		inputs[i].PromptStyle = config.NoStyle
		inputs[i].TextStyle = config.NoStyle
	}
	return textinput.Blink
}
