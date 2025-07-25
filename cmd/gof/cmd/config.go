package cmd

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	SERVER_URL   = "https://admin.gofast.live"
	VERSION      = "v1.2.1"
	NoStyle      = lipgloss.NewStyle()
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ActiveStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	ErrStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	HelpStyle    = BlurredStyle
)
