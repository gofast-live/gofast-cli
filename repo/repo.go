package repo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	GITHUB_URL = "https://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@github.com/gofast-live/gofast-app.git"

	noStyle      = lipgloss.NewStyle()
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
    errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	helpStyle    = blurredStyle
)

type (
	errMsg            error
	configValid       string
	tokenMsg          string
	githubFinishedMsg struct{ err error }
)

type model struct {
	step             int
	focusIndex       int
	spinner          spinner.Model
	emailInput       textinput.Model
	apiKeyInput      textinput.Model
	projectNameInput textinput.Model
	err              error
}

func InitialModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))

	ei := textinput.New()
	ei.SetValue("mateuszpiorowski@gmail.com")
	ei.Placeholder = "Enter your email address"
	ei.Focus()
	ei.CharLimit = 156
	ei.Width = 40
	ei.PromptStyle = focusedStyle
	ei.TextStyle = focusedStyle

	ai := textinput.New()
	ai.SetValue("admin")
	ai.Placeholder = "Enter your API key"
	ai.CharLimit = 156
	ai.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Enter your project name"
	pi.CharLimit = 156
	pi.Width = 40

	return model{
		focusIndex:       0,
		step:             1,
		spinner:          sp,
		emailInput:       ei,
		apiKeyInput:      ai,
		projectNameInput: pi,
		err:              nil,
	}
}

func (m model) Init() tea.Cmd {
	path, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	_, err = os.OpenFile(path+"/gofast.json", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEnter:
			if m.step == 1 {
				email := m.emailInput.Value()
				apiKey := m.apiKeyInput.Value()
				return m, checkConfig(email, apiKey)
			}
			return m, cmd

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			if m.step == 1 {
				cmd := m.toggleFocus([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
				return m, cmd
			}
		case tea.KeyCtrlQ:
            m.err = nil
			blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput, &m.projectNameInput})
			m.emailInput.Focus()
			m.emailInput.PromptStyle = focusedStyle
			m.emailInput.TextStyle = focusedStyle
			m.focusIndex = 0
			m.step = 1
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil

	case configValid:
		m.step = 2
		blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
        cmd = m.getToken()
		return m, cmd

	case githubFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		} else {
			m.step = 5
		}
	}

	cmd = m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 4)
	m.emailInput, cmds[0] = m.emailInput.Update(msg)
	m.apiKeyInput, cmds[1] = m.apiKeyInput.Update(msg)
	m.projectNameInput, cmds[2] = m.projectNameInput.Update(msg)
	m.spinner, cmds[3] = m.spinner.Update(msg)
	return tea.Batch(cmds...)
}

func blurAll(inputs []*textinput.Model) {
	for _, i := range inputs {
		i.Blur()
		i.PromptStyle = noStyle
		i.TextStyle = noStyle
	}
}

func (m *model) toggleFocus(inputs []*textinput.Model) tea.Cmd {
	cmds := make([]tea.Cmd, len(inputs))
	m.focusIndex++
	if m.focusIndex >= len(inputs) {
		m.focusIndex = 0
	}
	for i := range inputs {
		if i == m.focusIndex {
			cmds[i] = inputs[i].Focus()
			inputs[i].PromptStyle = focusedStyle
			inputs[i].TextStyle = focusedStyle
			continue
		}
		inputs[i].Blur()
		inputs[i].PromptStyle = noStyle
		inputs[i].TextStyle = noStyle
	}
	return tea.Batch(cmds...)
}

func (m *model) getToken() tea.Cmd {
	return func() tea.Msg {
		token, err := m.valdiate()
		if err != nil {
			return errMsg(err)
		}
		return tokenMsg(token)
	}
}

func copyRepo(projectName string) tea.Cmd {
	// run git clone command
	if projectName == "" {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("project name is required"))
		}
	}
	c := exec.Command("git", "clone", GITHUB_URL, projectName)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return githubFinishedMsg{err: err}
	})
}
