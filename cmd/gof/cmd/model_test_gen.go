package cmd

import (
	"fmt"
	"os"
	"strings"
)

// generateServiceTestContent generates test file by copying skeleton and replacing markers
func generateServiceTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	content := string(contentBytes)

	// Build replacement content for each marker type
	entityFields := buildEntityFields(columns)
	createFields := buildCreateProtoFields(columns, capitalizedModelName)
	editFields := buildEditProtoFields(columns)
	invalidFields := buildInvalidProtoFields(columns)

	// Replace marker regions
	content = replaceMarkerRegion(content, "GF_TP_TEST_ENTITY_FIELDS_START", "GF_TP_TEST_ENTITY_FIELDS_END", entityFields)
	content = replaceMarkerRegion(content, "GF_TP_TEST_CREATE_FIELDS_START", "GF_TP_TEST_CREATE_FIELDS_END", createFields)
	content = replaceMarkerRegion(content, "GF_TP_TEST_EDIT_FIELDS_START", "GF_TP_TEST_EDIT_FIELDS_END", editFields)
	content = replaceMarkerRegion(content, "GF_TP_TEST_INVALID_FIELDS_START", "GF_TP_TEST_INVALID_FIELDS_END", invalidFields)

	// Token replacement
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "skeletons", pluralLower)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}

// buildEntityFields generates InsertParams fields for createTest<Model> helper
func buildEntityFields(columns []Column) string {
	var lines []string
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			lines = append(lines, fmt.Sprintf("%s:   \"%s \" + uuid.New().String()[:8],", field, capitalize(c.Name)))
		case "number":
			lines = append(lines, fmt.Sprintf("%s:    \"100\",", field))
		case "date":
			lines = append(lines, fmt.Sprintf("%s:  time.Now(),", field))
		case "bool":
			lines = append(lines, fmt.Sprintf("%s: true,", field))
		}
	}
	return strings.Join(lines, "\n\t\t")
}

// buildCreateProtoFields generates proto fields for create request (full version)
func buildCreateProtoFields(columns []Column, modelName string) string {
	var lines []string
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			lines = append(lines, fmt.Sprintf("%s:   \"Test %s\",", field, modelName))
		case "number":
			lines = append(lines, fmt.Sprintf("%s:    \"100\",", field))
		case "date":
			lines = append(lines, fmt.Sprintf("%s:  \"2023-10-31\",", field))
		case "bool":
			lines = append(lines, fmt.Sprintf("%s: true,", field))
		}
	}
	return strings.Join(lines, "\n\t\t\t\t")
}

// buildInvalidProtoFields generates proto fields with invalid values for validation error tests
func buildInvalidProtoFields(columns []Column) string {
	var lines []string
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			lines = append(lines, fmt.Sprintf("%s:  \"\",", field))
		case "number":
			lines = append(lines, fmt.Sprintf("%s:   \"invalid\",", field))
		case "date":
			lines = append(lines, fmt.Sprintf("%s: \"bad-date\",", field))
		case "bool":
			// bools don't have invalid values, skip or use false
		}
	}
	return strings.Join(lines, "\n\t\t\t\t")
}

// buildEditProtoFields generates proto fields for edit request
func buildEditProtoFields(columns []Column) string {
	var lines []string
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			lines = append(lines, fmt.Sprintf("%s:   \"Updated %s\",", field, capitalize(c.Name)))
		case "number":
			lines = append(lines, fmt.Sprintf("%s:    \"200\",", field))
		case "date":
			lines = append(lines, fmt.Sprintf("%s:  \"2024-01-01\",", field))
		case "bool":
			lines = append(lines, fmt.Sprintf("%s: false,", field))
		}
	}
	return strings.Join(lines, "\n\t\t\t\t")
}

