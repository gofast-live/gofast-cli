package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	// 1. Remove files route folder
	filesPath := filepath.Join(clientPath, "src", "routes", "(app)", "files")
	if err := os.RemoveAll(filesPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing files folder: %w", err)
	}

	// 2. Strip Files nav entry from layout
	layoutPath := filepath.Join(clientPath, "src", "routes", "(app)", "+layout.svelte")
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("reading layout: %w", err)
	}

	s := string(content)

	// Remove Files icon from import (FolderOpen or similar)
	s = strings.Replace(s, ", FolderOpen", "", 1)

	// Remove Files nav entry
	s = regexp.MustCompile(`(?s)\s*\{\s*name:\s*['"]Files['"],\s*href:\s*['"][^'"]+['"],\s*icon:\s*FolderOpen,?\s*\},?`).ReplaceAllString(s, "")

	if err := os.WriteFile(layoutPath, []byte(s), 0644); err != nil {
		return fmt.Errorf("writing layout: %w", err)
	}

	return nil
}

// R2AddClient adds R2-related content to an existing Svelte client.
// Called by 'gof add r2' when client already exists.
func R2AddClient(tmpProject, clientPath string) error {
	// 1. Copy files route folder
	srcFiles := filepath.Join(tmpProject, "app", "service-client", "src", "routes", "(app)", "files")
	dstFiles := filepath.Join(clientPath, "src", "routes", "(app)", "files")
	if err := CopyDir(srcFiles, dstFiles); err != nil {
		return fmt.Errorf("copying files folder: %w", err)
	}

	// 2. Add Files nav entry and icon import to layout
	layoutPath := filepath.Join(clientPath, "src", "routes", "(app)", "+layout.svelte")
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("reading layout: %w", err)
	}

	s := string(content)

	// Add FolderOpen to import
	if !strings.Contains(s, "FolderOpen") {
		s = regexp.MustCompile(`(from "@lucide[^"]*";)`).ReplaceAllString(s, `from "@lucide/svelte";
    import { FolderOpen } from "@lucide/svelte";`)
		s = strings.Replace(s, `} from "@lucide/svelte";
    import { FolderOpen } from "@lucide/svelte";`, `, FolderOpen } from "@lucide/svelte";`, 1)
	}

	// Add Files nav entry
	if !strings.Contains(s, `href: "/files"`) {
		s = regexp.MustCompile(`(\s*)(];)\s*\n(\s*function isActive)`).ReplaceAllString(s, `$1{
$1    name: "Files",
$1    href: "/files",
$1    icon: FolderOpen,
$1},
$1$2
$3`)
	}

	if err := os.WriteFile(layoutPath, []byte(s), 0644); err != nil {
		return fmt.Errorf("writing layout: %w", err)
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
	defer os.RemoveAll(tmpDir)

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
	}

	return nil
}
