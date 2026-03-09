package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/e2e"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/svelte"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/tanstack"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client [client_type]",
	Short: "Create a new client service",
	Long:  "Create a new client service connected to your Go service",
	Args:  cobra.ExactArgs(1),
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

		serviceType := args[0]
		spec, ok := clients.SpecFor(serviceType)
		if !ok {
			cmd.Println("Invalid service type. Valid types are: svelte, tanstack")
			return
		}

		if config.HasService(spec.Name) {
			cmd.Printf("%s service already exists.\n", spec.DisplayName)
			return
		}

		tmpDir, err := os.MkdirTemp("", "gofast-app-*")
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

		if err := copyComposeFile(tmpDir, srcRepoName, cwd, con.ProjectName, spec.ComposeFile); err != nil {
			cmd.Printf("Error copying %s: %v\n", spec.ComposeFile, err)
			return
		}

		srcClientPath := filepath.Join(tmpDir, srcRepoName, "app", spec.ServiceDir)
		dstClientPath := filepath.Join(cwd, "app", spec.ServiceDir)

		if _, err := os.Stat(srcClientPath); err != nil {
			cmd.Printf("Source client folder not found in template: %v\n", err)
			return
		}
		if _, err := os.Stat(dstClientPath); err == nil {
			if err := os.RemoveAll(dstClientPath); err != nil {
				cmd.Printf("Destination '%s' already exists and could not be removed: %v\n", dstClientPath, err)
				return
			}
		}
		if err := os.MkdirAll(filepath.Dir(dstClientPath), 0o755); err != nil {
			cmd.Printf("Error creating destination directory: %v\n", err)
			return
		}

		if err := os.Rename(srcClientPath, dstClientPath); err != nil {
			if copyErr := copyDir(srcClientPath, dstClientPath); copyErr != nil {
				cmd.Printf("Error copying client folder: %v (original move error: %v)\n", copyErr, err)
				return
			}
		}

		enabledIntegrations := make(map[string]bool)
		for _, integration := range con.Integrations {
			enabledIntegrations[integration] = true
		}

		if !enabledIntegrations["stripe"] {
			if err := integrations.StripeStripClient(spec.Name, dstClientPath); err != nil {
				cmd.Printf("Error stripping stripe from client: %v\n", err)
				return
			}
		}
		if !enabledIntegrations["s3"] {
			if err := integrations.S3StripClient(spec.Name, dstClientPath); err != nil {
				cmd.Printf("Error stripping s3 from client: %v\n", err)
				return
			}
		}
		if !enabledIntegrations["postmark"] {
			if err := integrations.PostmarkStripClient(spec.Name, dstClientPath); err != nil {
				cmd.Printf("Error stripping postmark from client: %v\n", err)
				return
			}
		}

		srcE2E := filepath.Join(tmpDir, srcRepoName, "e2e")
		dstE2E := filepath.Join(cwd, "e2e")
		if _, err := os.Stat(srcE2E); err == nil {
			if err := copyDir(srcE2E, dstE2E); err != nil {
				cmd.Printf("Error copying e2e folder: %v\n", err)
				return
			}
			if !enabledIntegrations["stripe"] {
				if err := integrations.StripeStripE2E(dstE2E); err != nil {
					cmd.Printf("Error stripping stripe from e2e: %v\n", err)
					return
				}
			}
			if !enabledIntegrations["s3"] {
				if err := integrations.S3StripE2E(dstE2E); err != nil {
					cmd.Printf("Error stripping s3 from e2e: %v\n", err)
					return
				}
			}
			if !enabledIntegrations["postmark"] {
				if err := integrations.PostmarkStripE2E(dstE2E); err != nil {
					cmd.Printf("Error stripping postmark from e2e: %v\n", err)
					return
				}
			}
		}

		if err := os.Chdir(cwd); err != nil {
			cmd.Printf("Error changing back to original directory: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Printf("Adding %s client service...\n", spec.DisplayName)

		for _, m := range con.Models {
			if m.Name == "skeleton" {
				continue
			}

			cmd.Printf("Generating pages for '%s'...\n", m.Name)

			e2eColumns := make([]e2e.Column, len(m.Columns))
			for i, col := range m.Columns {
				e2eColumns[i] = e2e.Column{Name: col.Name, Type: col.Type}
			}
			if err := e2e.GenerateClientE2ETest(m.Name, e2eColumns); err != nil {
				cmd.Printf("Error generating e2e test for '%s': %v\n", m.Name, err)
				return
			}

			if err := generateClientScaffolding(spec.Name, m.Name, m.Columns); err != nil {
				cmd.Printf("Error generating '%s' client pages: %v\n", m.Name, err)
				return
			}
		}

		if err := formatClientProject(spec.Name); err != nil {
			cmd.Printf("Error formatting %s client: %v\n", spec.DisplayName, err)
			return
		}

		if err := config.AddService(spec.Name, spec.Port); err != nil {
			cmd.Printf("Error updating %s: %v\n", config.ConfigFileName, err)
			return
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render(spec.DisplayName + " client added successfully!"))
		cmd.Println("")

		var routes []string
		for _, m := range con.Models {
			if m.Name == "skeleton" {
				continue
			}
			routes = append(routes, clientModelPath(spec.Name, m.Name))
		}
		if enabledIntegrations["stripe"] {
			routes = append(routes, "/payments")
		}
		if enabledIntegrations["s3"] {
			routes = append(routes, "/files")
		}
		if enabledIntegrations["postmark"] {
			routes = append(routes, "/emails")
		}
		if len(routes) > 0 {
			cmd.Println("Add these routes to your navigation:")
			for _, route := range routes {
				cmd.Printf("  %s\n", config.SuccessStyle.Render(route))
			}
			cmd.Println("")
		}

		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		switch spec.Name {
		case clients.Svelte:
			cmd.Printf("  2. Run %s to launch your app with the Svelte client\n", config.SuccessStyle.Render("'make starts'"))
		case clients.Tanstack:
			cmd.Printf("  2. Run %s to launch your app with the TanStack client\n", config.SuccessStyle.Render("'make startt'"))
		}
		cmd.Println("")
	},
}

