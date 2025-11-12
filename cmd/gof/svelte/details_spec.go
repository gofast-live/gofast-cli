package svelte

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateClientDetailPageSpec scaffolds a client detail page test by copying
// the skeleton detail spec and performing token and marker-based replacements
// for singular/plural model variants and column-aware assertions.
func GenerateClientDetailPageSpec(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/[skeleton_id]/page.svelte.spec.ts"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists: /(app)/models/<plural>/[<model>_id]
	destDir := filepath.Join(
		"app/service-client/src/routes/(app)/models",
		pluralLower,
		"["+modelName+"_id]",
	)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "page.svelte.spec.ts")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}
	s := string(contentBytes)

	// Token replacements (plural/title before singular to avoid partial stomps)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Helper: human readable label from snake_case
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

	// Build detail field configs based on column definitions
	var fieldsB strings.Builder
	for i, c := range columns {
		fieldLabel := toTitle(c.Name)
		if fieldsB.Len() > 0 {
			fieldsB.WriteString("\n")
		}
		fieldsB.WriteString("    {\n")
		switch c.Type {
		case "string":
			fieldsB.WriteString("        type: \"string\",\n")
		case "number":
			fieldsB.WriteString("        type: \"number\",\n")
		case "time":
			fieldsB.WriteString("        type: \"date\",\n")
		case "bool":
			fieldsB.WriteString("        type: \"boolean\",\n")
		default:
			return fmt.Errorf("unsupported column type %q for field %s", c.Type, c.Name)
		}
		fieldsB.WriteString(fmt.Sprintf("        name: %q,\n", c.Name))
		fieldsB.WriteString(fmt.Sprintf("        label: %q,\n", fieldLabel))

		switch c.Type {
		case "string":
			fieldsB.WriteString(fmt.Sprintf("        apiValue: %q,\n", fmt.Sprintf("Existing %s %d", fieldLabel, i+1)))
			fieldsB.WriteString(fmt.Sprintf("        formValue: %q,\n", fmt.Sprintf("New %s %d", fieldLabel, i+1)))
			fieldsB.WriteString(fmt.Sprintf("        updateValue: %q,\n", fmt.Sprintf("Updated %s %d", fieldLabel, i+1)))
		case "number":
			apiNum := 100 + i
			formNum := 200 + i
			updateNum := 300 + i
			fieldsB.WriteString(fmt.Sprintf("        apiValue: %q,\n", fmt.Sprintf("%d", apiNum)))
			fieldsB.WriteString(fmt.Sprintf("        formValue: %q,\n", fmt.Sprintf("%d", formNum)))
			fieldsB.WriteString(fmt.Sprintf("        updateValue: %q,\n", fmt.Sprintf("%d", updateNum)))
		case "time":
			day := i + 1
			fieldsB.WriteString(fmt.Sprintf("        apiValue: %q,\n", fmt.Sprintf("2023-01-%02dT00:00:00Z", day)))
			fieldsB.WriteString(fmt.Sprintf("        displayValue: %q,\n", fmt.Sprintf("2023-01-%02d", day)))
			fieldsB.WriteString(fmt.Sprintf("        formValue: %q,\n", fmt.Sprintf("2024-05-%02d", day)))
			fieldsB.WriteString(fmt.Sprintf("        updateValue: %q,\n", fmt.Sprintf("2024-06-%02d", day)))
		case "bool":
			apiBool := i%2 == 1
			formBool := !apiBool
			updateBool := !apiBool
			fieldsB.WriteString(fmt.Sprintf("        apiValue: %t,\n", apiBool))
			fieldsB.WriteString(fmt.Sprintf("        formValue: %t,\n", formBool))
			fieldsB.WriteString(fmt.Sprintf("        updateValue: %t,\n", updateBool))
		}

		fieldsB.WriteString("    },")
	}
	fieldsSnippet := strings.TrimRight(fieldsB.String(), "\n")

	// Prepare other marker replacements
	modelConfigSnippet := fmt.Sprintf("    name: %q,\n    route: %q,\n    paramIdField: %q,", capitalizedModelName, "/models/"+pluralLower, modelName+"_id")
	toastSnippet := fmt.Sprintf(
		"    createSuccess: %q,\n    updateSuccess: %q,\n    notFound: %q,",
		fmt.Sprintf("%s created successfully.", capitalizedModelName),
		fmt.Sprintf("%s updated successfully.", capitalizedModelName),
		fmt.Sprintf("%s not found.", capitalizedModelName),
	)

	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + "\n" + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_MODEL_CONFIG_START", "// GF_MODEL_CONFIG_END", modelConfigSnippet)
	if rErr != nil {
		return fmt.Errorf("replacing model config: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_FIELDS_CONFIG_START", "// GF_DETAIL_FIELDS_CONFIG_END", fieldsSnippet)
	if rErr != nil {
		return fmt.Errorf("replacing detail field config: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_MODEL_TOAST_MESSAGES_START", "// GF_MODEL_TOAST_MESSAGES_END", toastSnippet)
	if rErr != nil {
		return fmt.Errorf("replacing toast messages: %w", rErr)
	}

	// Remove lines containing generation markers to keep output clean
	markers := []string{
		"// GF_MODEL_CONFIG_START", "// GF_MODEL_CONFIG_END",
		"// GF_DETAIL_FIELDS_CONFIG_START", "// GF_DETAIL_FIELDS_CONFIG_END",
		"// GF_MODEL_TOAST_MESSAGES_START", "// GF_MODEL_TOAST_MESSAGES_END",
		"// GF_MODEL_BASE_FIELDS_START", "// GF_MODEL_BASE_FIELDS_END",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, marker := range markers {
			if strings.Contains(line, marker) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client detail page spec %s: %w", destPath, err)
	}
	return nil
}
