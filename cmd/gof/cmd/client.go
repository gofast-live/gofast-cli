package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func newClientCmd() *cobra.Command {
	return &cobra.Command{
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

			err = copyComposeFile(tmpDir, srcRepoName, cwd, con.ProjectName, spec.ComposeFile)
			if err != nil {
				cmd.Printf("Error copying %s: %v\n", spec.ComposeFile, err)
				return
			}

			dstClientPath, err := installClientFolder(tmpDir, srcRepoName, cwd, spec)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}

			enabledIntegrations := make(map[string]bool)
			for _, integration := range con.Integrations {
				enabledIntegrations[integration] = true
			}
			err = stripDisabledClientIntegrations(spec.Name, dstClientPath, enabledIntegrations)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}

			err = copyE2EFolder(filepath.Join(tmpDir, srcRepoName, "e2e"), filepath.Join(cwd, "e2e"))
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}

			err = os.Chdir(cwd)
			if err != nil {
				cmd.Printf("Error changing back to original directory: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Printf("Adding %s client service...\n", spec.DisplayName)

			err = generateModelPages(cmd, spec.Name, con.Models)
			if err != nil {
				cmd.Printf("Error %v\n", err)
				return
			}

			err = formatClientProject(cmd.Context(), spec.Name)
			if err != nil {
				cmd.Printf("Error formatting %s client: %v\n", spec.DisplayName, err)
				return
			}

			err = config.AddService(spec.Name, spec.Port)
			if err != nil {
				cmd.Printf("Error updating %s: %v\n", config.ConfigFileName, err)
				return
			}

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render(spec.DisplayName + " client added successfully!"))
			cmd.Println("")

			printClientNextSteps(cmd, spec, con.Models, enabledIntegrations)
		},
	}
}

// installClientFolder moves the template's client folder into the project, replacing any existing one.
func installClientFolder(tmpDir, srcRepoName, cwd string, spec clients.Spec) (string, error) {
	srcClientPath := filepath.Join(tmpDir, srcRepoName, "app", spec.ServiceDir)
	dstClientPath := filepath.Join(cwd, "app", spec.ServiceDir)

	_, err := os.Stat(srcClientPath)
	if err != nil {
		return "", fmt.Errorf("finding source client folder in template: %w", err)
	}
	_, err = os.Stat(dstClientPath)
	if err == nil {
		err := os.RemoveAll(dstClientPath)
		if err != nil {
			return "", fmt.Errorf("removing existing %s: %w", dstClientPath, err)
		}
	}
	err = os.MkdirAll(filepath.Dir(dstClientPath), 0o755)
	if err != nil {
		return "", fmt.Errorf("creating destination directory: %w", err)
	}

	err = os.Rename(srcClientPath, dstClientPath)
	if err != nil {
		copyErr := copyDir(srcClientPath, dstClientPath)
		if copyErr != nil {
			return "", fmt.Errorf("copying client folder: %w (original move error: %w)", copyErr, err)
		}
	}
	return dstClientPath, nil
}

// stripDisabledClientIntegrations removes client code for integrations the project has not added.
func stripDisabledClientIntegrations(clientName, dstClientPath string, enabled map[string]bool) error {
	if !enabled["stripe"] {
		err := integrations.StripeStripClient(clientName, dstClientPath)
		if err != nil {
			return fmt.Errorf("stripping stripe from client: %w", err)
		}
	}
	if !enabled["s3"] {
		err := integrations.S3StripClient(clientName, dstClientPath)
		if err != nil {
			return fmt.Errorf("stripping s3 from client: %w", err)
		}
	}
	if !enabled["postmark"] {
		err := integrations.PostmarkStripClient(clientName, dstClientPath)
		if err != nil {
			return fmt.Errorf("stripping postmark from client: %w", err)
		}
	}
	return nil
}

// copyE2EFolder copies the template's e2e tests when the template has them.
func copyE2EFolder(srcE2E, dstE2E string) error {
	_, err := os.Stat(srcE2E)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("checking %s: %w", srcE2E, err)
	}
	err = copyDir(srcE2E, dstE2E)
	if err != nil {
		return fmt.Errorf("copying e2e folder: %w", err)
	}
	return nil
}