func generateClientScaffolding(clientType, modelName string, columns []config.Column) error {
	switch clientType {
	case clients.Svelte:
		svelteColumns := make([]svelte.Column, len(columns))
		for i, col := range columns {
			svelteColumns[i] = svelte.Column{Name: col.Name, Type: col.Type}
		}
		if err := svelte.GenerateSvelteScaffolding(modelName, svelteColumns); err != nil {
			return err
		}
		return svelte.UpdateUserPermissions(modelName)
	case clients.Tanstack:
		tanstackColumns := make([]tanstack.Column, len(columns))
		for i, col := range columns {
			tanstackColumns[i] = tanstack.Column{Name: col.Name, Type: col.Type}
		}
		return tanstack.GenerateTanstackScaffolding(modelName, tanstackColumns)
	default:
		return fmt.Errorf("unsupported client type %q", clientType)
	}
}

func formatClientProject(clientType string) error {
	switch clientType {
	case clients.Svelte:
		return svelte.FormatProject()
	case clients.Tanstack:
		return tanstack.FormatProject()
	default:
		return fmt.Errorf("unsupported client type %q", clientType)
	}
}

func clientModelPath(clientType, modelName string) string {
	switch clientType {
	case clients.Tanstack:
		return tanstack.GetModelPath(modelName)
	default:
		return svelte.GetModelPath(modelName)
	}
}

func copyComposeFile(tmpDir, srcRepoName, cwd, projectName, composeFile string) error {
	projCompose := filepath.Join(cwd, composeFile)
	srcCompose := filepath.Join(tmpDir, srcRepoName, composeFile)
	if err := copyFile(srcCompose, projCompose); err != nil {
		return err
	}
	content, err := os.ReadFile(projCompose)
	if err != nil {
		return err
	}
	updated := strings.ReplaceAll(string(content), "gofast", projectName)
	return os.WriteFile(projCompose, []byte(updated), 0o644)
}

// copyDir copies a directory recursively from src to dst.
func copyDir(src string, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	if err := os.MkdirAll(dst, fi.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		sPath := filepath.Join(src, e.Name())
		dPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(sPath, dPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(sPath, dPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sf.Close() }()

	sInfo, err := sf.Stat()
	if err != nil {
		return err
	}

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = df.Close() }()

	if _, err := io.Copy(df, sf); err != nil {
		return err
	}
	return nil
}
