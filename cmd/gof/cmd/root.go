package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gof",
		Short: "GoFast CLI",
		Long: `
	GoFast CLI is a command line builder for Go related projects.
	Complete documentation is available at https://docs.gofast.live.
	For any issues, suggestions, or help, please visit our Discord server at https://discord.com/invite/EdSZbQbRyJ.
	`,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "Welcome to GoFast! Use 'gof help' for more information.")
		},
	}
}

func Execute() {
	addCmd := newAddCmd()
	addCmd.AddCommand(newAddStripeCmd(), newAddS3Cmd(), newAddPostmarkCmd())

	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newAuthCmd(),
		newInitCmd(),
		addCmd,
		newClientCmd(),
		newModelCmd(),
		newInfraCmd(),
		newMonCmd(),
		newVersionCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
