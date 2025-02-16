package main

import (
	"strings"
)

func (m model) View() string {
	var b strings.Builder
	if m.step == 0 {
		b.WriteString("\n\n\n\n\n\n\n")
	} else if m.step == authStep {
		b.WriteRune('\n')
		b.WriteString("Enter your email address and API key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == validateStep {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View())
		b.WriteString(" Validating credentials")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == startStep {
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
	} else if m.step == clientStep || m.step == startOptionStep || m.step == paymentsProviderStep || m.step == emailProviderStep || m.step == filesProviderStep || m.step == monitoringStep {
		b.WriteRune('\n')
		if m.step == clientStep {
			b.WriteString("Choose a frontend framework\n\n")
		} else if m.step == startOptionStep {
			b.WriteString("Do you want to generate a ready-to-run project with pre-filled test secrets? (don't forget to change them later!)\n\n")
		} else if m.step == paymentsProviderStep {
			b.WriteString("Choose the payment provider you want to use\n\n")
		} else if m.step == emailProviderStep {
			b.WriteString("Choose the email provider you want to use\n\n")
		} else if m.step == filesProviderStep {
			b.WriteString("Choose the storage provider you want to use\n\n")
		} else if m.step == monitoringStep {
			b.WriteString("Choose your deployment/monitoring option\n\n")
		}
		var s string
		var d []string
		if m.step == clientStep {
			d = m.clients
		} else if m.step == startOptionStep {
			d = m.startOptions
		} else if m.step == paymentsProviderStep {
			d = m.paymentsProviders
		} else if m.step == emailProviderStep {
			d = m.emailsProviders
		} else if m.step == filesProviderStep {
			d = m.filesProviders
		} else if m.step == monitoringStep {
			d = m.monitoringOptions
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
		if m.step == clientStep {
			b.WriteString("")
		} else if m.step == startOptionStep {
			b.WriteString("\n")
			b.WriteString("\n")
		} else if m.step == paymentsProviderStep {
			b.WriteString("\n")
		} else if m.step == emailProviderStep {
		} else if m.step == filesProviderStep {
		} else if m.step == monitoringStep {
			b.WriteString("\n")
			b.WriteString("\n")
		}
	} else if m.step == projectNameStep {
		b.WriteRune('\n')
		b.WriteString("Enter the name of the project\n\n")
		b.WriteString(m.projectNameInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == cleaningStep || m.step == finishStep {
		b.WriteRune('\n')
		b.WriteString(m.spinner.View() + "  ")
		if m.step == cleaningStep {
			b.WriteString("Downloading project")
		} else if m.step == finishStep {
			b.WriteString("Cleaning project")
		}
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	} else if m.step == successStep {
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Project setup successfully!"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString("Now jump into the project directory:")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(focusedStyle.Render("cd " + m.projectNameInput.Value()))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString("Copy and paste the docker compose command into your terminal, change secrets and run it:")
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
		b.WriteString(successStyle.Render("Checkout our discord server for any help, feedback or suggestions: https://discord.com/invite/hFqr2SuVXA"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(focusedStyle.Render("Press enter to exit and start bulding the best project ever!"))
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
	b.WriteString(helpStyle.Render("Join our discord server for help: https://discord.com/invite/hFqr2SuVXA"))
	return b.String()
}
