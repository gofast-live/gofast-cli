package svelte

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateClientListPageSpec scaffolds a client list page test by copying the
// skeleton spec and performing token and marker-based replacements for
// singular/plural model variants and column-aware assertions.
func GenerateClientListPageSpec(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/page.svelte.spec.ts"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists
	destDir := filepath.Join("app/service-client/src/routes/(app)/models", pluralLower)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "page.svelte.spec.ts")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	// Token replacements (plural/title before singular to avoid partial stomps)
	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Helper: Title-case a label from snake_case
	toTitle := func(name string) string {
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, " ")
	}

	// Build mock data creation function
	var mockB strings.Builder
	mockB.WriteString(fmt.Sprintf("function createMock%s(overrides: Partial<%s> = {}): %s {\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
	mockB.WriteString("    return {\n")
	mockB.WriteString(fmt.Sprintf("        $typeName: \"proto.v1.%s\" as const,\n", capitalizedModelName))
	mockB.WriteString("        id: \"123\",\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			mockB.WriteString(fmt.Sprintf("        %s: \"Test %s %d\",\n", c.Name, capitalizedModelName, i))
		case "number":
			mockB.WriteString(fmt.Sprintf("        %s: \"%d\",\n", c.Name, 25+i))
		case "time":
			mockB.WriteString(fmt.Sprintf("        %s: \"2023-01-%02dT00:00:00Z\",\n", c.Name, 15+i))
		case "bool":
			mockB.WriteString(fmt.Sprintf("        %s: true,\n", c.Name))
		}
	}
	mockB.WriteString("        created: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        updated: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        ...overrides,\n")
	mockB.WriteString("    };\n")
	mockB.WriteString("}\n")
	mockFields := mockB.String()

	// Build column header assertions
	var headersB strings.Builder
	for _, c := range columns {
		title := toTitle(c.Name)
		headersB.WriteString("        const " + c.Name + "Header = page.getByRole(\"columnheader\", { name: \"" + title + "\" });\n")
		headersB.WriteString("        await expect.element(" + c.Name + "Header).toBeInTheDocument();\n")
	}
	headersAssert := strings.TrimRight(headersB.String(), "\n")

	// Build row assertion for first test case
	var rowAssertB strings.Builder
	rowIndent := "            "
	boolIndex := 0
	for i, c := range columns {
		switch c.Type {
		case "string":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"Test %s %d\"))\n", capitalizedModelName, i))
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "number":
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"%d\")).toBeInTheDocument();\n", 25+i))
		case "time":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"1/%d/2023\"))\n", 15+i))
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "bool":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByRole(\"cell\", { name: \"Yes\", exact: true }).nth(%d))\n", boolIndex))
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
			boolIndex++
		}
	}
	rowAssert := strings.TrimRight(rowAssertB.String(), "\n")

	// Build row assertion for second test case (different values)
	var rowAssertB2 strings.Builder
	boolIndex2 := 0
	for i, c := range columns {
		switch c.Type {
		case "string":
			rowAssertB2.WriteString(rowIndent + "await expect\n")
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"Another %s %d\"))\n", capitalizedModelName, i))
			rowAssertB2.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "number":
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"%d\")).toBeInTheDocument();\n", 30+i))
		case "time":
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"6/%d/2024\")).toBeInTheDocument();\n", 1+i))
		case "bool":
			rowAssertB2.WriteString(rowIndent + "await expect\n")
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByRole(\"cell\", { name: \"No\", exact: true }).nth(%d))\n", boolIndex2))
			rowAssertB2.WriteString(rowIndent + "    .toBeInTheDocument();\n")
			boolIndex2++
		}
	}
	rowAssert2 := strings.TrimRight(rowAssertB2.String(), "\n")

	// Replace regions delimited by markers
	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + "\n" + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END", mockFields)
	if rErr != nil {
		return fmt.Errorf("replacing mock fields: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END", headersAssert)
	if rErr != nil {
		return fmt.Errorf("replacing headers assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END", rowAssert)
	if rErr != nil {
		return fmt.Errorf("replacing row assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_ROW_ASSERT_2_START", "// GF_ROW_ASSERT_2_END", rowAssert2)
	if rErr != nil {
		return fmt.Errorf("replacing second row assert: %w", rErr)
	}

	// Update second test case in testData array with different mock values
	var overridesB strings.Builder
	overridesB.WriteString("                id: \"456\",\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			overridesB.WriteString(fmt.Sprintf("                %s: \"Another %s %d\",\n", c.Name, capitalizedModelName, i))
		case "number":
			overridesB.WriteString(fmt.Sprintf("                %s: \"%d\",\n", c.Name, 30+i))
		case "time":
			overridesB.WriteString(fmt.Sprintf("                %s: \"2024-06-%02dT00:00:00Z\",\n", c.Name, 1+i))
		case "bool":
			overridesB.WriteString(fmt.Sprintf("                %s: false,\n", c.Name))
		}
	}
	overridesText := strings.TrimRight(overridesB.String(), ",\n")

	startMarker := fmt.Sprintf("createMock%s({", capitalizedModelName)
	endMarker := "}),"
	startIndex := strings.Index(s, startMarker)
	if startIndex != -1 {
		endIndex := strings.Index(s[startIndex:], endMarker)
		if endIndex != -1 {
			endIndex += startIndex
			// Find the opening brace of the overrides object
			openBraceIndex := strings.Index(s[startIndex:], "{") + startIndex
			// We want to replace everything between the braces
			oldOverrides := s[openBraceIndex+1 : endIndex]
			s = strings.Replace(s, oldOverrides, "\n"+overridesText+"\n            ", 1)
		}
	}

	// Remove lines that contain marker comments
	markers := []string{
		"// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END",
		"// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END",
		"// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END",
		"// GF_ROW_ASSERT_2_START", "// GF_ROW_ASSERT_2_END",
		"// GF_ROW_SELECT_START", "// GF_ROW_SELECT_END",
		"// GF_ROW_SELECT_DELETE_START", "// GF_ROW_SELECT_DELETE_END",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, m := range markers {
			if strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	// Write out the generated file
	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client list page spec %s: %w", destPath, err)
	}
	return nil
}
