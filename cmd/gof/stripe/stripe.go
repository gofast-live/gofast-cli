package stripe

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// Strip removes all stripe-related code from a freshly initialized project.
// Called by init command after downloading the template.
func Strip(projectPath string) error {
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
	if err := replaceCheckUserAccess(projectPath); err != nil {
		return fmt.Errorf("replacing CheckUserAccess: %w", err)
	}

	return nil
}

// replaceCheckUserAccess replaces the stripped CheckUserAccess with a simple fallback
func replaceCheckUserAccess(projectPath string) error {
	loginServicePath := filepath.Join(projectPath, "app", "service-core", "domain", "login", "service.go")

	content, err := os.ReadFile(loginServicePath)
	if err != nil {
		return err
	}

	// The simple fallback function to insert
	fallback := `func CheckUserAccess(_ context.Context, _ Deps, user query.User) (bool, int64, error) {
	return false, user.Access, nil
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

// StripIntegration removes all GF_<integration>_START/END blocks from all files in the project
func StripIntegration(projectPath string, integration string) error {
	startMarker := fmt.Sprintf("// GF_%s_START", integration)
	endMarker := fmt.Sprintf("// GF_%s_END", integration)
	sqlStartMarker := fmt.Sprintf("-- GF_%s_START", integration)
	sqlEndMarker := fmt.Sprintf("-- GF_%s_END", integration)

	return filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".proto" && ext != ".sql" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		s := string(content)
		original := s

		// Use appropriate markers based on file type
		if ext == ".sql" {
			s = removeMarkerBlocks(s, sqlStartMarker, sqlEndMarker)
		} else {
			s = removeMarkerBlocks(s, startMarker, endMarker)
		}

		// Only write if changed
		if s != original {
			if err := os.WriteFile(path, []byte(s), 0644); err != nil {
				return err
			}
		}

		return nil
	})
}

// removeMarkerBlocks removes all blocks between startMarker and endMarker (inclusive of markers)
func removeMarkerBlocks(content, startMarker, endMarker string) string {
	for {
		startIdx := strings.Index(content, startMarker)
		if startIdx == -1 {
			break
		}

		// Find start of line containing start marker
		lineStart := strings.LastIndex(content[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++ // Move past the newline
		}

		// Find end marker
		endIdx := strings.Index(content[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx = startIdx + endIdx + len(endMarker)

		// Skip to end of line
		if endIdx < len(content) && content[endIdx] == '\n' {
			endIdx++
		}

		content = content[:lineStart] + content[endIdx:]
	}
	return content
}

// Add adds Stripe payment integration to an existing project.
// Called by 'gof add stripe' command.
func Add(email, apiKey string) error {
	// 1. Download template to temp location
	tmpDir, err := os.MkdirTemp("", "gofast-stripe-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpProject := filepath.Join(tmpDir, "template")
	if err := repo.DownloadRepo(email, apiKey, tmpProject); err != nil {
		return fmt.Errorf("downloading template: %w", err)
	}

	// 2. Copy payment domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "payment")
	dstDomain := filepath.Join("app", "service-core", "domain", "payment")
	if err := copyDir(srcDomain, dstDomain); err != nil {
		return fmt.Errorf("copying payment domain: %w", err)
	}

	// 3. Copy payment transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "payment")
	dstTransport := filepath.Join("app", "service-core", "transport", "payment")
	if err := copyDir(srcTransport, dstTransport); err != nil {
		return fmt.Errorf("copying payment transport: %w", err)
	}

	// 4. Copy and renumber subscriptions migration
	if err := addSubscriptionsMigration(tmpProject); err != nil {
		return fmt.Errorf("adding subscriptions migration: %w", err)
	}

	// 5. Copy files with GF_STRIPE markers from template, keeping stripe markers intact
	if err := copyFilesWithMarkers(tmpProject, ".", "STRIPE"); err != nil {
		return fmt.Errorf("copying files with stripe markers: %w", err)
	}

	return nil
}

// copyFilesWithMarkers copies files that have GF_<integration> markers from src to dst
// It preserves the specified integration's markers while stripping others
func copyFilesWithMarkers(srcProject, dstProject, keepIntegration string) error {
	// Walk service-core directory
	srcServiceCore := filepath.Join(srcProject, "app", "service-core")
	if err := copyMarkedFiles(srcServiceCore, filepath.Join(dstProject, "app", "service-core"), keepIntegration); err != nil {
		return err
	}

	// Walk proto directory
	srcProto := filepath.Join(srcProject, "proto")
	if err := copyMarkedFiles(srcProto, filepath.Join(dstProject, "proto"), keepIntegration); err != nil {
		return err
	}

	return nil
}

// copyMarkedFiles walks srcDir and copies files with markers to dstDir
func copyMarkedFiles(srcDir, dstDir, keepIntegration string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(srcPath)
		if ext != ".go" && ext != ".proto" && ext != ".sql" {
			return nil
		}

		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		marker := fmt.Sprintf("GF_%s_", keepIntegration)
		if !strings.Contains(string(content), marker) {
			return nil // Skip files without our integration markers
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		// Strip other integrations' markers (not the one we're adding)
		s := string(content)
		s = stripOtherIntegrations(s, keepIntegration)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(dstPath, []byte(s), 0644)
	})
}

// stripOtherIntegrations removes marker blocks for all integrations except the specified one
func stripOtherIntegrations(content, keepIntegration string) string {
	// Find all integration markers in the content
	re := regexp.MustCompile(`// GF_([A-Z]+)_START`)
	matches := re.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			seen[match[1]] = true
		}
	}

	// Also check for SQL-style markers
	reSql := regexp.MustCompile(`-- GF_([A-Z]+)_START`)
	matchesSql := reSql.FindAllStringSubmatch(content, -1)
	for _, match := range matchesSql {
		if len(match) > 1 {
			seen[match[1]] = true
		}
	}

	// Strip all integrations except the one we're keeping
	for integration := range seen {
		if integration != keepIntegration {
			content = removeMarkerBlocks(content, fmt.Sprintf("// GF_%s_START", integration), fmt.Sprintf("// GF_%s_END", integration))
			content = removeMarkerBlocks(content, fmt.Sprintf("-- GF_%s_START", integration), fmt.Sprintf("-- GF_%s_END", integration))
		}
	}

	return content
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, info.Mode())
	})
}

