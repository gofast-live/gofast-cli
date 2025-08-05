package cmd

import (
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(modelCmd)
}

var modelCmd = &cobra.Command{
	Use:   "model [model_name]",
	Short: "Create a new model",
	Long:  "Create a new model including database migrations, query generation, validation, API endpoints and UI views.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := config.AddModel(args[0])
		if err != nil {
			cmd.Printf("Error adding model: %v.\n", err)
			return
		}
	},
}
