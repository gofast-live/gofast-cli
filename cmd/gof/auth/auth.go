package auth

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
)

func Run() {
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	m, ok := finalModel.(model)
	if !ok {
		log.Fatalf("unexpected bubbletea model type %T", finalModel)
	}
	if m.authenticated {
		fmt.Fprintln(os.Stdout, config.SuccessStyle().Render("Authentication successful!"))
	} else {
		fmt.Fprintln(os.Stdout, "Authentication cancelled.")
	}
}
