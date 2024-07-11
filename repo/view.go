package repo

import (
	"strings"
)

func (m model) View() string {
	var b strings.Builder
	s := ""
	if m.step == 1 {
		b.WriteRune('\n')
		b.WriteString("Enter your email address and api key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
	} else if m.step == 3 {
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Credentials validated successfully!"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Welcome to the GoFast CLI :)"))
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 4 {
		b.WriteRune('\n')
		b.WriteString("Enter the name of the project\n\n")
		b.WriteString(m.projectNameInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 2 || m.step == 5 || m.step == 6 {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View())
		if m.step == 2 {
			b.WriteString(" Validating project")
		} else if m.step == 5 {
			b.WriteString(" Downloading project")
		} else {
			b.WriteString(" Cloning project")
		}
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	}

	b.WriteRune('\n')
	b.WriteRune('\n')
	if m.err != nil {
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(s)
	b.WriteString(activeStyle.Render("enter"))
	b.WriteString(helpStyle.Render(" next"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("tab / up / down"))
	b.WriteString(helpStyle.Render(" navigate"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+q"))
	b.WriteString(helpStyle.Render(" config"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+c / esc"))
	b.WriteString(helpStyle.Render(" quit"))
	return b.String()
}
