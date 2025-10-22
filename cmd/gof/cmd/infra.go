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
	rootCmd.AddCommand(infraCmd)
}

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Add infrastructure files (otel compose and infra folder)",
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

		projOtelCompose := filepath.Join(cwd, "docker-compose.otel.yml")
		srcOtelCompose := filepath.Join(srcRoot, "docker-compose.otel.yml")
		if _, err := os.Stat(projOtelCompose); err == nil {
			cmd.Printf("File '%s' already exists. Skipping copy.\n", projOtelCompose)
		} else {
			if err := copyFile(srcOtelCompose, projOtelCompose); err != nil {
				cmd.Printf("Error copying %s: %v\n", projOtelCompose, err)
				return
			}
			composeContent, err := os.ReadFile(projOtelCompose)
			if err != nil {
				cmd.Printf("Error reading %s: %v\n", projOtelCompose, err)
				return
			}
			newComposeContent := strings.ReplaceAll(string(composeContent), "gofast", con.ProjectName)
			info, err := os.Stat(projOtelCompose)
			if err != nil {
				cmd.Printf("Error getting file info for %s: %v\n", projOtelCompose, err)
				return
			}
			if err := os.WriteFile(projOtelCompose, []byte(newComposeContent), info.Mode()); err != nil {
				cmd.Printf("Error updating %s: %v\n", projOtelCompose, err)
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

		srcOtelDir := filepath.Join(srcRoot, "otel")
		dstOtelDir := filepath.Join(cwd, "otel")
		if _, err := os.Stat(srcOtelDir); err == nil {
			if _, err := os.Stat(dstOtelDir); err == nil {
				cmd.Printf("Directory '%s' already exists. Skipping copy.\n", dstOtelDir)
			} else if err := copyDir(srcOtelDir, dstOtelDir); err != nil {
				cmd.Printf("Error copying otel directory: %v\n", err)
				return
			}
		} else {
			cmd.Printf("Warning: otel directory not found in template, skipping copy.\n")
		}

		startScriptPath := filepath.Join(cwd, "start.sh")
		startInfo, err := os.Stat(startScriptPath)
		if err != nil {
			cmd.Printf("Error locating %s: %v\n", startScriptPath, err)
			return
		}
		startContent, err := os.ReadFile(startScriptPath)
		if err != nil {
			cmd.Printf("Error reading %s: %v\n", startScriptPath, err)
			return
		}
		if !strings.Contains(string(startContent), "-f docker-compose.otel.yml") {
			lines := strings.Split(string(startContent), "\n")
			updated := false
			for i, line := range lines {
				if strings.Contains(line, "docker compose") && strings.Contains(line, " up") {
					upIdx := strings.LastIndex(line, " up")
					if upIdx == -1 {
						lines[i] = line + " -f docker-compose.otel.yml"
					} else {
						lines[i] = line[:upIdx] + " -f docker-compose.otel.yml" + line[upIdx:]
					}
					updated = true
					break
				}
			}
			if updated {
				if err := os.WriteFile(startScriptPath, []byte(strings.Join(lines, "\n")), startInfo.Mode()); err != nil {
					cmd.Printf("Error updating %s: %v\n", startScriptPath, err)
					return
				}
			} else {
				cmd.Printf("Warning: could not locate docker compose command in %s for otel compose update\n", startScriptPath)
			}
		}

		err = config.MarkInfraPopulated()
		if err != nil {
			cmd.Printf("Error updating gofast config: %v\n", err)
			return
		}

		cmd.Printf("Infrastructure files added successfully. Follow the guide in the %s to set up deployment.\n",
			config.SuccessStyle.Render("infra/README.md"),
		)
		cmd.Println("")
		cmd.Printf("Run %s to launch your app with a local monitoring stack.\n",
			config.SuccessStyle.Render("'sh start.sh'"),
		)
	},
}
