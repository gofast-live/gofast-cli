package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

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
	emailMsg          string
	githubFinishedMsg struct{ err error }
)

type model struct {
	step             int
	emailInput       textinput.Model
	projectNameInput textinput.Model
	err              error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your email address"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Enter your project name"
	pi.CharLimit = 156
	pi.Width = 40

	return model{
		step:             1,
		emailInput:       ti,
		projectNameInput: pi,
		err:              nil,
	}
}

func (m model) Blink() tea.Cmd {
	return textinput.Blink
}

func (m model) Init() tea.Cmd {
	return m.Blink()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.step == 1 {
				m.step = 2
				email := m.emailInput.Value()
				return m, checkEmail(email)
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

	case emailMsg:
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
	} else if m.step == 3 {
		m.projectNameInput, cmd = m.projectNameInput.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	s := ""
	if m.err != nil {
		s += fmt.Sprintf("\nWe had some trouble: %v\n\n",
			m.err,
		)
	}
	if m.step == 1 {
		s += fmt.Sprintf(
			"Step 1: Enter your email address\n\n%s",
			m.emailInput.View(),
		) + "\n\n"
	} else if m.step == 2 {
		s += "Step 2: Checking your email address ...\n\n\n\n"
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
	s += "(esc to quit)"
	return s
}

func checkEmail(email string) tea.Cmd {
	return func() tea.Msg {
		if email == "" {
			return errMsg(fmt.Errorf("email is required"))
		}
		// mock wait 3 sec
		time.Sleep(1 * time.Second)
		return emailMsg("")
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
