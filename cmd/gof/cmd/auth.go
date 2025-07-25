package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GoFast CLI",
	Long:  "Authenticate with GoFast CLI",
	Run: func(cmd *cobra.Command, args []string) {
		m := initialModel()

		var b strings.Builder
		b.WriteRune('\n')
		b.WriteString("Enter your email address and API key\n\n")
		b.WriteString(m.emailInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(m.apiKeyInput.View())
		b.WriteRune('\n')
		b.WriteRune('\n')

		fmt.Println(b.String())

		// fmt.Println(SuccessStyle.Render("Authentication command is not implemented yet.\nPlease use 'gofast login' to authenticate with GoFast CLI."))
	},
}

