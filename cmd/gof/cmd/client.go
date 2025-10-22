package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/svelte"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client [client_type]",
	Short: "Create a new client service",
	Long:  "Create a new client service (e.g., Svelte/Next/Vue) connected to your Go service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}
		// Ensure we are inside a valid gofast project (has gofast.json)
		con, err := config.ParseConfig()
		if err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		serviceType := args[0]
		validServiceTypes := map[string]bool{
			"svelte": true,
			"next":   true,
			"vue":    true,
		}
		if !validServiceTypes[serviceType] {
			cmd.Println("Invalid service type. Valid types are: svelte, next, vue")
			return
		}
		if serviceType == "vue" || serviceType == "next" {
			cmd.Println("Vue and Next clients are not implemented yet. Please use 'svelte' for now.")
			return
		}

		// Ensure gofast.json includes the svelte service on port 3000
		if serviceType == "svelte" {
			hasSvelte := false
			for _, svc := range con.Services {
				if svc.Name == "svelte" {
					hasSvelte = true
					break
				}
			}
			if !hasSvelte {
				con.Services = append(con.Services, config.Service{Name: "svelte", Port: "3000"})
				data, jerr := json.MarshalIndent(con, "", "  ")
				if jerr != nil {
					cmd.Printf("Error serializing config with svelte service: %v\n", jerr)
				} else if werr := os.WriteFile(config.ConfigFileName, data, 0644); werr != nil {
					cmd.Printf("Error writing %s: %v\n", config.ConfigFileName, werr)
				}
			}
		}

		// Prepare a temp workspace and download the template repo into it
		tmpDir, err := os.MkdirTemp("", "gofast-app-*")
		if err != nil {
			cmd.Printf("Error creating temp directory: %v\n", err)
			return
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Work within the temp directory
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

		// Download the full repo into the temp dir (no removal of client here)
		srcRepoName := "gofast-app-src"
		if err := repo.DownloadRepo(email, apiKey, srcRepoName); err != nil {
			cmd.Printf("Error downloading repository to temp directory: %v\n", err)
			return
		}

		// Ensure the client compose file is present and matches the project name.
		projClientCompose := filepath.Join(cwd, "docker-compose.client.yml")
		srcClientCompose := filepath.Join(tmpDir, srcRepoName, "docker-compose.client.yml")
		if err := copyFile(srcClientCompose, projClientCompose); err != nil {
			cmd.Printf("Error copying %s: %v\n", projClientCompose, err)
			return
		}
		clientComposeContent, err := os.ReadFile(projClientCompose)
		if err != nil {
			cmd.Printf("Error reading %s: %v\n", projClientCompose, err)
			return
		}
		newClientComposeContent := strings.ReplaceAll(string(clientComposeContent), "gofast", con.ProjectName)
		if err := os.WriteFile(projClientCompose, []byte(newClientComposeContent), 0644); err != nil {
			cmd.Printf("Error writing to %s: %v\n", projClientCompose, err)
			return
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
		if !strings.Contains(string(startContent), "-f docker-compose.client.yml") {
			lines := strings.Split(string(startContent), "\n")
			updated := false
			for i, line := range lines {
				if strings.Contains(line, "docker compose") && strings.Contains(line, " up") {
					upIdx := strings.LastIndex(line, " up")
					if upIdx == -1 {
						lines[i] = line + " -f docker-compose.client.yml"
					} else {
						lines[i] = line[:upIdx] + " -f docker-compose.client.yml" + line[upIdx:]
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
				cmd.Printf("Warning: could not locate docker compose command in %s for client compose update\n", startScriptPath)
			}
		}

		// Determine source client folder based on requested type
		var srcClientPath string
		switch serviceType {
		case "svelte":
			srcClientPath = filepath.Join(tmpDir, srcRepoName, "app", "service-client")
		case "next":
			srcClientPath = filepath.Join(tmpDir, srcRepoName, "app", "service-next")
		case "vue":
			srcClientPath = filepath.Join(tmpDir, srcRepoName, "app", "service-vue")
		}

		// Destination is always app/service-client inside the project
		dstClientPath := filepath.Join(cwd, "app", "service-client")

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
		// Ensure destination parent exists
		if err := os.MkdirAll(filepath.Dir(dstClientPath), 0o755); err != nil {
			cmd.Printf("Error creating destination directory: %v\n", err)
			return
		}

		// Try moving first. If cross-device rename fails, fall back to copy.
		if err := os.Rename(srcClientPath, dstClientPath); err != nil {
			if copyErr := copyDir(srcClientPath, dstClientPath); copyErr != nil {
				cmd.Printf("Error copying client folder: %v (original move error: %v)\n", copyErr, err)
				return
			}
		}

		// Change back to the original directory before generating svelte files.
		if err := os.Chdir(cwd); err != nil {
			cmd.Printf("Error changing back to original directory: %v\n", err)
			return
		}

		cmd.Printf("Generating client pages for existing models...\n")

		for _, m := range con.Models {
			if m.Name == "skeleton" {
				continue
			}

			svelteColumns := make([]svelte.Column, len(m.Columns))
			for i, col := range m.Columns {
				svelteColumns[i] = svelte.Column{
					Name: col.Name,
					Type: col.Type,
				}
			}

			if err := svelte.GenerateSvelteScaffolding(m.Name, svelteColumns); err != nil {
				cmd.Printf("Error generating '%s' client pages: %v\n", m.Name, err)
			} else {
				cmd.Printf("Successfully generated client pages for model '%s'\n", m.Name)
			}
		}

		cmd.Printf("Generating gRPC code...\n")
		bufCmd := exec.Command("sh", "scripts/run_grpc.sh")
		bufOut, err := bufCmd.CombinedOutput()
		if err != nil {
			cmd.Printf("Error running buf generation: %v\nOutput: %s\n", err, string(bufOut))
			return
		}

		cmd.Printf("Setup complete at %s.\n",
			config.SuccessStyle.Render("'app/service-client'"),
		)
	},
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
