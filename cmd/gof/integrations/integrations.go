package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
)

// ComposeFileName is the base compose file; it carries "# GF_<integration>" blocks for integration services.
const ComposeFileName = "docker-compose.yml"

// StripIntegration removes all GF_<integration>_START/END blocks from all files in the project
func StripIntegration(projectPath string, integration string) error {
	startMarker := fmt.Sprintf("// GF_%s_START", integration)
	endMarker := fmt.Sprintf("// GF_%s_END", integration)
	sqlStartMarker := fmt.Sprintf("-- GF_%s_START", integration)
	sqlEndMarker := fmt.Sprintf("-- GF_%s_END", integration)
	composeStartMarker := fmt.Sprintf("# GF_%s_START", integration)
	composeEndMarker := fmt.Sprintf("# GF_%s_END", integration)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		isCompose := filepath.Base(path) == ComposeFileName
		if ext != ".go" && ext != ".sql" && !isCompose {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		s := string(content)
		original := s

		// Use appropriate markers based on file type
		switch {
		case isCompose:
			s = RemoveMarkerBlocks(s, composeStartMarker, composeEndMarker)
		case ext == ".sql":
			s = RemoveMarkerBlocks(s, sqlStartMarker, sqlEndMarker)
		default:
			s = RemoveMarkerBlocks(s, startMarker, endMarker)
		}

		// Only write if changed
		if s != original {
			err := os.WriteFile(path, []byte(s), 0644)
			if err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking %s: %w", projectPath, err)
	}
	return nil
}

// RemoveMarkerBlocks removes all blocks between startMarker and endMarker (inclusive of markers)
func RemoveMarkerBlocks(content, startMarker, endMarker string) string {
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

// CopyDir copies a directory recursively
func CopyDir(src, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("resolving %s relative to %s: %w", path, src, err)
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		return os.WriteFile(dstPath, content, info.Mode())
	})
	if err != nil {
		return fmt.Errorf("walking %s: %w", src, err)
	}
	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}
	// Ensure destination directory exists
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("creating directory for %s: %w", dst, err)
	}
	err = os.WriteFile(dst, content, 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dst, err)
	}
	return nil
}

// GetNextMigrationNumber returns the next available migration number
func GetNextMigrationNumber() (int, error) {
	migrationsDir := filepath.Join("app", "service-core", "storage", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", migrationsDir, err)
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

// CopyFilesWithMarkers copies files that have GF_<integration> markers from src to dst
// It preserves the specified integration's markers while stripping others
func CopyFilesWithMarkers(srcProject, dstProject, keepIntegration string) error {
	srcServiceCore := filepath.Join(srcProject, "app", "service-core")
	return copyMarkedFiles(srcServiceCore, filepath.Join(dstProject, "app", "service-core"), keepIntegration)
}

// copyMarkedFiles walks srcDir and copies files with markers to dstDir
func copyMarkedFiles(srcDir, dstDir, keepIntegration string) error {
	err := filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Skip migrations directory - migrations are handled separately with proper numbering
		if strings.Contains(srcPath, "migrations") {
			return nil
		}

		ext := filepath.Ext(srcPath)
		if ext != ".go" && ext != ".sql" {
			return nil
		}

		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", srcPath, err)
		}

		marker := fmt.Sprintf("GF_%s_", keepIntegration)
		if !strings.Contains(string(content), marker) {
			return nil // Skip files without our integration markers
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return fmt.Errorf("resolving %s relative to %s: %w", srcPath, srcDir, err)
		}
		dstPath := filepath.Join(dstDir, relPath)

		// Ensure directory exists
		err = os.MkdirAll(filepath.Dir(dstPath), 0755)
		if err != nil {
			return fmt.Errorf("creating directory for %s: %w", dstPath, err)
		}

		// For query.sql, append only the marker block instead of overwriting
		if filepath.Base(dstPath) == "query.sql" {
			return AppendMarkerBlock(srcPath, dstPath, keepIntegration)
		}

		// For main.go, merge marker blocks instead of overwriting
		if filepath.Base(dstPath) == "main.go" {
			return MergeMainGoMarkers(srcPath, dstPath, keepIntegration)
		}

		// For config.go, merge marker blocks instead of overwriting
		if filepath.Base(dstPath) == "config.go" {
			return MergeConfigMarkers(srcPath, dstPath, keepIntegration)
		}

		// Strip other integrations' markers (not the one we're adding)
		s := string(content)
		s = StripOtherIntegrations(s, keepIntegration)

		return os.WriteFile(dstPath, []byte(s), 0644)
	})
	if err != nil {
		return fmt.Errorf("walking %s: %w", srcDir, err)
	}
	return nil
}

