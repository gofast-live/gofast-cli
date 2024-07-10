package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

var GITHUB_URL = "https://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@github.com/gofast-live/gofast-app.git"

type (
	errMsg            error
	emailValidated    string
	keyValidated      string
	githubFinishedMsg struct{ err error }
)

type model struct {
	step             int
	emailInput       textinput.Model
	apiKeyInput      textinput.Model
	projectNameInput textinput.Model
	err              error
}

func initialModel() model {
	ei := textinput.New()
	ei.Placeholder = "Enter your email address"
	ei.Focus()
	ei.CharLimit = 156
	ei.Width = 40

	ai := textinput.New()
	ai.Placeholder = "Enter your API key"
	ai.CharLimit = 156
	ai.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Enter your project name"
	pi.CharLimit = 156
	pi.Width = 40

	return model{
		step:             1,
		emailInput:       ei,
		apiKeyInput:      ai,
		projectNameInput: pi,
		err:              nil,
	}
}

func (m model) Blink() tea.Cmd {
	return textinput.Blink
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

	return m.Blink()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.step == 1 {
				email := m.emailInput.Value()
				return m, checkEmail(email)
            } else if m.step == 2 {
                apiKey := m.apiKeyInput.Value()
                return m, checkKey(apiKey)
			} else if m.step == 3 {
				m.step = 4
				return m, copyRepo(m.projectNameInput.Value())
			}
			return m, cmd
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil

	case emailValidated:
		email := string(msg)
		saveToConfig("email", email)
		m.step = 2
		m.apiKeyInput.Focus()
		return m, m.Blink()

	case keyValidated:
		key := string(msg)
		saveToConfig("api_key", key)
		m.step = 3
		m.projectNameInput.Focus()
		return m, m.Blink()

	case githubFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		} else {
			m.step = 5
		}
	}

	if m.step == 1 {
		m.emailInput, cmd = m.emailInput.Update(msg)
	} else if m.step == 2 {
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
	} else if m.step == 3 {
		m.projectNameInput, cmd = m.projectNameInput.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	s := ""
	if m.step == 1 {
		s += fmt.Sprintf(
			"Step 1: Enter your email address\n\n%s",
			m.emailInput.View(),
		) + "\n\n"
	} else if m.step == 2 {
		s += fmt.Sprintf("Step 2: Enter your API key\n\n%s",
			m.apiKeyInput.View(),
		) + "\n\n"
	} else if m.step == 3 {
		s += fmt.Sprintf(
			"Step 3: Input project name\n\n%s",
			m.projectNameInput.View(),
		) + "\n\n"
	} else if m.step == 4 {
		s += "Step 4: Copying repository ...\n\n\n\n"
	} else if m.step == 5 {
		s += "Step 5: Finished\n\n\n\n"
	}
	if m.err != nil {
		s += fmt.Sprintf("\n%v\n\n",
			m.err,
		)
	}
	s += "(esc to quit)"
	return s
}

type Config struct {
	Email  string `json:"email"`
	ApiKey string `json:"api_key"`
}

func saveToConfig(key string, value string) {
	path, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	// marshal existing json
	// add new key value
	// write to file
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		// write empty json
		c = Config{}
	}
	if key == "email" {
		c.Email = value
	}
	if key == "api_key" {
		c.ApiKey = value
	}
	data, err = json.Marshal(c)
	if err != nil {
		panic(err)
	}
	_, err = jsonFile.WriteAt(data, 0)
	if err != nil {
		panic(err)
	}
}

func checkEmail(e string) tea.Cmd {
	return func() tea.Msg {
		if e == "" {
			return errMsg(fmt.Errorf("Email is required"))
		}
		if _, err := mail.ParseAddress(e); err != nil {
			return errMsg(fmt.Errorf("Invalid email address"))
		}
		return emailValidated(e)
	}
}

func checkKey(k string) tea.Cmd {
	return func() tea.Msg {
		if k == "" {
			return errMsg(fmt.Errorf("API key is required"))
		}
		return keyValidated(k)
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
