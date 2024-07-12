package repo

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	GITHUB_URL   = "@github.com/gofast-live/gofast-app.git"
	noStyle      = lipgloss.NewStyle()
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	helpStyle    = blurredStyle
)

type (
	errMsg      error
	configValid string
	tokenMsg    string
	copyMsg     struct{ err error }
	finishMsg   struct{}
)

type model struct {
	step             int
	focusIndex       int
	token            string
	spinner          spinner.Model
	emailInput       textinput.Model
	apiKeyInput      textinput.Model
	projectNameInput textinput.Model
	err              error

	protocols                []string
	selectedProtocol         string
	databases                []string
	selectedDatabase         string
	paymentsProviders        []string
	selectedPaymentsProvider string
	emailsProviders          []string
	selectedEmailProvider    string
	filesProviders           []string
	selectedFilesProvider    string
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
	ei.PromptStyle = focusedStyle
	ei.TextStyle = focusedStyle

	ai := textinput.New()
	ai.SetValue("sk_live_12379277dd33f815dde9f5acc8b62ec4c4acb8d7ee05abb3a50140fc9e6d7e7b46a4b75a")
	ai.Placeholder = "Enter your API key"
	ai.CharLimit = 156

	pi := textinput.New()
	pi.Placeholder = "Enter your project name"
	pi.CharLimit = 156
	pi.Width = 40

	return model{
		focusIndex:       0,
		step:             1,
		token:            "",
		spinner:          sp,
		emailInput:       ei,
		apiKeyInput:      ai,
		projectNameInput: pi,
		err:              nil,

		protocols:                []string{"HTTP", "gRPC"},
		selectedProtocol:         "HTTP",
		databases:                []string{"Turso", "PostgreSQL", "MySQL", "SQLite"},
		selectedDatabase:         "Turso",
		paymentsProviders:        []string{"Stripe", "Lemon Squeezy (not implemented)", "None"},
		selectedPaymentsProvider: "Stripe",
		emailsProviders:          []string{"Postmark", "Sendgrid", "Resend", "None"},
		selectedEmailProvider:    "Postmark",
		filesProviders:           []string{"S3", "None"},
		selectedFilesProvider:    "S3",
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
			m.err = nil
			if m.step == 1 {
				email := m.emailInput.Value()
				apiKey := m.apiKeyInput.Value()
				return m, checkConfig(email, apiKey)
			} else if m.step == 3 {
				blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
				m.projectNameInput.Focus()
				m.projectNameInput.PromptStyle = focusedStyle
				m.projectNameInput.TextStyle = focusedStyle
				m.step = 4
				return m, textinput.Blink
			} else if m.step == 4 {
				m.selectedProtocol = m.protocols[m.focusIndex]
				m.focusIndex = 0
				m.step = 5
			} else if m.step == 5 {
				m.selectedDatabase = m.databases[m.focusIndex]
				m.focusIndex = 0
				m.step = 6
			} else if m.step == 6 {
				m.selectedPaymentsProvider = m.paymentsProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = 7
			} else if m.step == 7 {
				m.selectedEmailProvider = m.emailsProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = 8
			} else if m.step == 8 {
				m.selectedFilesProvider = m.filesProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = 9
			} else if m.step == 9 {
				projectName := m.projectNameInput.Value()
				if projectName == "" {
					return m, func() tea.Msg {
						return errMsg(fmt.Errorf("Project name cannot be empty"))
					}
				}
                // check if there is a dir with the same name
                if _, err := os.Stat(projectName); !os.IsNotExist(err) {
                    return m, func() tea.Msg {
                        return errMsg(fmt.Errorf("Directory with the same name already exists"))
                    }
                }
				m.step = 10
				return m, m.copyRepo(m.token, projectName)
			} else if m.step == 12 {
				return m, tea.Quit
			}

			return m, cmd

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			if m.step == 1 {
				cmd := m.toggleFocus([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
				return m, cmd
			} else if m.step == 4 || m.step == 5 || m.step == 6 || m.step == 7 || m.step == 8 {
				var d []string
				if m.step == 4 {
					d = m.protocols
				} else if m.step == 5 {
					d = m.databases
				} else if m.step == 6 {
					d = m.paymentsProviders
				} else if m.step == 7 {
					d = m.emailsProviders
				} else if m.step == 8 {
					d = m.filesProviders
				}
				if tea.KeyDown == msg.Type || tea.KeyTab == msg.Type {
					if m.focusIndex == len(d)-1 {
						m.focusIndex = 0
					} else {
						m.focusIndex++
					}
				} else {
					if m.focusIndex == 0 {
						m.focusIndex = len(d) - 1
					} else {
						m.focusIndex--
					}
				}
				return m, nil
			}
		case tea.KeyCtrlQ:
			m.err = nil
			blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput, &m.projectNameInput})
			m.emailInput.Focus()
			m.emailInput.PromptStyle = focusedStyle
			m.emailInput.TextStyle = focusedStyle
			m.focusIndex = 0
			m.step = 1
			return m, textinput.Blink
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	case configValid:
		m.step = 2
		cmd = m.getToken()
		return m, cmd
	case tokenMsg:
		m.step = 3
		m.token = string(msg)
		return m, nil
	case copyMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		} else {
			m.step = 11
		}
		return m, m.cleaningRepo()
	case finishMsg:
		m.step = 12
		return m, nil
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
		// min 1 sec
		now := time.Now()
		token, err := validateConfig()
		if err != nil {
			return errMsg(err)
		}
		elapsed := time.Since(now)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
		return tokenMsg(token)
	}
}

func (m *model) copyRepo(token string, projectName string) tea.Cmd {
	return func() tea.Msg {
		authURL := fmt.Sprintf("https://%s%s", token, GITHUB_URL)
		c := exec.Command("git", "clone", authURL, projectName)
		c.Stdout = os.Stdout
		err := c.Start()
		if err != nil {
			return errMsg(err)
		}
		return copyMsg{err: c.Wait()}
	}
}

func (m *model) cleaningRepo() tea.Cmd {
	return func() tea.Msg {
		err := cleaning(m.projectNameInput.Value(), m.selectedProtocol, m.selectedDatabase, m.selectedPaymentsProvider, m.selectedEmailProvider, m.selectedFilesProvider)
		if err != nil {
			return errMsg(err)
		}
		return finishMsg{}
	}
}
