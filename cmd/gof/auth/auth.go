package auth

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gofast-live/gofast-cli/cmd/gof/cmd"
)

func Run() {
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	m := finalModel.(model)
	if m.authenticated {
		fmt.Println(cmd.SuccessStyle.Render("Authentication successful!"))
	} else {
		fmt.Println("Authentication cancelled.")
	}
}
