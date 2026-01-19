package integrations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// R2Strip removes all R2/files-related code from a freshly initialized project.
// Called by init command after downloading the template.
func R2Strip(projectPath string) error {
	// 1. Remove file domain folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "domain", "file")); err != nil {
		return fmt.Errorf("removing file domain: %w", err)
	}

	// 2. Remove file transport folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "transport", "file")); err != nil {
		return fmt.Errorf("removing file transport: %w", err)
	}

	// 3. Remove files migration
	if err := os.Remove(filepath.Join(projectPath, "app", "service-core", "storage", "migrations", "00004_create_files.sql")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing files migration: %w", err)
	}

	// 4. Strip all GF_FILE marker blocks from all files
	if err := StripIntegration(projectPath, "FILE"); err != nil {
		return fmt.Errorf("stripping file markers: %w", err)
	}

	return nil
}

// R2StripClient removes R2-related content from the Svelte client.
// Called by 'gof client svelte' command after copying the client folder.
func R2StripClient(clientPath string) error {
	// Remove files route folder
	filesPath := filepath.Join(clientPath, "src", "routes", "(app)", "files")
	if err := os.RemoveAll(filesPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing files folder: %w", err)
	}
	return nil
}

// R2AddClient adds R2-related content to an existing Svelte client.
// Called by 'gof add r2' when client already exists.
func R2AddClient(tmpProject, clientPath string) error {
	// Copy files route folder
	srcFiles := filepath.Join(tmpProject, "app", "service-client", "src", "routes", "(app)", "files")
	dstFiles := filepath.Join(clientPath, "src", "routes", "(app)", "files")
	if err := CopyDir(srcFiles, dstFiles); err != nil {
		return fmt.Errorf("copying files folder: %w", err)
	}
	return nil
}

// R2StripE2E removes R2-related e2e tests.
// Called by 'gof client svelte' when r2 is not enabled.
func R2StripE2E(e2ePath string) error {
	if err := os.Remove(filepath.Join(e2ePath, "files.test.ts")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing files.test.ts: %w", err)
	}
	return nil
}

// R2AddE2E adds R2-related e2e tests.
// Called by 'gof add r2' when client exists.
func R2AddE2E(tmpProject, e2ePath string) error {
	src := filepath.Join(tmpProject, "e2e", "files.test.ts")
	dst := filepath.Join(e2ePath, "files.test.ts")
	if err := CopyFile(src, dst); err != nil {
		return fmt.Errorf("copying files.test.ts: %w", err)
	}
	return nil
}

// R2Add adds Cloudflare R2 file storage integration to an existing project.
// Called by 'gof add r2' command.
func R2Add(email, apiKey string) error {
	// 1. Download template to temp location
	tmpDir, err := os.MkdirTemp("", "gofast-files-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Save current directory and chdir to tmpDir for download
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current dir: %w", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		return fmt.Errorf("changing to temp dir: %w", err)
	}

	if err := repo.DownloadRepo(email, apiKey, "template"); err != nil {
		_ = os.Chdir(cwd)
		return fmt.Errorf("downloading template: %w", err)
	}

	// Return to original directory
	if err := os.Chdir(cwd); err != nil {
		return fmt.Errorf("returning to original dir: %w", err)
	}

	tmpProject := filepath.Join(tmpDir, "template")

	// 2. Copy file domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "file")
	dstDomain := filepath.Join("app", "service-core", "domain", "file")
	if err := CopyDir(srcDomain, dstDomain); err != nil {
		return fmt.Errorf("copying file domain: %w", err)
	}

	// 3. Copy file transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "file")
	dstTransport := filepath.Join("app", "service-core", "transport", "file")
	if err := CopyDir(srcTransport, dstTransport); err != nil {
		return fmt.Errorf("copying file transport: %w", err)
	}

	// 4. Copy and renumber files migration
	if err := AddMigration(tmpProject, "00004_create_files.sql", "create_files.sql"); err != nil {
		return fmt.Errorf("adding files migration: %w", err)
	}

	// 5. Copy files with GF_FILE markers from template
	if err := CopyFilesWithMarkers(tmpProject, ".", "FILE"); err != nil {
		return fmt.Errorf("copying files with FILE markers: %w", err)
	}

	// 6. Add client-side Files content if Svelte client is configured
	if config.IsSvelte() {
		clientPath := filepath.Join("app", "service-client")
		if err := R2AddClient(tmpProject, clientPath); err != nil {
			return fmt.Errorf("adding files to client: %w", err)
		}
		// Add e2e tests if e2e folder exists
		e2ePath := "e2e"
		if _, err := os.Stat(e2ePath); err == nil {
			if err := R2AddE2E(tmpProject, e2ePath); err != nil {
				return fmt.Errorf("adding r2 e2e tests: %w", err)
			}
		}
	}

	return nil
}
