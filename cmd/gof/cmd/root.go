package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gof",
	Short: "GoFast CLI",
	Long: `
GoFast CLI is a command line builder for Go related projects.
Complete documentation is available at https://docs.gofast.live.
For any issues, suggestions, or help, please visit our Discord server at https://discord.com/invite/EdSZbQbRyJ.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to GoFast! Use 'gof help' for more information.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
