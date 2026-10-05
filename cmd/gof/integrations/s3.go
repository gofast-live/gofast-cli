package integrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// S3Strip removes all S3/files-related code from a freshly initialized project.
// Called by init command after downloading the template.
func S3Strip(projectPath string) error {
	// 1. Remove file domain folder
	err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "domain", "file"))
	if err != nil {
		return fmt.Errorf("removing file domain: %w", err)
	}

	// 2. Remove file transport folder
	err = os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "transport", "file"))
	if err != nil {
		return fmt.Errorf("removing file transport: %w", err)
	}

	// 3. Remove files migration
	err = os.Remove(filepath.Join(projectPath, "app", "service-core", "storage", "migrations", "00004_create_files.sql"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing files migration: %w", err)
	}

	// 4. Strip all GF_FILE marker blocks from all files
	err = StripIntegration(projectPath, "FILE")
	if err != nil {
		return fmt.Errorf("stripping file markers: %w", err)
	}

	return nil
}

// S3StripClient removes S3-related content from a generated client.
func S3StripClient(clientType, clientPath string) error {
	return StripClientIntegration(clientType, clientPath, "s3")
}

// S3AddClient adds S3-related content to an existing client.
func S3AddClient(tmpProject, clientType, clientPath string) error {
	return AddClientIntegration(tmpProject, clientType, clientPath, "s3")
}

// S3Add adds S3 file storage integration to an existing project.
// Called by 'gof add s3' command.
func S3Add(ctx context.Context, email, apiKey string) error {
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
	err = os.Chdir(tmpDir)
	if err != nil {
		return fmt.Errorf("changing to temp dir: %w", err)
	}

	err = repo.DownloadRepo(ctx, email, apiKey, "template")
	if err != nil {
		_ = os.Chdir(cwd)
		return fmt.Errorf("downloading template: %w", err)
	}

	// Return to original directory
	err = os.Chdir(cwd)
	if err != nil {
		return fmt.Errorf("returning to original dir: %w", err)
	}

	tmpProject := filepath.Join(tmpDir, "template")

	// 2. Copy file domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "file")
	dstDomain := filepath.Join("app", "service-core", "domain", "file")
	err = CopyDir(srcDomain, dstDomain)
	if err != nil {
		return fmt.Errorf("copying file domain: %w", err)
	}

	// 3. Copy file transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "file")
	dstTransport := filepath.Join("app", "service-core", "transport", "file")
	err = CopyDir(srcTransport, dstTransport)
	if err != nil {
		return fmt.Errorf("copying file transport: %w", err)
	}

	// 4. Copy and renumber files migration
	err = AddMigration(tmpProject, "00004_create_files.sql", "create_files.sql")
	if err != nil {
		return fmt.Errorf("adding files migration: %w", err)
	}

	// 5. Copy files with GF_FILE markers from template
	err = CopyFilesWithMarkers(tmpProject, ".", "FILE")
	if err != nil {
		return fmt.Errorf("copying files with FILE markers: %w", err)
	}

	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// 6. Add the local rustfs (S3-compatible) services to docker-compose.yml
	err = AppendComposeBlock(tmpProject, "FILE", cfg.ProjectName)
	if err != nil {
		return fmt.Errorf("adding rustfs to docker compose: %w", err)
	}

	enabledClients := clients.Enabled(cfg)
	for _, client := range enabledClients {
		clientPath := filepath.Join("app", client.ServiceDir)
		err := S3AddClient(tmpProject, client.Name, clientPath)
		if err != nil {
			return fmt.Errorf("adding files to %s client: %w", client.DisplayName, err)
		}
	}

	return nil
}
