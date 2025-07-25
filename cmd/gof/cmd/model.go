package cmd

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
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
	databaseProviders        []string
	selectedDatabaseProvider string
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
		startOptions:             []string{"Generate base project (Local PostgreSQL, Mocked payments, Log Emails, Local files)", "Start new configuration"},
		selectedStartOption:      "Generate base project (Local PostgreSQL, Mocked payments, Local files, Log Emails)",
		databaseProviders:        []string{"PostgreSQL", "SQLite", "Turso"},
		selectedDatabaseProvider: "PostgreSQL",
		paymentsProviders:        []string{"Local (mock)", "Stripe"},
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
