package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/spf13/cobra"
)

func newMonCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mon",
		Short: "Add monitoring stack (OTel Collector, VictoriaMetrics, VictoriaLogs, VictoriaTraces, Grafana)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
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

			err = os.Chdir(tmpDir)
			if err != nil {
				cmd.Printf("Error changing to temp directory: %v\n", err)
				return
			}
			defer func() { _ = os.Chdir(cwd) }()

			srcRepoName := "gofast-app-src"
			err = repo.DownloadRepo(cmd.Context(), email, apiKey, srcRepoName)
			if err != nil {
				cmd.Printf("Error downloading repository to temp directory: %v\n", err)
				return
			}

			srcRoot := filepath.Join(tmpDir, srcRepoName)

			err = installMonitoringCompose(cmd, filepath.Join(srcRoot, "docker-compose.monitoring.yml"), filepath.Join(cwd, "docker-compose.monitoring.yml"), con.ProjectName)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			err = installMonitoringDir(cmd, filepath.Join(srcRoot, "monitoring"), filepath.Join(cwd, "monitoring"))
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}

			// If infra was already added, copy monitoring.tf into it
			if con.InfraPopulated {
				srcMonitoringTf := filepath.Join(srcRoot, "infra", "monitoring.tf")
				dstMonitoringTf := filepath.Join(cwd, "infra", "monitoring.tf")
				_, err := os.Stat(dstMonitoringTf)
				if err == nil {
					cmd.Printf("File '%s' already exists. Skipping copy.\n", dstMonitoringTf)
				} else {
					err := copyFile(srcMonitoringTf, dstMonitoringTf)
					if err != nil {
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
			cmd.Println(config.SuccessStyle().Render("Monitoring stack added successfully!"))
			cmd.Println("")
			cmd.Println("Files added:")
			cmd.Printf("  - %s\n", config.SuccessStyle().Render("docker-compose.monitoring.yml"))
			cmd.Printf("  - %s\n", config.SuccessStyle().Render("monitoring/"))
			if con.InfraPopulated {
				cmd.Printf("  - %s\n", config.SuccessStyle().Render("infra/monitoring.tf"))
			}
			cmd.Println("")
			cmd.Println("Next steps:")
			cmd.Printf("  Run %s to launch your app with local monitoring stack\n", config.SuccessStyle().Render("'make startm'"))
			cmd.Println("")
			cmd.Println("Access Grafana at http://localhost:3001 (no login required)")
			cmd.Println("")
			if !con.InfraPopulated {
				cmd.Printf("Run %s to add Kubernetes deployment files.\n", config.SuccessStyle().Render("'gof infra'"))
				cmd.Println("")
			}
		},
	}
}

// installMonitoringCompose copies the monitoring compose file and renames the template's
// "gofast" container names to the project name. An existing file is left alone.
func installMonitoringCompose(cmd *cobra.Command, src, dst, projectName string) error {
	_, err := os.Stat(dst)
	if err == nil {
		cmd.Printf("File '%s' already exists. Skipping copy.\n", dst)
		return nil
	}
	err = copyFile(src, dst)
	if err != nil {
		return fmt.Errorf("copying %s: %w", dst, err)
	}
	composeContent, err := os.ReadFile(dst)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dst, err)
	}
	newComposeContent := strings.ReplaceAll(string(composeContent), "gofast", projectName)
	info, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("getting file info for %s: %w", dst, err)
	}
	err = os.WriteFile(dst, []byte(newComposeContent), info.Mode())
	if err != nil {
		return fmt.Errorf("updating %s: %w", dst, err)
	}
	return nil
}

// installMonitoringDir copies the template's monitoring/ folder. An existing folder is left alone.
func installMonitoringDir(cmd *cobra.Command, src, dst string) error {
	_, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("finding monitoring directory in template: %w", err)
	}
	_, err = os.Stat(dst)
	if err == nil {
		cmd.Printf("Directory '%s' already exists. Skipping copy.\n", dst)
		return nil
	}
	err = copyDir(src, dst)
	if err != nil {
		return fmt.Errorf("copying monitoring directory: %w", err)
	}
	return nil
}
