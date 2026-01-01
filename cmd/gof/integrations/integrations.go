package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

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
		if ext != ".go" && ext != ".sql" {
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
			s = RemoveMarkerBlocks(s, sqlStartMarker, sqlEndMarker)
		} else {
			s = RemoveMarkerBlocks(s, startMarker, endMarker)
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

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
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

// CopyFilesWithMarkers copies files that have GF_<integration> markers from src to dst
// It preserves the specified integration's markers while stripping others
func CopyFilesWithMarkers(srcProject, dstProject, keepIntegration string) error {
	srcServiceCore := filepath.Join(srcProject, "app", "service-core")
	return copyMarkedFiles(srcServiceCore, filepath.Join(dstProject, "app", "service-core"), keepIntegration)
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

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
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
}

// AppendMarkerBlock extracts the marker block from src and appends it to dst
func AppendMarkerBlock(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
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
		return err
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

	return os.WriteFile(dstPath, []byte(result), 0644)
}

// MergeMainGoMarkers extracts marker blocks from src main.go and injects them into dst main.go
// Import blocks are injected before GF_MAIN_IMPORT_SERVICES_START
// Init blocks are injected before GF_MAIN_INIT_SERVICES_START
func MergeMainGoMarkers(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return err
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
		importMarker := "// GF_MAIN_IMPORT_SERVICES_START"
		idx := strings.Index(dst, importMarker)
		if idx != -1 {
			// Find start of line
			lineStart := strings.LastIndex(dst[:idx], "\n")
			if lineStart == -1 {
				lineStart = 0
			} else {
				lineStart++
			}
			insertContent := strings.Join(importBlocks, "")
			dst = dst[:lineStart] + insertContent + dst[lineStart:]
		}
	}

	// Inject init blocks before GF_MAIN_INIT_SERVICES_START
	if len(initBlocks) > 0 {
		initMarker := "// GF_MAIN_INIT_SERVICES_START"
		idx := strings.Index(dst, initMarker)
		if idx != -1 {
			// Find start of line
			lineStart := strings.LastIndex(dst[:idx], "\n")
			if lineStart == -1 {
				lineStart = 0
			} else {
				lineStart++
			}
			// Add newline before if needed
			insertContent := strings.Join(initBlocks, "")
			if !strings.HasPrefix(insertContent, "\n") {
				insertContent = "\n" + insertContent
			}
			dst = dst[:lineStart] + insertContent + dst[lineStart:]
		}
	}

	return os.WriteFile(dstPath, []byte(dst), 0644)
}

// MergeConfigMarkers extracts marker blocks from src config.go and injects them into dst config.go
// Blocks are inserted at GF_CONFIG_STRUCT_INSERT and GF_CONFIG_INIT_INSERT markers
func MergeConfigMarkers(srcPath, dstPath, integration string) error {
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return err
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

	return os.WriteFile(dstPath, []byte(dst), 0644)
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
		return err
	}

	dstName := fmt.Sprintf("%05d_%s", nextNum, dstMigrationSuffix)
	dstPath := filepath.Join("app", "service-core", "storage", "migrations", dstName)
	return os.WriteFile(dstPath, content, 0644)
}
