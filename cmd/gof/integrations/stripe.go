package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// StripeStrip removes all stripe-related code from a freshly initialized project.
// Called by init command after downloading the template.
func StripeStrip(projectPath string) error {
	// 1. Remove payment domain folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "domain", "payment")); err != nil {
		return fmt.Errorf("removing payment domain: %w", err)
	}

	// 2. Remove payment transport folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "transport", "payment")); err != nil {
		return fmt.Errorf("removing payment transport: %w", err)
	}

	// 3. Remove subscriptions migration
	if err := os.Remove(filepath.Join(projectPath, "app", "service-core", "storage", "migrations", "00003_create_subscriptions.sql")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing subscriptions migration: %w", err)
	}

	// 4. Strip all GF_STRIPE marker blocks from all files
	if err := StripIntegration(projectPath, "STRIPE"); err != nil {
		return fmt.Errorf("stripping stripe markers: %w", err)
	}

	// 5. Replace CheckUserAccess with simple fallback (the Stripe version was stripped)
	if err := stripeReplaceCheckUserAccess(projectPath); err != nil {
		return fmt.Errorf("replacing CheckUserAccess: %w", err)
	}

	return nil
}

// stripeReplaceCheckUserAccess replaces the stripped CheckUserAccess with a simple fallback
func stripeReplaceCheckUserAccess(projectPath string) error {
	loginServicePath := filepath.Join(projectPath, "app", "service-core", "domain", "login", "service.go")

	content, err := os.ReadFile(loginServicePath)
	if err != nil {
		return err
	}

	// The simple fallback function to insert
	fallback := `func CheckUserAccess(_ context.Context, _ *Deps, user query.User) (int64, error) {
	return user.Access, nil
}

`

	// Find where to insert - right before ForceRefresh function
	s := string(content)
	insertPoint := strings.Index(s, "func ForceRefresh(")
	if insertPoint == -1 {
		return fmt.Errorf("could not find ForceRefresh function")
	}

	// Insert the fallback
	s = s[:insertPoint] + fallback + s[insertPoint:]

	return os.WriteFile(loginServicePath, []byte(s), 0644)
}

// StripeStripClient removes Stripe-related content from a generated client.
func StripeStripClient(clientType, clientPath string) error {
	return StripClientIntegration(clientType, clientPath, "stripe")
}

// StripeAddClient adds Stripe-related content to an existing client.
func StripeAddClient(tmpProject, clientType, clientPath string) error {
	return AddClientIntegration(tmpProject, clientType, clientPath, "stripe")
}

// StripeStripE2E removes Stripe-related e2e tests.
// Called by 'gof client svelte' when stripe is not enabled.
func StripeStripE2E(e2ePath string) error {
	if err := os.Remove(filepath.Join(e2ePath, "payments.test.ts")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing payments.test.ts: %w", err)
	}
	return nil
}

// StripeAddE2E adds Stripe-related e2e tests.
// Called by 'gof add stripe' when client exists.
func StripeAddE2E(tmpProject, e2ePath string) error {
	src := filepath.Join(tmpProject, "e2e", "payments.test.ts")
	dst := filepath.Join(e2ePath, "payments.test.ts")
	if err := CopyFile(src, dst); err != nil {
		return fmt.Errorf("copying payments.test.ts: %w", err)
	}
	return nil
}

// StripeAdd adds Stripe payment integration to an existing project.
// Called by 'gof add stripe' command.
func StripeAdd(email, apiKey string) error {
	// 1. Download template to temp location
	tmpDir, err := os.MkdirTemp("", "gofast-stripe-*")
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

	// 2. Copy payment domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "payment")
	dstDomain := filepath.Join("app", "service-core", "domain", "payment")
	if err := CopyDir(srcDomain, dstDomain); err != nil {
		return fmt.Errorf("copying payment domain: %w", err)
	}

	// 3. Copy payment transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "payment")
	dstTransport := filepath.Join("app", "service-core", "transport", "payment")
	if err := CopyDir(srcTransport, dstTransport); err != nil {
		return fmt.Errorf("copying payment transport: %w", err)
	}

	// 4. Copy and renumber subscriptions migration
	if err := AddMigration(tmpProject, "00003_create_subscriptions.sql", "create_subscriptions.sql"); err != nil {
		return fmt.Errorf("adding subscriptions migration: %w", err)
	}

	// 5. Copy files with GF_STRIPE markers from template, keeping stripe markers intact
	if err := CopyFilesWithMarkers(tmpProject, ".", "STRIPE"); err != nil {
		return fmt.Errorf("copying files with stripe markers: %w", err)
	}

	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	enabledClients := clients.Enabled(cfg)
	for _, client := range enabledClients {
		clientPath := filepath.Join("app", client.ServiceDir)
		if err := StripeAddClient(tmpProject, client.Name, clientPath); err != nil {
			return fmt.Errorf("adding stripe to %s client: %w", client.DisplayName, err)
		}
	}

	if len(enabledClients) > 0 {
		e2ePath := "e2e"
		if _, err := os.Stat(e2ePath); err == nil {
			if err := StripeAddE2E(tmpProject, e2ePath); err != nil {
				return fmt.Errorf("adding stripe e2e tests: %w", err)
			}
		}
	}

	return nil
}
