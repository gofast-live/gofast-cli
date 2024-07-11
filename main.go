package main

import (
	"fmt"
	"gofast-cli/repo"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	p := tea.NewProgram(repo.InitialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

