package cmd

import (
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client [client_type]",
	Short: "Create a new client service",
	Long:  "Create a new client service (e.g., Svelte) connected to your Go service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, _, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}
		serviceType := args[0]
		validServiceTypes := map[string]bool{
			"svelte": true,
			"next":  true,
			"vue":    true,
		}
		if !validServiceTypes[serviceType] {
			cmd.Println("Invalid service type. Valid types are: svelte, next, vue")
			return
		}
		if serviceType != "next" && serviceType != "vue" {
			cmd.Printf("%s client not yet implemented.\n", serviceType)
			return
		}

	},
}

