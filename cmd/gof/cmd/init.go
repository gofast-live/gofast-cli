package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [project_name]",
		Short: "Initialize the Go service",
		Long:  "Initialize the Go service with Docker and PostgreSQL setup",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			missingDeps, dependencies := missingDependencies(cmd.Context())
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
			err = repo.DownloadRepo(cmd.Context(), email, apiKey, projectName)
			if err != nil {
				cmd.Printf("Error downloading repository: %v\n", err)
				return
			}
			removeTemplateOnlyFiles(cmd, projectName)
			err = stripOptionalIntegrations(projectName)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			err = renameComposeProject(projectName)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			err = config.Initialize(projectName)
			if err != nil {
				cmd.Printf("Error creating gofast.json file: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Printf("Initializing project '%s'...\n", projectName)
			err = runSetupScripts(cmd, projectName)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			formatAndCommit(cmd, projectName)

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render("Project '" + projectName + "' initialized successfully!"))
			cmd.Println("")
			cmd.Println("Next steps:")
			cmd.Printf("  1. Run %s\n", config.SuccessStyle().Render("'cd "+projectName+"'"))
			cmd.Printf("  2. Run %s to start the server\n", config.SuccessStyle().Render("'make start'"))
			cmd.Println("")
			cmd.Println("To create a GitHub repo:")
			cmd.Printf("  %s\n", config.SuccessStyle().Render("gh repo create "+projectName+" --private --source="+projectName+" --push"))
			cmd.Println("")
		},
	}
}

// missingDependencies returns the required tools that are not installed, with install links for each.
func missingDependencies(ctx context.Context) ([]string, map[string]string) {
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
	_, err := exec.LookPath("docker")
	if err == nil {
		err := exec.CommandContext(ctx, "docker", "compose", "version").Run()
		if err != nil {
			missingDeps = append(missingDeps, "docker compose")
			dependencies["docker compose"] = "https://docs.docker.com/compose/install/"
		}
	}
	return missingDeps, dependencies
}

// removeTemplateOnlyFiles drops folders that `gof client`, `gof mon` and `gof infra` add back later.
// Failures only leave extra files behind, so they are warnings.
func removeTemplateOnlyFiles(cmd *cobra.Command, projectName string) {
	err := os.RemoveAll(filepath.Join(projectName, ".git"))
	if err != nil {
		cmd.Printf("Warning: could not remove template git metadata: %v\n", err)
	}
	for _, client := range clients.All() {
		err := os.RemoveAll(filepath.Join(projectName, "app", client.ServiceDir))
		if err != nil {
			cmd.Printf("Warning: could not remove initial %s client folder: %v\n", client.DisplayName, err)
		}
	}
	err = os.RemoveAll(filepath.Join(projectName, "monitoring"))
	if err != nil {
		cmd.Printf("Warning: could not remove monitoring folder: %v\n", err)
	}
	err = os.RemoveAll(filepath.Join(projectName, "infra"))
	if err != nil {
		cmd.Printf("Warning: could not remove infra folder: %v\n", err)
	}
	err = os.Remove(filepath.Join(projectName, "docker-compose.monitoring.yml"))
	if err != nil && !os.IsNotExist(err) {
		cmd.Printf("Warning: could not remove monitoring docker compose file: %v\n", err)
	}
	for _, client := range clients.All() {
		err := os.Remove(filepath.Join(projectName, client.ComposeFile))
		if err != nil && !os.IsNotExist(err) {
			cmd.Printf("Warning: could not remove %s docker compose file: %v\n", client.DisplayName, err)
		}
	}
	err = os.RemoveAll(filepath.Join(projectName, "e2e"))
	if err != nil {
		cmd.Printf("Warning: could not remove e2e folder: %v\n", err)
	}
	err = os.RemoveAll(filepath.Join(projectName, ".github"))
	if err != nil {
		cmd.Printf("Warning: could not remove .github folder: %v\n", err)
	}
}

// stripOptionalIntegrations removes integrations the user can add back with 'gof add <integration>'.
func stripOptionalIntegrations(projectName string) error {
	err := integrations.StripeStrip(projectName)
	if err != nil {
		return fmt.Errorf("stripping stripe: %w", err)
	}
	err = integrations.S3Strip(projectName)
	if err != nil {
		return fmt.Errorf("stripping s3: %w", err)
	}
	err = integrations.PostmarkStrip(projectName)
	if err != nil {
		return fmt.Errorf("stripping postmark: %w", err)
	}
	return nil
}

// renameComposeProject replaces the template's "gofast" names in docker-compose.yml.
func renameComposeProject(projectName string) error {
	dcPath := filepath.Join(projectName, "docker-compose.yml")
	dcContent, err := os.ReadFile(dcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dcPath, err)
	}
	newDcContent := strings.ReplaceAll(string(dcContent), "gofast", projectName)
	err = os.WriteFile(dcPath, []byte(newDcContent), 0644)
	if err != nil {
		return fmt.Errorf("writing to %s: %w", dcPath, err)
	}
	return nil
}

// runSetupScripts generates keys and code, then migrates a temporary Postgres.
func runSetupScripts(cmd *cobra.Command, projectName string) error {
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
		cmdExec := exec.CommandContext(cmd.Context(), parts[0], parts[1:]...) //nolint:gosec // G204: scripts are fixed literals above, not user input
		cmdExec.Dir = projectName
		output, err := cmdExec.CombinedOutput()
		if err != nil {
			return fmt.Errorf("running '%s': %w\nOutput: %s", script, err, output)
		}
	}
	return nil
}

// formatAndCommit runs go fmt and makes the initial git commit. Failures are warnings.
func formatAndCommit(cmd *cobra.Command, projectName string) {
	gofmtCmd := exec.CommandContext(cmd.Context(), "go", "fmt", "./...")
	gofmtCmd.Dir = filepath.Join(projectName, "app", "service-core")
	output, err := gofmtCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
	}

	gitInitCmd := exec.CommandContext(cmd.Context(), "git", "init")
	gitInitCmd.Dir = projectName
	output, err = gitInitCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git init failed: %v\nOutput: %s\n", err, output)
	}
	gitAddCmd := exec.CommandContext(cmd.Context(), "git", "add", ".")
	gitAddCmd.Dir = projectName
	output, err = gitAddCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git add failed: %v\nOutput: %s\n", err, output)
	}
	gitCommitCmd := exec.CommandContext(cmd.Context(), "git", "commit", "-m", "Initial commit")
	gitCommitCmd.Dir = projectName
	output, err = gitCommitCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git commit failed: %v\nOutput: %s\n", err, output)
	}
}
