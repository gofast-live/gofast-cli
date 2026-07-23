package cmd

import (
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

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().IntVar(&postgresPort, "postgres-port", defaultPostgresHostPort, "Host port published for PostgreSQL")
}

var postgresPort int

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

		projectDir, projectName, err := parseProjectArg(args[0])
		if err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		if _, err = os.Stat(projectDir); err == nil {
			cmd.Printf("Project directory '%s' already exists. Please choose a different name.\n", projectDir)
			return
		}

		if postgresPort < 1 || postgresPort > 65535 {
			cmd.Printf("Invalid --postgres-port %d: must be between 1 and 65535.\n", postgresPort)
			return
		}
		if hostPortInUse(postgresPort) {
			suggested := suggestFreePostgresPort()
			cmd.Printf("Port %d is already in use.\n", postgresPort)
			cmd.Printf("Retry with: gof init %s --postgres-port %d\n", args[0], suggested)
			return
		}

		err = repo.DownloadRepo(email, apiKey, projectDir)
		if err != nil {
			cmd.Printf("Error downloading repository: %v\n", err)
			return
		}

		success := false
		defer func() {
			if !success {
				rollbackProject(projectDir)
				cmd.Printf("Removed partial project at '%s'.\n", projectDir)
			}
		}()

		if err := os.RemoveAll(filepath.Join(projectDir, ".git")); err != nil {
			cmd.Printf("Warning: could not remove template git metadata: %v\n", err)
		}
		// remove template-only folders and files
		for _, client := range clients.All() {
			if err := os.RemoveAll(filepath.Join(projectDir, "app", client.ServiceDir)); err != nil {
				cmd.Printf("Warning: could not remove initial %s client folder: %v\n", client.DisplayName, err)
			}
		}
		if err := os.RemoveAll(filepath.Join(projectDir, "monitoring")); err != nil {
			cmd.Printf("Warning: could not remove monitoring folder: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(projectDir, "infra")); err != nil {
			cmd.Printf("Warning: could not remove infra folder: %v\n", err)
		}
		if err := os.Remove(filepath.Join(projectDir, "docker-compose.monitoring.yml")); err != nil && !os.IsNotExist(err) {
			cmd.Printf("Warning: could not remove monitoring docker compose file: %v\n", err)
		}
		for _, client := range clients.All() {
			if err := os.Remove(filepath.Join(projectDir, client.ComposeFile)); err != nil && !os.IsNotExist(err) {
				cmd.Printf("Warning: could not remove %s docker compose file: %v\n", client.DisplayName, err)
			}
		}
		if err := os.RemoveAll(filepath.Join(projectDir, "e2e")); err != nil {
			cmd.Printf("Warning: could not remove e2e folder: %v\n", err)
		}
		if err := os.RemoveAll(filepath.Join(projectDir, ".github")); err != nil {
			cmd.Printf("Warning: could not remove .github folder: %v\n", err)
		}
		// Strip optional integrations - user can add them back with 'gof add <integration>'
		if err := integrations.StripeStrip(projectDir); err != nil {
			cmd.Printf("Error stripping stripe: %v\n", err)
			return
		}
		if err := integrations.S3Strip(projectDir); err != nil {
			cmd.Printf("Error stripping s3: %v\n", err)
			return
		}
		if err := integrations.PostmarkStrip(projectDir); err != nil {
			cmd.Printf("Error stripping postmark: %v\n", err)
			return
		}

		dcPath := filepath.Join(projectDir, "docker-compose.yml")
		dcContent, err := os.ReadFile(dcPath)
		if err != nil {
			cmd.Printf("Error reading %s: %v\n", dcPath, err)
			return
		}
		makefilePath := filepath.Join(projectDir, "Makefile")
		makefileContent, err := os.ReadFile(makefilePath)
		if err != nil {
			cmd.Printf("Error reading %s: %v\n", makefilePath, err)
			return
		}

		newDcContent := applyProjectName(string(dcContent), projectName)
		newDcContent, newMakefileContent, err := applyHostPostgresPort(newDcContent, string(makefileContent), postgresPort)
		if err != nil {
			cmd.Printf("Error applying Postgres host port: %v\n", err)
			return
		}
		if err := os.WriteFile(dcPath, []byte(newDcContent), 0644); err != nil {
			cmd.Printf("Error writing to %s: %v\n", dcPath, err)
			return
		}
		if postgresPort != defaultPostgresHostPort {
			if err := os.WriteFile(makefilePath, []byte(newMakefileContent), 0644); err != nil {
				cmd.Printf("Error writing to %s: %v\n", makefilePath, err)
				return
			}
		}

		if err := config.Initialize(projectDir, projectName); err != nil {
			cmd.Printf("Error creating gofast.json file: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Printf("Initializing project '%s'...\n", projectName)
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
			cmdExec.Dir = projectDir
			output, runErr := cmdExec.CombinedOutput()
			if runErr != nil {
				cmd.Printf("Error running '%s': %v\nOutput: %s\n", script, runErr, output)
				return
			}
		}

		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = filepath.Join(projectDir, "app", "service-core")
		if output, err := gofmtCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
		}

		gitInitCmd := exec.Command("git", "init")
		gitInitCmd.Dir = projectDir
		if output, err := gitInitCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git init failed: %v\nOutput: %s\n", err, output)
		}
		gitAddCmd := exec.Command("git", "add", ".")
		gitAddCmd.Dir = projectDir
		if output, err := gitAddCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git add failed: %v\nOutput: %s\n", err, output)
		}
		gitCommitCmd := exec.Command("git", "commit", "-m", "Initial commit")
		gitCommitCmd.Dir = projectDir
		if output, err := gitCommitCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: git commit failed: %v\nOutput: %s\n", err, output)
		}

		success = true

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Project '" + projectName + "' initialized successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s\n", config.SuccessStyle.Render("'cd "+projectDir+"'"))
		cmd.Printf("  2. Run %s to start the server\n", config.SuccessStyle.Render("'make start'"))
		cmd.Println("")
		cmd.Println("To create a GitHub repo:")
		cmd.Printf("  %s\n", config.SuccessStyle.Render("gh repo create "+projectName+" --private --source="+projectDir+" --push"))
		cmd.Println("")
	},
}
