package repo

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	var b strings.Builder
	s := ""
	if m.step == 1 {
        b.WriteRune('\n')
        b.WriteString("Enter your email address and api key\n\n")
        b.WriteString(m.emailInput.View())
        b.WriteString("\n\n")
        b.WriteString(m.apiKeyInput.View())
        b.WriteString("\n\n")

	} else if m.step == 2 {
        b.WriteRune('\n')
        b.WriteRune('\n')
        b.WriteRune('\n')
        b.WriteString(m.spinner.View())
		b.WriteString(" Validating credentials\n\n\n\n")
	} else if m.step == 3 {
        b.WriteString("Enter the name of the project\n\n")
        b.WriteString(m.projectNameInput.View())
        b.WriteString("\n\n")
	} else if m.step == 4 {
        b.WriteString("Copying repository ...\n\n\n\n")
	} else if m.step == 5 {
        b.WriteString("Finished\n\n\n\n")
	}
	if m.err != nil {
		s += fmt.Sprintf("\n%v\n\n",
			m.err,
		)
	}
	b.WriteString(s)
    b.WriteString(activeStyle.Render("enter"))
    b.WriteString(helpStyle.Render(" submit"))
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
