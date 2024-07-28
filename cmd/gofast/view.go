package main

import (
	"strings"
)

func (m model) View() string {
	var b strings.Builder
    if m.step == 0 {
        b.WriteString("\n\n\n\n\n")
    } else if m.step == 1 {
		b.WriteRune('\n')
		b.WriteString("Enter your email address and API key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
	} else if m.step == 2 {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View())
		b.WriteString(" Validating credentials")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 3 {
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Credentials validated successfully!"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Welcome to the GoFast CLI :)"))
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 4 || m.step == 5 || m.step == 6 || m.step == 7 || m.step == 8 || m.step == 9 {
		b.WriteRune('\n')
		if m.step == 4 {
			b.WriteString("Choose the protocol you want to use\n\n")
		} else if m.step == 5 {
			b.WriteString("Choose a frontend framework\n\n")
		} else if m.step == 6 {
			b.WriteString("Choose the database you want to use\n\n")
		} else if m.step == 7 {
			b.WriteString("Choose the payment provider you want to use\n\n")
		} else if m.step == 8 {
			b.WriteString("Choose the email provider you want to use\n\n")
		} else if m.step == 9 {
			b.WriteString("Choose the storage provider you want to use\n\n")
		}
		var s string
		var d []string
		if m.step == 4 {
			d = m.protocols
		} else if m.step == 5 {
			d = m.clients
		} else if m.step == 6 {
			d = m.databases
		} else if m.step == 7 {
			d = m.paymentsProviders
		} else if m.step == 8 {
			d = m.emailsProviders
		} else if m.step == 9 {
			d = m.filesProviders
		}
		for i, c := range d {
			cursor := " "
			if m.focusIndex == i {
				cursor = focusedStyle.Render(">")
			}
			checked := " "
			if d[m.focusIndex] == c {
				checked = focusedStyle.Render("•")
			}
			s += cursor + " [" + checked + "] " + c + "\n"
		}
		b.WriteString(s)
	} else if m.step == 10 {
		b.WriteRune('\n')
		b.WriteString("Enter the name of the project\n\n")
		b.WriteString(m.projectNameInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 11 || m.step == 12 || m.step == 13 {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View() + "  ")
		if m.step == 11 {
			b.WriteString("Downloading project")
		} else if m.step == 12 {
			b.WriteString("Cleaning project")
		} else if m.step == 13 {
			b.WriteString("Project downloaded successfully! Click enter to continue.")
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
	b.WriteString(activeStyle.Render("enter"))
	b.WriteString(helpStyle.Render(" next"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("space"))
	b.WriteString(helpStyle.Render(" select"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("tab / shift+tab / up / down"))
	b.WriteString(helpStyle.Render(" navigate"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+q"))
	b.WriteString(helpStyle.Render(" config"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+c / esc"))
	b.WriteString(helpStyle.Render(" quit"))
	return b.String()
}