// AppendMarkerBlock extracts the marker block from src and appends it to dst
func AppendMarkerBlock(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcPath, err)
	}

	// Determine marker style based on file extension
	var startMarker, endMarker string
	if filepath.Ext(srcPath) == ".sql" {
		startMarker = fmt.Sprintf("-- GF_%s_START", integration)
		endMarker = fmt.Sprintf("-- GF_%s_END", integration)
	} else {
		startMarker = fmt.Sprintf("// GF_%s_START", integration)
		endMarker = fmt.Sprintf("// GF_%s_END", integration)
	}

	s := string(srcContent)
	startIdx := strings.Index(s, startMarker)
	if startIdx == -1 {
		return nil // No marker block to append
	}

	// Find start of line containing start marker
	lineStart := strings.LastIndex(s[:startIdx], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++ // Move past the newline
	}

	// Find end marker
	endIdx := strings.Index(s[startIdx:], endMarker)
	if endIdx == -1 {
		return nil // Malformed markers
	}
	endIdx = startIdx + endIdx + len(endMarker)

	// Include the newline after end marker if present
	if endIdx < len(s) && s[endIdx] == '\n' {
		endIdx++
	}

	markerBlock := s[lineStart:endIdx]

	// Read existing destination file
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dstPath, err)
	}

	// Check if marker block already exists
	if strings.Contains(string(dstContent), startMarker) {
		return nil // Already has this integration
	}

	// Append the marker block
	result := string(dstContent)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	result += markerBlock

	err = os.WriteFile(dstPath, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dstPath, err)
	}
	return nil
}

// AppendComposeBlock copies the "# GF_<integration>" services block from the template's
// docker-compose.yml to the end of the project's, renaming "gofast" to the project name the
// same way init does. Only the copied block is renamed, so the rest of the file is untouched.
func AppendComposeBlock(tmpProject, integration, projectName string) error {
	startMarker := fmt.Sprintf("# GF_%s_START", integration)
	endMarker := fmt.Sprintf("# GF_%s_END", integration)

	srcContent, err := os.ReadFile(filepath.Join(tmpProject, ComposeFileName))
	if err != nil {
		return fmt.Errorf("reading template compose file: %w", err)
	}
	src := string(srcContent)

	startIdx := strings.Index(src, startMarker)
	if startIdx == -1 {
		return fmt.Errorf("template %s has no %s block", ComposeFileName, startMarker)
	}
	endIdx := strings.Index(src[startIdx:], endMarker)
	if endIdx == -1 {
		return fmt.Errorf("template %s has no %s marker", ComposeFileName, endMarker)
	}
	// Block runs from the start of the start-marker line through the end-marker line
	lineStart := strings.LastIndex(src[:startIdx], "\n") + 1
	blockEnd := startIdx + endIdx + len(endMarker)
	block := strings.ReplaceAll(src[lineStart:blockEnd], "gofast", projectName) + "\n"

	dstContent, err := os.ReadFile(ComposeFileName)
	if err != nil {
		return fmt.Errorf("reading project compose file: %w", err)
	}
	dst := string(dstContent)
	if strings.Contains(dst, startMarker) {
		return nil // Already has this integration
	}

	// Services are the last top-level key in the compose file, so the block lands inside it
	dst = strings.TrimRight(dst, "\n") + "\n\n" + block
	err = os.WriteFile(ComposeFileName, []byte(dst), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", ComposeFileName, err)
	}
	return nil
}

func StripClientIntegration(clientType, clientPath, integration string) error {
	spec, ok := clients.SpecFor(clientType)
	if !ok {
		return fmt.Errorf("unknown client type %q", clientType)
	}

	routeSubpath, err := integrationRouteSubpath(spec, integration)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(clientPath, filepath.FromSlash(routeSubpath))
	err = os.RemoveAll(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %s route %s: %w", integration, targetPath, err)
	}
	return nil
}