// generateModelPages generates client pages and e2e tests for every user model.
func generateModelPages(cmd *cobra.Command, clientName string, models []config.Model) error {
	for _, m := range models {
		if m.Name == "skeleton" {
			continue
		}

		cmd.Printf("Generating pages for '%s'...\n", m.Name)

		e2eColumns := make([]e2e.Column, len(m.Columns))
		for i, col := range m.Columns {
			e2eColumns[i] = e2e.Column{Name: col.Name, Type: col.Type}
		}
		err := e2e.GenerateClientE2ETest(m.Name, e2eColumns)
		if err != nil {
			return fmt.Errorf("generating e2e test for '%s': %w", m.Name, err)
		}

		err = generateClientScaffolding(clientName, m.Name, m.Columns)
		if err != nil {
			return fmt.Errorf("generating '%s' client pages: %w", m.Name, err)
		}
	}
	return nil
}

func printClientNextSteps(cmd *cobra.Command, spec clients.Spec, models []config.Model, enabled map[string]bool) {
	var routes []string
	for _, m := range models {
		if m.Name == "skeleton" {
			continue
		}
		routes = append(routes, clientModelPath(spec.Name, m.Name))
	}
	if enabled["stripe"] {
		routes = append(routes, "/payments")
	}
	if enabled["s3"] {
		routes = append(routes, "/files")
	}
	if enabled["postmark"] {
		routes = append(routes, "/emails")
	}
	if len(routes) > 0 {
		cmd.Println("Add these routes to your navigation:")
		for _, route := range routes {
			cmd.Printf("  %s\n", config.SuccessStyle().Render(route))
		}
		cmd.Println("")
	}

	cmd.Println("Next steps:")
	cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle().Render("'make gen'"))
	switch spec.Name {
	case clients.Svelte:
		cmd.Printf("  2. Run %s to launch your app with the Svelte client\n", config.SuccessStyle().Render("'make starts'"))
	case clients.Tanstack:
		cmd.Printf("  2. Run %s to launch your app with the TanStack client\n", config.SuccessStyle().Render("'make startt'"))
	}
	cmd.Println("")
}

func generateClientScaffolding(clientType, modelName string, columns []config.Column) error {
	switch clientType {
	case clients.Svelte:
		svelteColumns := make([]svelte.Column, len(columns))
		for i, col := range columns {
			svelteColumns[i] = svelte.Column{Name: col.Name, Type: col.Type}
		}
		err := svelte.GenerateSvelteScaffolding(modelName, svelteColumns)
		if err != nil {
			return fmt.Errorf("generating svelte pages: %w", err)
		}
		return nil
	case clients.Tanstack:
		tanstackColumns := make([]tanstack.Column, len(columns))
		for i, col := range columns {
			tanstackColumns[i] = tanstack.Column{Name: col.Name, Type: col.Type}
		}
		err := tanstack.GenerateTanstackScaffolding(modelName, tanstackColumns)
		if err != nil {
			return fmt.Errorf("generating tanstack pages: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported client type %q", clientType)
	}
}

func formatClientProject(ctx context.Context, clientType string) error {
	switch clientType {
	case clients.Svelte:
		err := svelte.FormatProject(ctx)
		if err != nil {
			return fmt.Errorf("formatting svelte client: %w", err)
		}
		return nil
	case clients.Tanstack:
		err := tanstack.FormatProject(ctx)
		if err != nil {
			return fmt.Errorf("formatting tanstack client: %w", err)
		}
		return nil
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
	err := copyFile(srcCompose, projCompose)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(projCompose)
	if err != nil {
		return fmt.Errorf("reading %s: %w", projCompose, err)
	}
	updated := strings.ReplaceAll(string(content), "gofast", projectName)
	err = os.WriteFile(projCompose, []byte(updated), 0o644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", projCompose, err)
	}
	return nil
}

// copyDir copies a directory recursively from src to dst.
func copyDir(src string, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	err = os.MkdirAll(dst, fi.Mode())
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("reading dir %s: %w", src, err)
	}
	for _, e := range entries {
		sPath := filepath.Join(src, e.Name())
		dPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			err := copyDir(sPath, dPath)
			if err != nil {
				return err
			}
			continue
		}
		err := copyFile(sPath, dPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer func() { _ = sf.Close() }()

	sInfo, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sInfo.Mode())
	if err != nil {
		return fmt.Errorf("opening %s: %w", dst, err)
	}
	defer func() { _ = df.Close() }()

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("copying %s to %s: %w", src, dst, err)
	}
	return nil
}
