package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Uncomment the following lines to log to a file
	// f, err := tea.LogToFile("debug.log", "debug")
	// if err != nil {
	// 	fmt.Println("fatal:", err)
	// 	os.Exit(1)
	// }
	// defer f.Close()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
