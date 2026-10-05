package cmd

import (
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with GoFast CLI",
		Long:  "Authenticate with GoFast CLI",
		Run: func(_ *cobra.Command, _ []string) {
			auth.Run()
		},
	}
}
