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
	var postgresPort int
	initCmd := &cobra.Command{
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
			projectDir, projectName, err := parseProjectArg(args[0])
			if err != nil {
				cmd.Printf("%v\n", err)
				return
			}
			// check if the project directory already exists
			_, err = os.Stat(projectDir)
			if err == nil {
				cmd.Printf("Project directory '%s' already exists. Please choose a different name.\n", projectDir)
				return
			}
			if postgresPort < 1 || postgresPort > 65535 {
				cmd.Printf("Invalid --postgres-port %d: must be between 1 and 65535.\n", postgresPort)
				return
			}
			if hostPortInUse(cmd.Context(), postgresPort) {
				suggested := suggestFreePostgresPort(cmd.Context())
				cmd.Printf("Port %d is already in use.\n", postgresPort)
				cmd.Printf("Retry with: gof init %s --postgres-port %d\n", args[0], suggested)
				return
			}
			composeExists, err := composeProjectExists(cmd.Context(), projectName)
			if err != nil {
				cmd.Printf("Error checking existing Docker Compose projects: %v\n", err)
				return
			}
			if composeExists {
				cmd.Printf("Docker already has containers for a Compose project named '%s'.\n", strings.ToLower(projectName))
				cmd.Println("Choose a different project name, or remove that project's containers first.")
				return
			}

			// everything from rollbackRoot down is created by this run, so a failed run may delete it
			rollbackRoot := topmostMissingDir(projectDir)
			composeStarted := false
			success := false
			defer func() {
				if success {
					return
				}
				err := rollbackProject(cmd.Context(), projectDir, rollbackRoot, composeStarted)
				if err != nil {
					cmd.Printf("Warning: could not fully remove partial project at '%s': %v\n", rollbackRoot, err)
					return
				}
				cmd.Printf("Removed partial project at '%s'.\n", rollbackRoot)
			}()

			err = os.MkdirAll(filepath.Dir(projectDir), 0o755)
			if err != nil {
				cmd.Printf("Error creating parent directory: %v\n", err)
				return
			}
			err = repo.DownloadRepo(cmd.Context(), email, apiKey, projectDir)
			if err != nil {
				cmd.Printf("Error downloading repository: %v\n", err)
				return
			}

			removeTemplateOnlyFiles(cmd, projectDir)
			err = stripOptionalIntegrations(projectDir)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			err = configureCompose(projectDir, projectName, postgresPort)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			err = config.Initialize(projectDir, projectName)
			if err != nil {
				cmd.Printf("Error creating gofast.json file: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Printf("Initializing project '%s'...\n", projectName)
			composeStarted, err = runSetupScripts(cmd, projectDir)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}
			formatAndCommit(cmd, projectDir)

			success = true

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render("Project '" + projectName + "' initialized successfully!"))
			cmd.Println("")
			cmd.Println("Next steps:")
			cmd.Printf("  1. Run %s\n", config.SuccessStyle().Render("'cd "+projectDir+"'"))
			cmd.Printf("  2. Run %s to start the server\n", config.SuccessStyle().Render("'make start'"))
			cmd.Println("")
			cmd.Println("To create a GitHub repo:")
			cmd.Printf("  %s\n", config.SuccessStyle().Render("gh repo create "+projectName+" --private --source="+projectDir+" --push"))
			cmd.Println("")
		},
	}
	initCmd.Flags().IntVar(&postgresPort, "postgres-port", defaultPostgresHostPort, "Host port published for PostgreSQL")
	return initCmd
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
func removeTemplateOnlyFiles(cmd *cobra.Command, projectDir string) {
	err := os.RemoveAll(filepath.Join(projectDir, ".git"))
	if err != nil {
		cmd.Printf("Warning: could not remove template git metadata: %v\n", err)
	}
	for _, client := range clients.All() {
		err := os.RemoveAll(filepath.Join(projectDir, "app", client.ServiceDir))
		if err != nil {
			cmd.Printf("Warning: could not remove initial %s client folder: %v\n", client.DisplayName, err)
		}
	}
	err = os.RemoveAll(filepath.Join(projectDir, "monitoring"))
	if err != nil {
		cmd.Printf("Warning: could not remove monitoring folder: %v\n", err)
	}
	err = os.RemoveAll(filepath.Join(projectDir, "infra"))
	if err != nil {
		cmd.Printf("Warning: could not remove infra folder: %v\n", err)
	}
	err = os.Remove(filepath.Join(projectDir, "docker-compose.monitoring.yml"))
	if err != nil && !os.IsNotExist(err) {
		cmd.Printf("Warning: could not remove monitoring docker compose file: %v\n", err)
	}
	for _, client := range clients.All() {
		err := os.Remove(filepath.Join(projectDir, client.ComposeFile))
		if err != nil && !os.IsNotExist(err) {
			cmd.Printf("Warning: could not remove %s docker compose file: %v\n", client.DisplayName, err)
		}
	}
	err = os.RemoveAll(filepath.Join(projectDir, "e2e"))
	if err != nil {
		cmd.Printf("Warning: could not remove e2e folder: %v\n", err)
	}
	err = os.RemoveAll(filepath.Join(projectDir, ".github"))
	if err != nil {
		cmd.Printf("Warning: could not remove .github folder: %v\n", err)
	}
}

