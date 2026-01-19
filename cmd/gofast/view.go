package main

import (
	"strings"
)

func (m model) View() string {
	var b strings.Builder
	switch m.step {
	case 0:
		b.WriteString("\n\n\n\n\n\n\n")
	case authStep:
		b.WriteRune('\n')
		b.WriteString("Enter your email address and API key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
	case validateStep:
		b.WriteRune('\n')
		b.WriteString(m.spinner.View())
		b.WriteString(" Validating credentials")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	case startStep:
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
	case clientStep, startOptionStep, databaseProviderStep, paymentsProviderStep, emailProviderStep, filesProviderStep, monitoringStep:
		b.WriteRune('\n')
		switch m.step {
		case clientStep:
			b.WriteString("Choose a frontend framework\n\n")
		case startOptionStep:
			b.WriteString("Do you want to generate a ready-to-run project with pre-filled test secrets? (don't forget to change them later!)\n\n")
		case databaseProviderStep:
			b.WriteString("Choose the database provider you want to use\n\n")
		case paymentsProviderStep:
			b.WriteString("Choose the payment provider you want to use\n\n")
		case emailProviderStep:
			b.WriteString("Choose the email provider you want to use\n\n")
		case filesProviderStep:
			b.WriteString("Choose the storage provider you want to use\n\n")
		case monitoringStep:
			b.WriteString("Choose your deployment/monitoring option\n\n")
		}
		var s string
		var d []string
		switch m.step {
		case clientStep:
			d = m.clients
		case startOptionStep:
			d = m.startOptions
		case databaseProviderStep:
			d = m.databaseProviders
		case paymentsProviderStep:
			d = m.paymentsProviders
		case emailProviderStep:
			d = m.emailsProviders
		case filesProviderStep:
			d = m.filesProviders
		case monitoringStep:
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
		switch m.step {
		case clientStep:
			b.WriteString("")
		case startOptionStep:
			b.WriteString("\n")
			b.WriteString("\n")
		case paymentsProviderStep:
			b.WriteString("\n")
		case emailProviderStep:
			// no extra spacing
		case filesProviderStep:
			// no extra spacing
		case monitoringStep:
			b.WriteString("\n")
			b.WriteString("\n")
		}
	case projectNameStep:
		b.WriteRune('\n')
		b.WriteString("Enter the name of the project\n\n")
		b.WriteString(m.projectNameInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	case cleaningStep, finishStep:
		b.WriteRune('\n')
		b.WriteString(m.spinner.View() + "  ")
		switch m.step {
		case cleaningStep:
			b.WriteString("Downloading project")
		case finishStep:
			b.WriteString("Cleaning project")
		}
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteRune('\n')
	case successStep:
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Project created successfully!"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString("Navigate to your new directory:")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(focusedStyle.Render("cd " + m.projectNameInput.Value()))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString("For more information on how to run the project, see the " + focusedStyle.Render("README.md") + " file.")
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Have fun building!"))
		b.WriteRune('\n')
		b.WriteString(successStyle.Render("Check out our discord server for any help, feedback, or suggestions: https://discord.com/invite/hFqr2SuVXA"))
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(focusedStyle.Render("Press Enter to exit and start creating something awesome!"))
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
	b.WriteString(helpStyle.Render("Join our discord server: https://discord.com/invite/hFqr2SuVXA"))
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(helpStyle.Render("Version: " + VERSION))
	return b.String()
}
