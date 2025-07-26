package cmd

import (
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
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
		auth.Run()
	},
}

