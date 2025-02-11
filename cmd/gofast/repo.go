package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	SERVER_URL   = "https://admin.gofast.live/api"
	noStyle      = lipgloss.NewStyle()
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	helpStyle    = blurredStyle

	authStep             = 1
	validateStep         = 2
	startStep            = 3
	clientStep           = 4
	startOptionStep      = 5
	databaseStep         = 6
	paymentsProviderStep = 7
	emailProviderStep    = 8
	filesProviderStep    = 9
	monitoringStep       = 10
	projectNameStep      = 11
	cleaningStep         = 12
	finishStep           = 13
	successStep          = 14
)

type (
	errMsg    error
	authMsg   struct{ email, apiKey string }
	copyMsg   struct{ err error }
	finishMsg struct{ docker []string }
)

type model struct {
	step             int
	focusIndex       int
	token            string
	spinner          spinner.Model
	email            string
	emailInput       textinput.Model
	apiKey           string
	apiKeyInput      textinput.Model
	projectNameInput textinput.Model
	err              error

	clients                  []string
	selectedClient           string
	startOptions             []string
	selectedStartOption      string
	databases                []string
	selectedDatabase         string
	paymentsProviders        []string
	selectedPaymentsProvider string
	emailsProviders          []string
	selectedEmailProvider    string
	filesProviders           []string
	selectedFilesProvider    string
	monitoringOptions        []string
	selectedMonitoring       string
	docker                   []string
}

func initialModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))

	ei := textinput.New()
	ei.Placeholder = "Enter your email address"
	ei.Focus()
	ei.CharLimit = 156
	ei.PromptStyle = focusedStyle
	ei.TextStyle = focusedStyle

	ai := textinput.New()
	ai.Placeholder = "Enter your API key"
	ai.CharLimit = 156

	pi := textinput.New()
	pi.Placeholder = "Enter your project name"
	pi.CharLimit = 156
	pi.Width = 40

	return model{
		focusIndex:       0,
		step:             0,
		token:            "",
		spinner:          sp,
		email:            "",
		emailInput:       ei,
		apiKey:           "",
		apiKeyInput:      ai,
		projectNameInput: pi,
		err:              nil,

		clients:                  []string{"SvelteKit", "Next.js", "Vue.js", "None"},
		selectedClient:           "SvelteKit",
		startOptions:             []string{"Generate base project (SQLite, Grafana Monitoring, Mocked payments, Local files, Log Emails)", "Start new configuration"},
		selectedStartOption:      "Generate base project (SQLite, Mocked payments, Log emails, Local files)",
		databases:                []string{"SQLite", "Turso with Embedded Replicas", "PostgreSQL (local)", "PostgreSQL (remote)"},
		selectedDatabase:         "SQLite",
		paymentsProviders:        []string{"Local (mock)", "Stripe", "Lemon Squeezy"},
		selectedPaymentsProvider: "Local (mock)",
		emailsProviders:          []string{"Local (log)", "Postmark", "Sendgrid", "Resend", "AWS SES"},
		selectedEmailProvider:    "Local (log)",
		filesProviders:           []string{"Local (folder)", "Cloudflare R2", "AWS S3", "Google Cloud Storage", "Azure Blob Storage"},
		selectedFilesProvider:    "Local (folder)",
		monitoringOptions:        []string{"Kubernetes + VictoriaMetrics Monitoring", "Grafana + Loki + Prometheus Monitoring using Docker", "No"},
		selectedMonitoring:       "Kubernetes + VictoriaMetrics Monitoring",
		docker:                   []string{},
	}
}