// replaceMarkerRegion replaces content between START and END markers (removes markers)
func replaceMarkerRegion(content, startMarker, endMarker, replacement string) string {
	for {
		startIdx := strings.Index(content, startMarker)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(content[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// Find the start of the start marker line
		startLineStart := strings.LastIndex(content[:startIdx], "\n") + 1

		// Find the newline after end marker
		endLineEnd := strings.Index(content[endIdx:], "\n")
		if endLineEnd == -1 {
			endLineEnd = len(content) - endIdx
		}
		endLineEnd += endIdx + 1

		// Detect indent from the start marker line (tabs only, exclude comment prefix)
		indent := "\t\t"
		if startLineStart < startIdx {
			linePrefix := content[startLineStart:startIdx]
			// Extract only whitespace (tabs/spaces), not comment characters
			tabsOnly := ""
			for _, ch := range linePrefix {
				if ch == '\t' || ch == ' ' {
					tabsOnly += string(ch)
				} else {
					break
				}
			}
			indent = tabsOnly
		}

		// Build replacement with proper indent
		replacementIndented := indent + replacement + "\n"

		content = content[:startLineStart] + replacementIndented + content[endLineEnd:]
	}
	return content
}

func generateValidationTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/validation_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	toFieldName := func(col string) string { return toCamelCase(col) }
	toVarName := func(camel string) string {
		if camel == "" {
			return camel
		}
		return strings.ToLower(camel[:1]) + camel[1:]
	}
	// Build params signature and body fields for create and edit helpers
	var createParams []string
	var createFields []string
	var editParams []string
	var editFields []string

	// Edit first param is id string
	editParams = append(editParams, "id string")

	for _, c := range columns {
		field := toFieldName(c.Name)
		varType := "string"
		switch c.Type {
		case "string":
			varType = "string"
		case "number":
			varType = "string"
		case "date":
			varType = "string"
		case "bool":
			varType = "bool"
		}
		vn := toVarName(field)
		createParams = append(createParams, fmt.Sprintf("%s %s", vn, varType))
		editParams = append(editParams, fmt.Sprintf("%s %s", vn, varType))
		createFields = append(createFields, fmt.Sprintf("%s: %s,", field, vn))
		editFields = append(editFields, fmt.Sprintf("%s: %s,", field, vn))
	}

	// Render into fixtures region
	lines := strings.Split(string(contentBytes), "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "// GF_FIXTURES_START":
			out = append(out, line)
			indent := strings.Repeat("\t", strings.Count(line, "\t"))
			// makeCreate<Model>Proto
			out = append(out, indent+fmt.Sprintf("func makeCreate%sProto(%s) *proto.%s {", capitalizedModelName, strings.Join(createParams, ", "), capitalizedModelName))
			out = append(out, indent+"\treturn &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\tId: \"\",")
			out = append(out, indent+"\t\tCreated: \"\",")
			out = append(out, indent+"\t\tUpdated: \"\",")
			for _, f := range createFields {
				out = append(out, indent+"\t\t"+f)
			}
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			// makeEdit<Model>Proto
			out = append(out, indent+fmt.Sprintf("func makeEdit%sProto(%s) *proto.%s {", capitalizedModelName, strings.Join(editParams, ", "), capitalizedModelName))
			out = append(out, indent+"\treturn &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\tId: id,")
			out = append(out, indent+"\t\tCreated: \"\",")
			out = append(out, indent+"\t\tUpdated: \"\",")
			for _, f := range editFields {
				out = append(out, indent+"\t\t"+f)
			}
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")

			// Skip lines until END
			for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "// GF_FIXTURES_END" {
				i++
			}
		case "// GF_FIXTURES_END":
			out = append(out, line)
		default:
			out = append(out, line)
		}
	}

	content := strings.Join(out, "\n")
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "skeletons", pluralLower)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	// Helper for building default valid args per type
	buildValidArgs := func(boolTrue bool) []string {
		args := []string{}
		for _, c := range columns {
			switch c.Type {
			case "string":
				args = append(args, "\"Valid\"")
			case "number":
				args = append(args, "\"10\"")
			case "date":
				args = append(args, "\"2025-01-01\"")
			case "bool":
				if boolTrue {
					args = append(args, "true")
				} else {
					args = append(args, "false")
				}
			default:
				args = append(args, "\"\"")
			}
		}
		return args
	}
	// Insert testCases generation
	insertHeader := "\ttestCases := []struct {\n\t\tname           string\n\t\t" + modelName + "       *proto." + capitalizedModelName + "\n\t\texpectError    bool\n\t\texpectedErrors []pkg.ValidationError\n\t}{\n"

	var insertCases strings.Builder
	// Valid case (bools true)
	fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    false,\n\t\t\texpectedErrors: nil,\n\t\t},\n", modelName, modelName, capitalizedModelName, strings.Join(buildValidArgs(true), ", "))

	// Per-column invalid cases for insert
	for _, c := range columns {
		fieldCamel := toFieldName(c.Name)
		// Build args default with bools false
		args := buildValidArgs(false)
		switch c.Type {
		case "string":
			// too short
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"ab\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"%s too short\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "number":
			argsNotNumber := buildValidArgs(false)
			for i := range columns {
				if columns[i].Name == c.Name {
					argsNotNumber[i] = "\"ten\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"%s is not a number\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(argsNotNumber, ", "), c.Name, fieldCamel)

			argsLess := buildValidArgs(false)
			for i := range columns {
				if columns[i].Name == c.Name {
					argsLess[i] = "\"0\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"%s less than 1\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(argsLess, ", "), c.Name, fieldCamel)
		case "date":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"invalid-date\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"invalid %s date\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		}
	}
	insertFooter := "\t}\n"

	// Update testCases generation
	updateHeader := "\ttestCases := []struct {\n\t\tname           string\n\t\t" + modelName + "       *proto." + capitalizedModelName + "\n\t\texpectError    bool\n\t\texpectedErrors []pkg.ValidationError\n\t}{\n"
	var updateCases strings.Builder
	// Valid case
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    false,\n\t\t\texpectedErrors: nil,\n\t\t},\n", modelName, modelName, capitalizedModelName, strings.Join(buildValidArgs(true), ", "))
	// invalid uuid case -> expect two errors
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"invalid uuid\",\n\t\t\t%s: makeEdit%sProto(\"invalid-uuid\", %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"id\", Tag: \"uuid\", Message: \"ID must be a valid UUID\"},\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", modelName, capitalizedModelName, strings.Join(buildValidArgs(false), ", "))
	// nil uuid case -> required only
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"nil uuid\",\n\t\t\t%s: makeEdit%sProto(uuid.Nil.String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", modelName, capitalizedModelName, strings.Join(buildValidArgs(false), ", "))
	// Per-column invalid cases for update
	for _, c := range columns {
		fieldCamel := toFieldName(c.Name)
		args := buildValidArgs(false)
		switch c.Type {
		case "string":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"ab\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"%s too short\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "number":
			argsNotNumber := buildValidArgs(false)
			for i := range columns {
				if columns[i].Name == c.Name {
					argsNotNumber[i] = "\"ten\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"%s is not a number\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(argsNotNumber, ", "), c.Name, fieldCamel)

			argsLess := buildValidArgs(false)
			for i := range columns {
				if columns[i].Name == c.Name {
					argsLess[i] = "\"0\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"%s less than 1\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(argsLess, ", "), c.Name, fieldCamel)
		case "date":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"invalid-date\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"invalid %s date\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		}
	}
	updateFooter := "\t}\n"

	// Replace the testCases blocks in both tests
	replaceCases := func(src string, funcName string, header string, cases string, footer string) (string, error) {
		fnIdx := strings.Index(src, funcName)
		if fnIdx == -1 {
			return src, fmt.Errorf("function %s not found", funcName)
		}
		tcIdx := strings.Index(src[fnIdx:], "testCases := []struct {")
		if tcIdx == -1 {
			return src, fmt.Errorf("testCases block not found in %s", funcName)
		}
		tcStart := fnIdx + tcIdx
		// Find the for-loop that iterates over testCases after tcStart
		forIdx := strings.Index(src[tcStart:], "for _, tc := range testCases")
		if forIdx == -1 {
			return src, fmt.Errorf("for loop after testCases not found in %s", funcName)
		}
		// Walk backwards from forIdx start to find the closing brace of the slice literal
		pre := src[:tcStart]
		rest := src[tcStart:]
		// Find the first '}' before the for loop start
		beforeFor := rest[:forIdx]
		closeIdx := strings.LastIndex(beforeFor, "}\n")
		if closeIdx == -1 {
			// try just '}' without newline
			closeIdx = strings.LastIndex(beforeFor, "}")
			if closeIdx == -1 {
				return src, fmt.Errorf("cannot locate end of testCases in %s", funcName)
			}
		}
		endPos := tcStart + closeIdx + 1
		newBlock := header + cases + footer
		return pre + newBlock + src[endPos:], nil
	}

	var rErr error
	content, rErr = replaceCases(content, "TestValidateAndBuildInsertParams", insertHeader, insertCases.String(), insertFooter)
	if rErr != nil {
		return "", rErr
	}
	content, rErr = replaceCases(content, "TestValidateAndBuildUpdateParams", updateHeader, updateCases.String(), updateFooter)
	if rErr != nil {
		return "", rErr
	}

	return content, nil
}
