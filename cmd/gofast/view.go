package main

import (
	"strings"
)

func (m model) View() string {
	var b strings.Builder
	if m.step == 0 {
		b.WriteString("\n\n\n\n\n\n\n")
	} else if m.step == 1 {
		b.WriteRune('\n')
		b.WriteString("Enter your email address and API key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 2 {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View())
		b.WriteString(" Validating credentials")
		b.WriteRune('\n')
		b.WriteRune('\n')
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
		b.WriteString(successStyle.Render("Press enter to start the configuration"))
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 4 || m.step == 5 || m.step == 6 || m.step == 7 || m.step == 8 || m.step == 9 || m.step == 10 {
		b.WriteRune('\n')
		if m.step == 4 {
			b.WriteString("Choose the protocol you want to use\n\n")
		} else if m.step == 5 {
			b.WriteString("Choose a frontend framework\n\n")
		} else if m.step == 6 {
			b.WriteString("Do you want to generate base project? (with working secrets) \n\n")
		} else if m.step == 7 {
			b.WriteString("Choose the database you want to use\n\n")
		} else if m.step == 8 {
			b.WriteString("Choose the payment provider you want to use\n\n")
		} else if m.step == 9 {
			b.WriteString("Choose the email provider you want to use\n\n")
		} else if m.step == 10 {
			b.WriteString("Choose the storage provider you want to use\n\n")
		}
		var s string
		var d []string
		if m.step == 4 {
			d = m.protocols
		} else if m.step == 5 {
			d = m.clients
		} else if m.step == 6 {
			d = m.startOptions
		} else if m.step == 7 {
			d = m.databases
		} else if m.step == 8 {
			d = m.paymentsProviders
		} else if m.step == 9 {
			d = m.emailsProviders
		} else if m.step == 10 {
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
		if m.step == 4 {
			b.WriteString("\n\n")
		} else if m.step == 5 {
			b.WriteString("\n")
		} else if m.step == 6 {
			b.WriteString("\n")
			b.WriteString("\n")
		} else if m.step == 7 {
		} else if m.step == 8 {
		} else if m.step == 9 {
		} else if m.step == 10 {
		}
	} else if m.step == 11 {
		b.WriteRune('\n')
		b.WriteString("Enter the name of the project\n\n")
		b.WriteString(m.projectNameInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 12 || m.step == 13 {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View() + "  ")
		if m.step == 12 {
			b.WriteString("Downloading project")
		} else if m.step == 13 {
			b.WriteString("Cleaning project")
		}
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == 14 {
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Project setup successfully!"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString("Copy and replace the secrets in the following docker compose command to run the project:")
		b.WriteRune('\n')
		b.WriteString("(Can be found in README.md)")
		b.WriteRune('\n')
		b.WriteRune('\n')
		for _, v := range m.docker {
			b.WriteString(focusedStyle.Render(v))
			b.WriteRune('\n')
		}
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Thank you for using GoFast :)!"))
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Checkout our discord server for any help, feedback or suggestions: https://discord.gg/zqXEzmhT"))
	}

	if m.err != nil {
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(activeStyle.Render("enter"))
	b.WriteString(helpStyle.Render(" next"))
	b.WriteString(helpStyle.Render(" • "))
	// b.WriteString(activeStyle.Render("space"))
	// b.WriteString(helpStyle.Render(" select"))
	// b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("tab / shift+tab / up / down"))
	b.WriteString(helpStyle.Render(" navigate"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+q"))
	b.WriteString(helpStyle.Render(" change config"))
	b.WriteString(helpStyle.Render(" • "))
	b.WriteString(activeStyle.Render("ctrl+c / esc"))
	b.WriteString(helpStyle.Render(" quit"))
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(helpStyle.Render("Join our discord server for help: https://discord.gg/zqXEzmhT"))
	return b.String()
}
