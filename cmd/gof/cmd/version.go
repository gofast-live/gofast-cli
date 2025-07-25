package cmd

import (
	"fmt"

	"github.com/gofast-live/gofast-cli/cmd/gof/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoFast CLI",
	Long:  "Print the version number of GoFast CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.VERSION)
	},
}