func AddClientIntegration(tmpProject, clientType, clientPath, integration string) error {
	spec, ok := clients.SpecFor(clientType)
	if !ok {
		return fmt.Errorf("unknown client type %q", clientType)
	}

	routeSubpath, err := integrationRouteSubpath(spec, integration)
	if err != nil {
		return err
	}

	srcPath := filepath.Join(tmpProject, "app", spec.ServiceDir, filepath.FromSlash(routeSubpath))
	dstPath := filepath.Join(clientPath, filepath.FromSlash(routeSubpath))

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat client integration source %s: %w", srcPath, err)
	}
	if info.IsDir() {
		err := CopyDir(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("copying client integration directory %s: %w", srcPath, err)
		}
		return nil
	}
	err = CopyFile(srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("copying client integration file %s: %w", srcPath, err)
	}
	return nil
}

func integrationRouteSubpath(spec clients.Spec, integration string) (string, error) {
	switch integration {
	case "stripe":
		return spec.PaymentsRouteSubpath, nil
	case "s3":
		return spec.FilesRouteSubpath, nil
	case "postmark":
		return spec.EmailsRouteSubpath, nil
	default:
		return "", fmt.Errorf("unknown integration %q", integration)
	}
}

// MergeMainGoMarkers extracts marker blocks from src main.go and injects them into dst main.go
// Import blocks are injected before GF_MAIN_IMPORT_SERVICES_START
// Init blocks are injected before GF_MAIN_INIT_SERVICES_START
func MergeMainGoMarkers(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcPath, err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dstPath, err)
	}

	startMarker := fmt.Sprintf("// GF_%s_START", integration)
	endMarker := fmt.Sprintf("// GF_%s_END", integration)

	// Check if already has this integration
	if strings.Contains(string(dstContent), startMarker) {
		return nil // Already has this integration
	}

	src := string(srcContent)
	dst := string(dstContent)

	// Extract all marker blocks from source
	var importBlocks, initBlocks []string
	remaining := src
	for {
		startIdx := strings.Index(remaining, startMarker)
		if startIdx == -1 {
			break
		}

		// Find start of line
		lineStart := strings.LastIndex(remaining[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++
		}

		// Find end marker
		endIdx := strings.Index(remaining[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx = startIdx + endIdx + len(endMarker)

		// Include newline after end marker
		if endIdx < len(remaining) && remaining[endIdx] == '\n' {
			endIdx++
		}

		block := remaining[lineStart:endIdx]

		// Determine if this is an import block (contains import paths like "gofast/")
		if strings.Contains(block, "\"gofast/") || strings.Contains(block, "Svc \"") || strings.Contains(block, "Route \"") {
			importBlocks = append(importBlocks, block)
		} else {
			initBlocks = append(initBlocks, block)
		}

		remaining = remaining[endIdx:]
	}

	// Inject import blocks before GF_MAIN_IMPORT_SERVICES_START
	if len(importBlocks) > 0 {
		dst = insertBeforeMarkerLine(dst, "// GF_MAIN_IMPORT_SERVICES_START", strings.Join(importBlocks, ""))
	}

	// Inject init blocks before GF_MAIN_INIT_SERVICES_START, on their own line
	if len(initBlocks) > 0 {
		insertContent := strings.Join(initBlocks, "")
		if !strings.HasPrefix(insertContent, "\n") {
			insertContent = "\n" + insertContent
		}
		dst = insertBeforeMarkerLine(dst, "// GF_MAIN_INIT_SERVICES_START", insertContent)
	}

	err = os.WriteFile(dstPath, []byte(dst), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dstPath, err)
	}
	return nil
}

// MergeConfigMarkers extracts marker blocks from src config.go and injects them into dst config.go
// Blocks are inserted at GF_CONFIG_STRUCT_INSERT and GF_CONFIG_INIT_INSERT markers
func MergeConfigMarkers(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcPath, err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dstPath, err)
	}

	startMarker := fmt.Sprintf("// GF_%s_START", integration)
	endMarker := fmt.Sprintf("// GF_%s_END", integration)

	// Check if already has this integration
	if strings.Contains(string(dstContent), startMarker) {
		return nil // Already has this integration
	}

	src := string(srcContent)
	dst := string(dstContent)

	// Extract all marker blocks from source, categorized by type
	var structBlocks, initBlocks []string
	remaining := src
	for {
		startIdx := strings.Index(remaining, startMarker)
		if startIdx == -1 {
			break
		}

		// Find start of line
		lineStart := strings.LastIndex(remaining[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++
		}

		// Find end marker
		endIdx := strings.Index(remaining[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx = startIdx + endIdx + len(endMarker)

		// Include newline after end marker
		if endIdx < len(remaining) && remaining[endIdx] == '\n' {
			endIdx++
		}

		block := remaining[lineStart:endIdx]

		// Categorize: blocks with MustSetEnv are initialization, others are struct fields
		if strings.Contains(block, "MustSetEnv") {
			initBlocks = append(initBlocks, block)
		} else {
			structBlocks = append(structBlocks, block)
		}

		remaining = remaining[endIdx:]
	}

	// Insert struct field blocks after GF_CONFIG_STRUCT_INSERT marker
	structInsertMarker := "// GF_CONFIG_STRUCT_INSERT"
	for _, block := range structBlocks {
		idx := strings.Index(dst, structInsertMarker)
		if idx != -1 {
			// Find end of the marker line
			lineEnd := strings.Index(dst[idx:], "\n")
			if lineEnd != -1 {
				insertPoint := idx + lineEnd + 1
				dst = dst[:insertPoint] + "\n" + block + dst[insertPoint:]
			}
		}
	}

	// Insert initialization blocks after GF_CONFIG_INIT_INSERT marker
	initInsertMarker := "// GF_CONFIG_INIT_INSERT"
	for _, block := range initBlocks {
		idx := strings.Index(dst, initInsertMarker)
		if idx != -1 {
			// Find end of the marker line
			lineEnd := strings.Index(dst[idx:], "\n")
			if lineEnd != -1 {
				insertPoint := idx + lineEnd + 1
				dst = dst[:insertPoint] + block + dst[insertPoint:]
			}
		}
	}

	err = os.WriteFile(dstPath, []byte(dst), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dstPath, err)
	}
	return nil
}

// StripOtherIntegrations removes marker blocks for all integrations except the specified one
func StripOtherIntegrations(content, keepIntegration string) string {
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
	reSQL := regexp.MustCompile(`-- GF_([A-Z]+)_START`)
	matchesSQL := reSQL.FindAllStringSubmatch(content, -1)
	for _, match := range matchesSQL {
		if len(match) > 1 {
			seen[match[1]] = true
		}
	}

	// Strip all integrations except the one we're keeping
	for integration := range seen {
		if integration != keepIntegration {
			content = RemoveMarkerBlocks(content, fmt.Sprintf("// GF_%s_START", integration), fmt.Sprintf("// GF_%s_END", integration))
			content = RemoveMarkerBlocks(content, fmt.Sprintf("-- GF_%s_START", integration), fmt.Sprintf("-- GF_%s_END", integration))
		}
	}

	return content
}

// AddMigration copies a migration file with the next available number
func AddMigration(tmpProject, srcMigrationName, dstMigrationSuffix string) error {
	nextNum, err := GetNextMigrationNumber()
	if err != nil {
		return err
	}

	srcPath := filepath.Join(tmpProject, "app", "service-core", "storage", "migrations", srcMigrationName)
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcPath, err)
	}

	dstName := fmt.Sprintf("%05d_%s", nextNum, dstMigrationSuffix)
	dstPath := filepath.Join("app", "service-core", "storage", "migrations", dstName)
	err = os.WriteFile(dstPath, content, 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dstPath, err)
	}
	return nil
}

// insertBeforeMarkerLine inserts text at the start of the line holding marker.
// Content without the marker comes back unchanged.
func insertBeforeMarkerLine(content, marker, insert string) string {
	before, _, found := strings.Cut(content, marker)
	if !found {
		return content
	}
	// LastIndex returns -1 when the marker is on the first line, so +1 lands on 0 either way
	lineStart := strings.LastIndex(before, "\n") + 1
	return content[:lineStart] + insert + content[lineStart:]
}
