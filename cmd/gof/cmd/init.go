package cmd

import (
	"github.com/gofast-live/gofast-cli/cmd/gof/auth"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Go service",
	Long:  "Initialize the Go service with Docker and PostgreSQL setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}
		cmd.Println("Go service initialized successfully. You can now run your Go application.")
	},
}

