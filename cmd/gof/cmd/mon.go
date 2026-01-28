package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(monCmd)
}

var monCmd = &cobra.Command{
	Use:   "mon",
	Short: "Add monitoring stack (Grafana, Loki, Tempo, Prometheus)",
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
		if con.MonitoringPopulated {
			cmd.Println("Monitoring files have already been added to this project.")
			return
		}

		tmpDir, err := os.MkdirTemp("", "gofast-mon-*")
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

		// Copy docker-compose.monitoring.yml
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

		// Copy monitoring/ directory
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
			cmd.Printf("Error: monitoring directory not found in template.\n")
			return
		}

		// If infra was already added, copy monitoring.tf into it
		if con.InfraPopulated {
			srcMonitoringTf := filepath.Join(srcRoot, "infra", "monitoring.tf")
			dstMonitoringTf := filepath.Join(cwd, "infra", "monitoring.tf")
			if _, err := os.Stat(dstMonitoringTf); err == nil {
				cmd.Printf("File '%s' already exists. Skipping copy.\n", dstMonitoringTf)
			} else {
				if err := copyFile(srcMonitoringTf, dstMonitoringTf); err != nil {
					cmd.Printf("Error copying monitoring.tf: %v\n", err)
					return
				}
			}
		}

		err = os.Chdir(cwd)
		if err != nil {
			cmd.Printf("Error returning to project directory: %v\n", err)
			return
		}

		err = config.MarkMonitoringPopulated()
		if err != nil {
			cmd.Printf("Error updating gofast config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println("Adding monitoring stack...")
		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Monitoring stack added successfully!"))
		cmd.Println("")
		cmd.Println("Files added:")
		cmd.Printf("  - %s\n", config.SuccessStyle.Render("docker-compose.monitoring.yml"))
		cmd.Printf("  - %s\n", config.SuccessStyle.Render("monitoring/"))
		if con.InfraPopulated {
			cmd.Printf("  - %s\n", config.SuccessStyle.Render("infra/monitoring.tf"))
		}
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  Run %s to launch your app with local monitoring stack\n", config.SuccessStyle.Render("'make startm'"))
		cmd.Println("")
		cmd.Println("Access Grafana at http://localhost:3001 (no login required)")
		cmd.Println("")
		if !con.InfraPopulated {
			cmd.Printf("Run %s to add Kubernetes deployment files.\n", config.SuccessStyle.Render("'gof infra'"))
			cmd.Println("")
		}
	},
}
