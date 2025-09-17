package svelte

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateClientE2ETest scaffolds a Playwright e2e test based on the skeleton
// template, expanding the model configuration block with column-aware values
// and default behaviours.
func generateClientE2ETest(modelName string, columns []Column) error {
	sourcePath := "./e2e/skeletons.test.ts"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	if err := os.MkdirAll("e2e", 0o755); err != nil {
		return fmt.Errorf("creating e2e directory: %w", err)
	}
	destPath := filepath.Join("e2e", pluralLower+".test.ts")

	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading e2e template %s: %w", sourcePath, err)
	}

	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

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

	type fieldMeta struct {
		name          string
		label         string
		typeLiteral   string
		createLiteral string
		validation    string
		useTimestamp  bool
		createBool    *bool
	}

	headers := make([]string, 0, len(columns)+2)
	fieldMetas := make([]fieldMeta, 0, len(columns))
	stringTimestampAssigned := false

	for i, c := range columns {
		label := toTitle(c.Name)
		headers = append(headers, label)

		meta := fieldMeta{
			name:  c.Name,
			label: label,
		}

		switch c.Type {
		case "string":
			meta.typeLiteral = "'string'"
			meta.createLiteral = fmt.Sprintf("'Test %s %d'", label, i+1)
			meta.validation = "'Enter at least 3 characters'"
			if !stringTimestampAssigned {
				meta.useTimestamp = true
				stringTimestampAssigned = true
			}
		case "number":
			meta.typeLiteral = "'number'"
			meta.createLiteral = fmt.Sprintf("'%d'", 100+i)
			meta.validation = "'Enter a positive number'"
		case "time":
			meta.typeLiteral = "'date'"
			meta.createLiteral = fmt.Sprintf("'2025-01-%02d'", i+1)
			meta.validation = "'Select a valid date'"
		case "bool":
			meta.typeLiteral = "'boolean'"
			boolVal := i%2 == 0
			meta.createLiteral = fmt.Sprintf("%t", boolVal)
			meta.createBool = &boolVal
		default:
			return fmt.Errorf("unsupported column type %q for e2e generation", c.Type)
		}

		fieldMetas = append(fieldMetas, meta)
	}
	headers = append(headers, "Created", "Updated")

	createAssertField := fieldMetas[0].name
	for _, meta := range fieldMetas {
		if meta.typeLiteral == "'string'" {
			createAssertField = meta.name
			break
		}
	}

	editMeta := fieldMetas[0]
	for _, meta := range fieldMetas {
		if meta.typeLiteral == "'string'" {
			editMeta = meta
			break
		}
	}

	var editValueLiteral string
	switch editMeta.typeLiteral {
	case "'string'":
		editValueLiteral = fmt.Sprintf("'Edited %s'", editMeta.label)
	case "'number'":
		editValueLiteral = "'200'"
	case "'date'":
		editValueLiteral = "'2026-02-01'"
	case "'boolean'":
		newVal := true
		if editMeta.createBool != nil {
			newVal = !*editMeta.createBool
		}
		editValueLiteral = fmt.Sprintf("%t", newVal)
	default:
		return fmt.Errorf("unsupported edit type literal %s", editMeta.typeLiteral)
	}

	var configB strings.Builder
	configB.WriteString(fmt.Sprintf("\tname: '%s',\n", capitalizedModelName))
	configB.WriteString(fmt.Sprintf("\tplural: '%s',\n", pluralCap))
	configB.WriteString(fmt.Sprintf("\troute: '/models/%s',\n", pluralLower))
	configB.WriteString(fmt.Sprintf("\tcreateRoute: '/models/%s/new',\n", pluralLower))
	configB.WriteString(fmt.Sprintf("\tcreateLinkLabel: 'Create New %s',\n", capitalizedModelName))
	configB.WriteString("\tsaveButtonLabel: 'Save',\n")
	configB.WriteString("\tdeleteButtonLabel: 'Delete',\n")
	configB.WriteString("\tlistHeaders: [\n")
	for _, header := range headers {
		configB.WriteString(fmt.Sprintf("\t\t'%s',\n", header))
	}
	configB.WriteString("\t],\n")
	configB.WriteString("\ttoastMessages: {\n")
	configB.WriteString(fmt.Sprintf("\t\tupdateSuccess: '%s updated successfully.',\n", capitalizedModelName))
	configB.WriteString("\t},\n")
	configB.WriteString("\tfields: [\n")
	for _, meta := range fieldMetas {
		configB.WriteString("\t\t{\n")
		configB.WriteString(fmt.Sprintf("\t\t\tname: '%s',\n", meta.name))
		configB.WriteString(fmt.Sprintf("\t\t\tlabel: '%s',\n", meta.label))
		configB.WriteString(fmt.Sprintf("\t\t\ttype: %s,\n", meta.typeLiteral))
		configB.WriteString(fmt.Sprintf("\t\t\tcreateValue: %s,\n", meta.createLiteral))
		if meta.validation != "" {
			configB.WriteString(fmt.Sprintf("\t\t\tvalidationMessage: %s,\n", meta.validation))
		}
		if meta.useTimestamp {
			configB.WriteString("\t\t\tuseTimestamp: true,\n")
		}
		configB.WriteString("\t\t},\n")
	}
	configB.WriteString("\t],\n")
	configB.WriteString(fmt.Sprintf("\tcreateAssertField: '%s',\n", createAssertField))
	configB.WriteString("\teditScenario: {\n")
	configB.WriteString(fmt.Sprintf("\t\tfieldName: '%s',\n", editMeta.name))
	configB.WriteString(fmt.Sprintf("\t\tnewValue: %s,\n", editValueLiteral))
	configB.WriteString("\t},\n")

	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		start += len(startMarker)
		return content[:start] + "\n" + replacement + content[end:], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_MODEL_CONFIG_START", "// GF_MODEL_CONFIG_END", configB.String())
	if rErr != nil {
		return fmt.Errorf("replacing model config: %w", rErr)
	}

	markers := []string{"// GF_MODEL_CONFIG_START", "// GF_MODEL_CONFIG_END"}
	var outLines []string
	for _, line := range strings.Split(s, "\n") {
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
		return fmt.Errorf("writing e2e test %s: %w", destPath, err)
	}
	return nil
}
