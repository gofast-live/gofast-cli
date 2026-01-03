package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
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

		projMonitoringCompose := filepath.Join(cwd, "docker-compose.monitoring.yml")
		srcMonitoringCompose := filepath.Join(srcRoot, "docker-compose.monitoring.yml")
		if _, err := os.Stat(projMonitoringCompose); err == nil {
			cmd.Printf("File '%s' already exists. Skipping copy.\n", projMonitoringCompose)
		} else {
			if err := copyFile(srcMonitoringCompose, projMonitoringCompose); err != nil {
				cmd.Printf("Error copying %s: %v\n", projMonitoringCompose, err)
				return
			}
			composeContent, err := os.ReadFile(projMonitoringCompose)
			if err != nil {
				cmd.Printf("Error reading %s: %v\n", projMonitoringCompose, err)
				return
			}
			newComposeContent := strings.ReplaceAll(string(composeContent), "gofast", con.ProjectName)
			info, err := os.Stat(projMonitoringCompose)
			if err != nil {
				cmd.Printf("Error getting file info for %s: %v\n", projMonitoringCompose, err)
				return
			}
			if err := os.WriteFile(projMonitoringCompose, []byte(newComposeContent), info.Mode()); err != nil {
				cmd.Printf("Error updating %s: %v\n", projMonitoringCompose, err)
				return
			}
		}

		srcInfraDir := filepath.Join(srcRoot, "infra")
		dstInfraDir := filepath.Join(cwd, "infra")
		if _, err := os.Stat(dstInfraDir); err == nil {
			cmd.Printf("Directory '%s' already exists. Skipping copy.\n", dstInfraDir)
		} else if err := copyDir(srcInfraDir, dstInfraDir); err != nil {
			cmd.Printf("Error copying infra directory: %v\n", err)
			return
		}

		srcMonitoringDir := filepath.Join(srcRoot, "monitoring")
		dstMonitoringDir := filepath.Join(cwd, "monitoring")
		if _, err := os.Stat(srcMonitoringDir); err == nil {
			if _, err := os.Stat(dstMonitoringDir); err == nil {
				cmd.Printf("Directory '%s' already exists. Skipping copy.\n", dstMonitoringDir)
			} else if err := copyDir(srcMonitoringDir, dstMonitoringDir); err != nil {
				cmd.Printf("Error copying monitoring directory: %v\n", err)
				return
			}
		} else {
			cmd.Printf("Warning: monitoring directory not found in template, skipping copy.\n")
		}

		err = os.Chdir(cwd)
		if err != nil {
			cmd.Printf("Error returning to project directory: %v\n", err)
			return
		}

		if err := integrations.ApplyInfraIntegrations(cwd, con.Integrations); err != nil {
			cmd.Printf("Error updating infra integrations: %v\n", err)
			return
		}

		err = config.MarkInfraPopulated()
		if err != nil {
			cmd.Printf("Error updating gofast config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Infrastructure files added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s\n", config.SuccessStyle.Render("'cd infra && cp .env.example .env'"))
		cmd.Println("  2. Update infra/.env with your server details")
		cmd.Println("  3. Review and run the setup scripts (setup_rke2.sh, setup_gh.sh, setup_gcp.sh, setup_cloudflare.sh)")
		cmd.Printf("  4. Run %s to launch your app with a local monitoring stack\n", config.SuccessStyle.Render("'make startm'"))
		cmd.Println("")
		cmd.Printf("See %s for the full workflow.\n", config.SuccessStyle.Render("'infra/README.md'"))
		if reqs := integrations.InfraRequirements(con.Integrations); len(reqs) > 0 {
			cmd.Println("")
			cmd.Println("GitHub integration secrets/vars to add:")
			for _, req := range reqs {
				if len(req.Secrets) > 0 {
					cmd.Printf("  %s secrets:\n", strings.ToUpper(req.Name))
					for _, name := range req.Secrets {
						cmd.Printf("    - %s\n", name)
					}
				}
				if len(req.Vars) > 0 {
					cmd.Printf("  %s vars:\n", strings.ToUpper(req.Name))
					for _, name := range req.Vars {
						cmd.Printf("    - %s\n", name)
					}
				}
			}
		}
		cmd.Println("")
	},
}
