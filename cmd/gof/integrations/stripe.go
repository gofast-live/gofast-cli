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

// StripeStripClient removes Stripe-related content from the Svelte client.
// Called by 'gof client svelte' command after copying the client folder.
func StripeStripClient(clientPath string) error {
	// 1. Remove payments route folder
	paymentsPath := filepath.Join(clientPath, "src", "routes", "(app)", "payments")
	if err := os.RemoveAll(paymentsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing payments folder: %w", err)
	}

	// 2. Strip Payments nav entry and Coins import from layout
	layoutPath := filepath.Join(clientPath, "src", "routes", "(app)", "+layout.svelte")
	if err := RemoveNavEntry(layoutPath, "/payments", "Coins"); err != nil {
		return fmt.Errorf("removing payments nav entry: %w", err)
	}

	// 3. Remove Stripe success handling
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("reading layout: %w", err)
	}

	s := string(content)

	// Remove Stripe success handling (the force refresh check)
	s = regexp.MustCompile(`(?s)\s*const force = page\.url\.searchParams\.get\("success"\) === "true";\s*if \(force\) \{\s*// Wait for Stripe webhook to process\s*await new Promise\(\(r\) => setTimeout\(r, 2000\)\);\s*\}`).ReplaceAllString(s, "")

	// Replace { force } with {} in refresh call
	s = strings.Replace(s, "const response = await login_client.refresh({ force });", "const response = await login_client.refresh({});", 1)

	if err := os.WriteFile(layoutPath, []byte(s), 0644); err != nil {
		return fmt.Errorf("writing layout: %w", err)
	}

	return nil
}

// StripeAddClient adds Stripe-related content to an existing Svelte client.
// Called by 'gof add stripe' when client already exists.
func StripeAddClient(tmpProject, clientPath string) error {
	// 1. Copy payments route folder
	srcPayments := filepath.Join(tmpProject, "app", "service-client", "src", "routes", "(app)", "payments")
	dstPayments := filepath.Join(clientPath, "src", "routes", "(app)", "payments")
	if err := CopyDir(srcPayments, dstPayments); err != nil {
		return fmt.Errorf("copying payments folder: %w", err)
	}

	// 2. Add Payments nav entry and Coins import to layout
	layoutPath := filepath.Join(clientPath, "src", "routes", "(app)", "+layout.svelte")
	if err := AddNavEntry(layoutPath, "Payments", "/payments", "Coins"); err != nil {
		return fmt.Errorf("adding payments nav entry: %w", err)
	}

	// 3. Add Stripe success handling
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("reading layout: %w", err)
	}

	s := string(content)

	if !strings.Contains(s, `page.url.searchParams.get("success")`) {
		s = strings.Replace(s,
			`const response = await login_client.refresh({});`,
			`const force = page.url.searchParams.get("success") === "true";
            if (force) {
                // Wait for Stripe webhook to process
                await new Promise((r) => setTimeout(r, 2000));
            }
            const response = await login_client.refresh({ force });`,
			1)
	}

	if err := os.WriteFile(layoutPath, []byte(s), 0644); err != nil {
		return fmt.Errorf("writing layout: %w", err)
	}

	return nil
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

	// 6. Add client-side Stripe content if Svelte client is configured
	if config.IsSvelte() {
		clientPath := filepath.Join("app", "service-client")
		if err := StripeAddClient(tmpProject, clientPath); err != nil {
			return fmt.Errorf("adding stripe to client: %w", err)
		}
		// Add e2e tests if e2e folder exists
		e2ePath := "e2e"
		if _, err := os.Stat(e2ePath); err == nil {
			if err := StripeAddE2E(tmpProject, e2ePath); err != nil {
				return fmt.Errorf("adding stripe e2e tests: %w", err)
			}
		}
	}

	return nil
}