// stripOptionalIntegrations removes integrations the user can add back with 'gof add <integration>'.
func stripOptionalIntegrations(projectDir string) error {
	err := integrations.StripeStrip(projectDir)
	if err != nil {
		return fmt.Errorf("stripping stripe: %w", err)
	}
	err = integrations.S3Strip(projectDir)
	if err != nil {
		return fmt.Errorf("stripping s3: %w", err)
	}
	err = integrations.PostmarkStrip(projectDir)
	if err != nil {
		return fmt.Errorf("stripping postmark: %w", err)
	}
	return nil
}

// configureCompose renames the template's compose containers to the project
// and publishes Postgres on hostPort, keeping the Makefile migrate DSN in sync.
func configureCompose(projectDir, projectName string, hostPort int) error {
	dcPath := filepath.Join(projectDir, "docker-compose.yml")
	dcContent, err := os.ReadFile(dcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dcPath, err)
	}
	makefilePath := filepath.Join(projectDir, "Makefile")
	makefileContent, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", makefilePath, err)
	}
	newDcContent := applyProjectName(string(dcContent), projectName)
	newDcContent, newMakefileContent, err := applyHostPostgresPort(newDcContent, string(makefileContent), hostPort)
	if err != nil {
		return fmt.Errorf("applying postgres host port: %w", err)
	}
	err = os.WriteFile(dcPath, []byte(newDcContent), 0644)
	if err != nil {
		return fmt.Errorf("writing to %s: %w", dcPath, err)
	}
	if hostPort == defaultPostgresHostPort {
		return nil
	}
	err = os.WriteFile(makefilePath, []byte(newMakefileContent), 0644)
	if err != nil {
		return fmt.Errorf("writing to %s: %w", makefilePath, err)
	}
	return nil
}

// runSetupScripts generates keys and code, then migrates a temporary Postgres.
// composeStarted reports whether a docker compose step was reached, so a failed run knows to tear it down.
func runSetupScripts(cmd *cobra.Command, projectDir string) (composeStarted bool, err error) {
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
		if parts[0] == "docker" {
			composeStarted = true
		}
		cmdExec := exec.CommandContext(cmd.Context(), parts[0], parts[1:]...) //nolint:gosec // G204: scripts are fixed literals above, not user input
		cmdExec.Dir = projectDir
		output, err := cmdExec.CombinedOutput()
		if err != nil {
			return composeStarted, fmt.Errorf("running '%s': %w\nOutput: %s", script, err, output)
		}
	}
	return composeStarted, nil
}

// formatAndCommit runs go fmt and makes the initial git commit. Failures are warnings.
func formatAndCommit(cmd *cobra.Command, projectDir string) {
	gofmtCmd := exec.CommandContext(cmd.Context(), "go", "fmt", "./...")
	gofmtCmd.Dir = filepath.Join(projectDir, "app", "service-core")
	output, err := gofmtCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
	}

	gitInitCmd := exec.CommandContext(cmd.Context(), "git", "init")
	gitInitCmd.Dir = projectDir
	output, err = gitInitCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git init failed: %v\nOutput: %s\n", err, output)
	}
	gitAddCmd := exec.CommandContext(cmd.Context(), "git", "add", ".")
	gitAddCmd.Dir = projectDir
	output, err = gitAddCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git add failed: %v\nOutput: %s\n", err, output)
	}
	gitCommitCmd := exec.CommandContext(cmd.Context(), "git", "commit", "-m", "Initial commit")
	gitCommitCmd.Dir = projectDir
	output, err = gitCommitCmd.CombinedOutput()
	if err != nil {
		cmd.Printf("Warning: git commit failed: %v\nOutput: %s\n", err, output)
	}
}