// addSubscriptionsMigration copies the subscriptions migration with the next available number
func addSubscriptionsMigration(tmpProject string) error {
	// Find next migration number
	migrationsDir := filepath.Join("app", "service-core", "storage", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var maxNum int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) >= 1 {
				num, _ := strconv.Atoi(parts[0])
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}
	nextNum := maxNum + 1

	// Read source migration
	srcPath := filepath.Join(tmpProject, "app", "service-core", "storage", "migrations", "00003_create_subscriptions.sql")
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Strip the GF_STRIPE markers from migration content (we want the actual SQL)
	s := removeMarkerBlocks(string(content), "-- GF_STRIPE_START", "-- GF_STRIPE_END")
	// But we need to keep the content between markers, not remove it
	// Actually, the migration file has markers around the whole content
	// So we should just copy as-is since it's a new file
	s = string(content)

	// Write with new number
	dstName := fmt.Sprintf("%05d_create_subscriptions.sql", nextNum)
	dstPath := filepath.Join(migrationsDir, dstName)
	return os.WriteFile(dstPath, []byte(s), 0644)
}

// GetNextMigrationNumber returns the next available migration number
func GetNextMigrationNumber() (int, error) {
	migrationsDir := filepath.Join("app", "service-core", "storage", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, err
	}

	var numbers []int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) >= 1 {
				num, _ := strconv.Atoi(parts[0])
				numbers = append(numbers, num)
			}
		}
	}

	if len(numbers) == 0 {
		return 1, nil
	}

	sort.Ints(numbers)
	return numbers[len(numbers)-1] + 1, nil
}
