package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
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
			"sqlc":   "https://docs.sqlc.dev/en/latest/overview/install.html",
			"goose":  "https://github.com/pressly/goose#install",
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
		if err := os.RemoveAll(filepath.Join(projectName, ".git")); err != nil {
			cmd.Printf("Warning: could not remove template git metadata: %v\n", err)
		}
		// remove template-only folders and files
		if err := os.RemoveAll(filepath.Join(projectName, "app", "service-client")); err != nil {
			cmd.Printf("Warning: could not remove initial client folder: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(projectName, "monitoring")); err != nil {
			cmd.Printf("Warning: could not remove monitoring folder: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(projectName, "infra")); err != nil {
			cmd.Printf("Warning: could not remove infra folder: %v\n", err)
		}
		if err := os.Remove(filepath.Join(projectName, "docker-compose.monitoring.yml")); err != nil && !os.IsNotExist(err) {
			cmd.Printf("Warning: could not remove monitoring docker compose file: %v\n", err)
		}
		if err := os.Remove(filepath.Join(projectName, "docker-compose.client.yml")); err != nil && !os.IsNotExist(err) {
			cmd.Printf("Warning: could not remove client docker compose file: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(projectName, "e2e")); err != nil {
			cmd.Printf("Warning: could not remove e2e folder: %v\n", err)
		}
		// Strip optional integrations - user can add them back with 'gof add <integration>'
		if err := integrations.StripeStrip(projectName); err != nil {
			cmd.Printf("Error stripping stripe: %v\n", err)
			return
		}
		if err := integrations.R2Strip(projectName); err != nil {
			cmd.Printf("Error stripping r2: %v\n", err)
			return
		}
		if err := integrations.PostmarkStrip(projectName); err != nil {
			cmd.Printf("Error stripping postmark: %v\n", err)
			return
		}
		dcPath := filepath.Join(projectName, "docker-compose.yml")
		dcContent, err := os.ReadFile(dcPath)
		if err != nil {
			cmd.Printf("Error reading %s: %v\n", dcPath, err)
			return
		}
		newDcContent := strings.ReplaceAll(string(dcContent), "gofast", projectName)
		if err := os.WriteFile(dcPath, []byte(newDcContent), 0644); err != nil {
			cmd.Printf("Error writing to %s: %v\n", dcPath, err)
			return
		}

		// create gofast.json config using the config package
		if err := config.Initialize(projectName); err != nil {
			cmd.Printf("Error creating gofast.json file: %v\n", err)
			return
		}

		// run scripts to set up the project
		cmd.Println("")
		cmd.Printf("Running initialization scripts for project '%s'...\n", projectName)
		scripts := []string{
			"make keys",
			"make sql",
			"make gen",
			"docker compose up postgres -d --wait",
			"make migrate",
			"docker compose stop",
		}
		messages := []string{
			"Generating Public/Private keys...",
			"Generating SQL queries...",
			"Generating proto code...",
			"Starting PostgreSQL container...",
			"Applying database migrations...",
			"Stopping PostgreSQL container...",
		}
		for i, script := range scripts {
			cmd.Printf("%s\n", messages[i])
			parts := strings.Fields(script)
			cmdExec := exec.Command(parts[0], parts[1:]...)
			cmdExec.Dir = projectName
			output, err := cmdExec.CombinedOutput()
			if err != nil {
				cmd.Printf("Error running '%s': %v\nOutput: %s\n", script, err, output)
				return
			}
		}

		// Format Go code
		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = filepath.Join(projectName, "app", "service-core")
		if output, err := gofmtCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
		}

		// Initialize git repo with initial commit
		gitInitCmd := exec.Command("git", "init")
		gitInitCmd.Dir = projectName
		if output, err := gitInitCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git init failed: %v\nOutput: %s\n", err, output)
		}
		gitAddCmd := exec.Command("git", "add", ".")
		gitAddCmd.Dir = projectName
		if output, err := gitAddCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git add failed: %v\nOutput: %s\n", err, output)
		}
		gitCommitCmd := exec.Command("git", "commit", "-m", "Initial commit")
		gitCommitCmd.Dir = projectName
		if output, err := gitCommitCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git commit failed: %v\nOutput: %s\n", err, output)
		}

		cmd.Printf(
			"Project %s initialized successfully!\n\nCD into the %s directory and run %s.\n",
			config.SuccessStyle.Render("'"+projectName+"'"),
			config.SuccessStyle.Render("'"+projectName+"'"),
			config.SuccessStyle.Render("'make start'"),
		)
		cmd.Println("")
		cmd.Println("To create a GitHub repo:")
		cmd.Printf("  %s\n", config.SuccessStyle.Render("gh repo create "+projectName+" --private --source="+projectName+" --push"))
		cmd.Println("")
	},
}