func (m model) Init() tea.Cmd {
	var validate = func() tea.Msg {
		email, apiKey, err := readConfig()
		if err != nil {
			return authMsg{email: "", apiKey: ""}
		}
		err = validateConfig(email, apiKey)
		if err != nil {
			return authMsg{email: "", apiKey: ""}
		}
		return authMsg{email: email, apiKey: apiKey}
	}
	return tea.Batch(textinput.Blink, m.spinner.Tick, validate)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEnter:
			m.err = nil
			if m.step == authStep {
				email := m.emailInput.Value()
				apiKey := m.apiKeyInput.Value()
				m.focusIndex = 0
				return m, checkConfig(email, apiKey)
			} else if m.step == startStep {
				m.step = clientStep
				return m, textinput.Blink
			} else if m.step == clientStep {
				m.selectedClient = m.clients[m.focusIndex]
				m.focusIndex = 0
				m.step = startOptionStep
			} else if m.step == startOptionStep {
				m.selectedStartOption = m.startOptions[m.focusIndex]
				m.focusIndex = 0
				if m.selectedStartOption == "Generate base project (SQLite, Grafana Monitoring, Mocked payments, Local files, Log Emails)" {
					m.step = projectNameStep
					blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput, &m.projectNameInput})
					m.projectNameInput.Focus()
					m.projectNameInput.PromptStyle = focusedStyle
					m.projectNameInput.TextStyle = focusedStyle
					return m, textinput.Blink
				} else {
					m.step = databaseStep
				}
			} else if m.step == databaseStep {
				m.selectedDatabase = m.databases[m.focusIndex]
				m.focusIndex = 0
				m.step = paymentsProviderStep
			} else if m.step == paymentsProviderStep {
				m.selectedPaymentsProvider = m.paymentsProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = emailProviderStep
			} else if m.step == emailProviderStep {
				m.selectedEmailProvider = m.emailsProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = filesProviderStep
			} else if m.step == filesProviderStep {
				m.selectedFilesProvider = m.filesProviders[m.focusIndex]
				m.focusIndex = 0
				m.step = monitoringStep
			} else if m.step == monitoringStep {
				m.selectedMonitoring = m.monitoringOptions[m.focusIndex]
				m.focusIndex = 0
				m.step = projectNameStep

				blurAll([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
				m.projectNameInput.Focus()
				m.projectNameInput.PromptStyle = focusedStyle
				m.projectNameInput.TextStyle = focusedStyle
				return m, textinput.Blink
			} else if m.step == projectNameStep {
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
				m.step = cleaningStep
				return m, m.downloadRepo(m.email, m.apiKey, projectName)
			} else if m.step == successStep {
				return m, tea.Quit
			}
			return m, cmd

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			if m.step == authStep {
				cmd := m.toggleFocus([]*textinput.Model{&m.emailInput, &m.apiKeyInput})
				return m, cmd
			} else if m.step == clientStep || m.step == startOptionStep || m.step == databaseStep || m.step == paymentsProviderStep || m.step == emailProviderStep || m.step == filesProviderStep || m.step == monitoringStep {
				var d []string
				if m.step == clientStep {
					d = m.clients
				} else if m.step == startOptionStep {
					d = m.startOptions
				} else if m.step == databaseStep {
					d = m.databases
				} else if m.step == paymentsProviderStep {
					d = m.paymentsProviders
				} else if m.step == emailProviderStep {
					d = m.emailsProviders
				} else if m.step == filesProviderStep {
					d = m.filesProviders
				} else if m.step == monitoringStep {
					d = m.monitoringOptions
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
			m.step = authStep
			return m, textinput.Blink
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case authMsg:
		email := msg.email
		apiKey := msg.apiKey
		if email == "" || apiKey == "" {
			m.step = authStep
			return m, nil
		}
		m.apiKey = apiKey
		m.email = email
		m.step = startStep
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	case copyMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		} else {
			m.step = finishStep
		}
		return m, m.cleaningRepo()
	case finishMsg:
		m.step = successStep
		m.docker = msg.docker
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

func (m *model) downloadRepo(email string, apiKey string, projectName string) tea.Cmd {
	return func() tea.Msg {
		// If test env, copy ../gofast-app to the current directory
		if os.Getenv("TEST") == "true" {
			cmd := exec.Command("cp", "-r", "../gofast-app", "./"+projectName)
			err := cmd.Run()
			if err != nil {
				return errMsg(err)
			}
			return copyMsg{err: nil}
		}
		// get the file
		err := getFile(email, apiKey)
		if err != nil {
			return errMsg(err)
		}
		// unzip the file
		err = unzipFile()
		if err != nil {
			return errMsg(err)
		}
		// remove the zip file
		err = os.Remove("gofast-app.zip")
		if err != nil {
			return errMsg(err)
		}
		// find and rename the folder `gofast-live-gofast-app-...` to the project name
		files, err := os.ReadDir(".")
		if err != nil {
			return errMsg(err)
		}
		for _, f := range files {
			if f.IsDir() {
				if strings.HasPrefix(f.Name(), "gofast-live-gofast-app-") {
					err := os.Rename(f.Name(), projectName)
					if err != nil {
						return errMsg(err)
					}
					break
				}
			}
		}

		return copyMsg{err: nil}
	}
}

// get the file
func getFile(email string, apiKey string) error {
	client := http.Client{}
	req, err := http.NewRequest("GET", SERVER_URL+"/download?email="+email, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error downloading repo")
	}
	defer resp.Body.Close()

	// save the file to the disk
	_, err = os.Create("gofast-app.zip")
	if err != nil {
		return err
	}
	file, err := os.OpenFile("gofast-app.zip", os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// unzip the file
func unzipFile() error {
	if os.Getenv("TEST") == "true" {
		return nil
	}
	archive, err := zip.OpenReader("gofast-app.zip")
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, file := range archive.File {
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(file.Name, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		dst, err := os.Create(file.Name)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *model) cleaningRepo() tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		d, err := cleaning(m.projectNameInput.Value(), m.selectedClient, m.selectedStartOption, m.selectedDatabase, m.selectedPaymentsProvider, m.selectedEmailProvider, m.selectedFilesProvider, m.selectedMonitoring)
		if err != nil {
			return errMsg(err)
		}
		elapsed := time.Since(now)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
		return finishMsg{docker: d}
	}
}
