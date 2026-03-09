package integrations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// PostmarkStrip removes all email/Postmark-related code from a freshly initialized project.
// Called by init command after downloading the template.
func PostmarkStrip(projectPath string) error {
	// 1. Remove email domain folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "domain", "email")); err != nil {
		return fmt.Errorf("removing email domain: %w", err)
	}

	// 2. Remove email transport folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "transport", "email")); err != nil {
		return fmt.Errorf("removing email transport: %w", err)
	}

	// 3. Remove emails migration
	if err := os.Remove(filepath.Join(projectPath, "app", "service-core", "storage", "migrations", "00005_create_emails.sql")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing emails migration: %w", err)
	}

	// 4. Strip all GF_EMAIL marker blocks from all files
	if err := StripIntegration(projectPath, "EMAIL"); err != nil {
		return fmt.Errorf("stripping email markers: %w", err)
	}

	return nil
}

// PostmarkStripClient removes Postmark-related content from a generated client.
func PostmarkStripClient(clientType, clientPath string) error {
	return StripClientIntegration(clientType, clientPath, "postmark")
}

// PostmarkAddClient adds Postmark-related content to an existing client.
func PostmarkAddClient(tmpProject, clientType, clientPath string) error {
	return AddClientIntegration(tmpProject, clientType, clientPath, "postmark")
}

// PostmarkStripE2E removes Postmark-related e2e tests.
// Called by 'gof client svelte' when postmark is not enabled.
func PostmarkStripE2E(e2ePath string) error {
	if err := os.Remove(filepath.Join(e2ePath, "emails.test.ts")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing emails.test.ts: %w", err)
	}
	return nil
}

// PostmarkAddE2E adds Postmark-related e2e tests.
// Called by 'gof add postmark' when client exists.
func PostmarkAddE2E(tmpProject, e2ePath string) error {
	src := filepath.Join(tmpProject, "e2e", "emails.test.ts")
	dst := filepath.Join(e2ePath, "emails.test.ts")
	if err := CopyFile(src, dst); err != nil {
		return fmt.Errorf("copying emails.test.ts: %w", err)
	}
	return nil
}

// PostmarkAdd adds Postmark email integration to an existing project.
// Called by 'gof add postmark' command.
func PostmarkAdd(email, apiKey string) error {
	// 1. Download template to temp location
	tmpDir, err := os.MkdirTemp("", "gofast-email-*")
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

	// 2. Copy email domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "email")
	dstDomain := filepath.Join("app", "service-core", "domain", "email")
	if err := CopyDir(srcDomain, dstDomain); err != nil {
		return fmt.Errorf("copying email domain: %w", err)
	}

	// 3. Copy email transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "email")
	dstTransport := filepath.Join("app", "service-core", "transport", "email")
	if err := CopyDir(srcTransport, dstTransport); err != nil {
		return fmt.Errorf("copying email transport: %w", err)
	}

	// 4. Copy and renumber emails migration
	if err := AddMigration(tmpProject, "00005_create_emails.sql", "create_emails.sql"); err != nil {
		return fmt.Errorf("adding emails migration: %w", err)
	}

	// 5. Copy files with GF_EMAIL markers from template
	if err := CopyFilesWithMarkers(tmpProject, ".", "EMAIL"); err != nil {
		return fmt.Errorf("copying files with EMAIL markers: %w", err)
	}

	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	enabledClients := clients.Enabled(cfg)
	for _, client := range enabledClients {
		clientPath := filepath.Join("app", client.ServiceDir)
		if err := PostmarkAddClient(tmpProject, client.Name, clientPath); err != nil {
			return fmt.Errorf("adding email to %s client: %w", client.DisplayName, err)
		}
	}

	if len(enabledClients) > 0 {
		e2ePath := "e2e"
		if _, err := os.Stat(e2ePath); err == nil {
			if err := PostmarkAddE2E(tmpProject, e2ePath); err != nil {
				return fmt.Errorf("adding postmark e2e tests: %w", err)
			}
		}
	}

	return nil
}
