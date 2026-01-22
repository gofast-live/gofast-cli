package cmd

import (
	"os"
	"path/filepath"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infraCmd)
}

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Add infrastructure files (monitoring compose and infra folder)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}

		con, err := config.ParseConfig()
		if err != nil {
			cmd.Printf("%v\n", err)
			return
		}
		if con.InfraPopulated {
			cmd.Println("Infrastructure files have already been added to this project.")
			return
		}

		tmpDir, err := os.MkdirTemp("", "gofast-infra-*")
		if err != nil {
			cmd.Printf("Error creating temp directory: %v\n", err)
			return
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		cwd, err := os.Getwd()
		if err != nil {
			cmd.Printf("Error getting working directory: %v\n", err)
			return
		}

		if err := os.Chdir(tmpDir); err != nil {
			cmd.Printf("Error changing to temp directory: %v\n", err)
			return
		}
		defer func() { _ = os.Chdir(cwd) }()

		srcRepoName := "gofast-app-src"
		if err := repo.DownloadRepo(email, apiKey, srcRepoName); err != nil {
			cmd.Printf("Error downloading repository to temp directory: %v\n", err)
			return
		}

		srcRoot := filepath.Join(tmpDir, srcRepoName)

		srcInfraDir := filepath.Join(srcRoot, "infra")
		dstInfraDir := filepath.Join(cwd, "infra")
		if _, err := os.Stat(dstInfraDir); err == nil {
			cmd.Printf("Directory '%s' already exists. Skipping copy.\n", dstInfraDir)
		} else if err := copyDir(srcInfraDir, dstInfraDir); err != nil {
			cmd.Printf("Error copying infra directory: %v\n", err)
			return
		}

		// If monitoring hasn't been added yet, remove monitoring.tf from infra
		// It will be added when user runs 'gof mon'
		if !con.MonitoringPopulated {
			monitoringTf := filepath.Join(dstInfraDir, "monitoring.tf")
			if err := os.Remove(monitoringTf); err != nil && !os.IsNotExist(err) {
				cmd.Printf("Warning: could not remove monitoring.tf: %v\n", err)
			}
		}

		err = os.Chdir(cwd)
		if err != nil {
			cmd.Printf("Error returning to project directory: %v\n", err)
			return
		}

		err = config.MarkInfraPopulated()
		if err != nil {
			cmd.Printf("Error updating gofast config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println("Adding infrastructure files...")
		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Infrastructure files added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s\n", config.SuccessStyle.Render("'cd infra && cp .env.example .env'"))
		cmd.Println("  2. Update infra/.env with your server details")
		cmd.Println("  3. Review and run the setup scripts (setup_rke2.sh, setup_gh.sh, setup_cloudflare.sh)")
		cmd.Println("")
		cmd.Printf("See %s for the full workflow.\n", config.SuccessStyle.Render("'infra/README.md'"))
		cmd.Println("")
		if !con.MonitoringPopulated {
			cmd.Printf("Run %s to add local development monitoring stack.\n", config.SuccessStyle.Render("'gof mon'"))
			cmd.Println("")
		}
	},
}
