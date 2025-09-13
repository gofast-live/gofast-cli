package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [project_name]",
	Short: "Initialize the Go service",
	Long:  "Initialize the Go service with Docker and PostgreSQL setup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dependencies := map[string]string{
			"buf":    "https://buf.build/docs/cli/installation/",
			"atlas":  "https://atlasgo.io/getting-started",
			"sqlc":   "https://docs.sqlc.dev/en/latest/overview/install.html",
			"docker": "https://docs.docker.com/engine/install/",
		}

		var missingDeps []string
		for dep := range dependencies {
			_, err := exec.LookPath(dep)
			if err != nil {
				missingDeps = append(missingDeps, dep)
			}
		}

		// Check for docker-compose
		if _, err := exec.LookPath("docker"); err == nil {
			if err := exec.Command("docker", "compose", "version").Run(); err != nil {
				missingDeps = append(missingDeps, "docker compose")
				dependencies["docker compose"] = "https://docs.docker.com/compose/install/"
			}
		}

		if len(missingDeps) > 0 {
			cmd.Println("Missing dependencies:")
			for _, dep := range missingDeps {
				cmd.Printf("  - %s: %s\n", dep, dependencies[dep])
			}
			return
		}

		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}
		projectName := args[0]
		if projectName == "" {
			cmd.Println("Project name cannot be empty.")
			return
		}
		// check if the project directory already exists
		_, err = os.Stat(projectName)
		if err == nil {
			cmd.Printf("Project directory '%s' already exists. Please choose a different name.\n", projectName)
			return
		}
		// download the repository
		err = repo.DownloadRepo(email, apiKey, projectName)
		if err != nil {
			cmd.Printf("Error downloading repository: %v\n", err)
			return
		}
		// create gofast.json config using the config package
		if err := config.Initialize(projectName); err != nil {
			cmd.Printf("Error creating gofast.json file: %v\n", err)
			return
		}

		// run scripts to set up the project
		cmd.Printf("Running initialization scripts for project '%s'...\n", projectName)
		scripts := []string{
			"scripts/generate_keys.sh",
			"scripts/run_sqlc.sh",
			"scripts/run_buf.sh",
			"docker compose up postgres -d",
			"scripts/run_atlas.sh --auto-approve",
			"docker compose stop",
		}
		messages := []string{
			"Generating Public/Private keys...",
			"Generating SQLC files...",
			"Generating gRPC files...",
			"Starting PostgreSQL container...",
			"Running Atlas migrations...",
			"Stopping PostgreSQL container...",
		}
		for i, script := range scripts {
			cmd.Printf("%s\n", messages[i])
			var cmdExec *exec.Cmd
			if strings.HasPrefix(script, "docker") {
				parts := strings.Fields(script)
				cmdExec = exec.Command(parts[0], parts[1:]...)
			} else {
				parts := strings.Fields(script)
				scriptPath := fmt.Sprintf("./%s", parts[0])
				args := []string{scriptPath}
				if len(parts) > 1 {
					args = append(args, parts[1:]...)
				}
				cmdExec = exec.Command("sh", args...)
			}
			cmdExec.Dir = projectName
			output, err := cmdExec.CombinedOutput()
			if err != nil {
				cmd.Printf("Error running script '%s': %v\nOutput: %s\n", script, err, output)
				return
			}
		}

		cmd.Printf(
			"Project %s initialized successfully!\n\nCD into the %s directory and run %s to start the service.\n",
			config.SuccessStyle.Render("'"+projectName+"'"),
			config.SuccessStyle.Render("'"+projectName+"'"),
			config.SuccessStyle.Render("'docker compose up --build'"),
		)
	},
}

