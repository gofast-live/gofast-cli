package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
)

// generateServiceTestContent generates test file by copying skeleton and replacing markers
func generateServiceTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	content := string(contentBytes)

	// Check if model has any date columns (which require time.Now())
	hasDateColumn := false
	for _, c := range columns {
		if c.Type == "date" {
			hasDateColumn = true
			break
		}
	}

	// Remove "time" import if no date columns
	if !hasDateColumn {
		content = strings.Replace(content, "\t\"time\"\n", "", 1)
	}

	// Check if model has any validatable columns (non-bool)
	hasValidatableColumn := false
	for _, c := range columns {
		if c.Type != "bool" {
			hasValidatableColumn = true
			break
		}
	}

	// If all columns are bool, remove the "Failure - Validation Error" test for Create
	// (Edit test is fine because it validates the UUID)
	if !hasValidatableColumn {
		content = removeCreateValidationErrorTest(content)
	}

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

	// Go naming conversions
	goPackageName := toGoPackageName(modelName)
	goVarName := toGoVarName(modelName)
	pluralLower := pluralize.NewClient().Plural(modelName)
	pluralCap := capitalize(pluralLower)
	pluralVarName := toGoVarName(pluralLower)

	// Token replacement
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)
	content = strings.Replace(content, "package skeleton", "package "+goPackageName, 1)
	// Fix import path to use goPackageName (lowercase directory) with alias
	content = strings.Replace(content, `"gofast/service-core/domain/skeleton"`, goVarName+` "gofast/service-core/domain/`+goPackageName+`"`, 1)
	content = strings.ReplaceAll(content, "skeletons", pluralVarName)
	content = strings.ReplaceAll(content, "skeleton", goVarName)

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
			lines = append(lines, field+":    \"100\",")
		case "date":
			lines = append(lines, field+":  time.Now(),")
		case "bool":
			lines = append(lines, field+": true,")
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
			lines = append(lines, field+":    \"100\",")
		case "date":
			lines = append(lines, field+":  \"2023-10-31\",")
		case "bool":
			lines = append(lines, field+": true,")
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
			lines = append(lines, field+":  \"\",")
		case "number":
			lines = append(lines, field+":   \"invalid\",")
		case "date":
			lines = append(lines, field+": \"bad-date\",")
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
			lines = append(lines, field+":    \"200\",")
		case "date":
			lines = append(lines, field+":  \"2024-01-01\",")
		case "bool":
			lines = append(lines, field+": false,")
		}
	}
	return strings.Join(lines, "\n\t\t\t\t")
}

// buildEditAssertFields generates assertion lines for the edit transport test
func buildEditAssertFields(columns []Column, modelName string) string {
	var lines []string
	for _, c := range columns {
		field := toCamelCase(c.Name)
		getter := "Get" + field + "()"
		switch c.Type {
		case "string":
			lines = append(lines, fmt.Sprintf("assert.Equal(t, \"Updated %s\", res.Msg.Get%s().%s)", capitalize(c.Name), modelName, getter))
		case "number":
			lines = append(lines, fmt.Sprintf("assert.Equal(t, \"200\", res.Msg.Get%s().%s)", modelName, getter))
		case "date":
			lines = append(lines, fmt.Sprintf("assert.NotEmpty(t, res.Msg.Get%s().%s)", modelName, getter))
		case "bool":
			lines = append(lines, fmt.Sprintf("assert.Equal(t, false, res.Msg.Get%s().%s)", modelName, getter))
		}
	}
	return strings.Join(lines, "\n\t\t")
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
			var tabsOnly strings.Builder
			for _, ch := range linePrefix {
				if ch != '\t' && ch != ' ' {
					break
				}
				tabsOnly.WriteRune(ch)
			}
			indent = tabsOnly.String()
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

	content := renderValidationFixtures(string(contentBytes), capitalizedModelName, columns)

	// Go naming conversions
	goPackageName := toGoPackageName(modelName)
	goVarName := toGoVarName(modelName)
	pluralLower := pluralize.NewClient().Plural(modelName)
	pluralCap := capitalize(pluralLower)
	pluralVarName := toGoVarName(pluralLower)

	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)
	content = strings.Replace(content, "package skeleton", "package "+goPackageName, 1)
	// Fix import path to use goPackageName (lowercase directory) with alias
	content = strings.Replace(content, `"gofast/service-core/domain/skeleton"`, goVarName+` "gofast/service-core/domain/`+goPackageName+`"`, 1)
	content = strings.ReplaceAll(content, "skeletons", pluralVarName)
	content = strings.ReplaceAll(content, "skeleton", goVarName)

	header := "\ttestCases := []struct {\n\t\tname           string\n\t\t" + goVarName + "       *proto." + capitalizedModelName + "\n\t\texpectError    bool\n\t\texpectedErrors []pkg.ValidationError\n\t}{\n"
	footer := "\t}\n"
	createCall := func(args string) string {
		return fmt.Sprintf("makeCreate%sProto(%s)", capitalizedModelName, args)
	}
	editCall := func(args string) string {
		return fmt.Sprintf("makeEdit%sProto(uuid.New().String(), %s)", capitalizedModelName, args)
	}

	var insertCases strings.Builder
	// Valid case (bools true)
	fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:    false,\n\t\t\texpectedErrors: nil,\n\t\t},\n", modelName, goVarName, capitalizedModelName, strings.Join(validArgs(columns, true), ", "))
	insertCases.WriteString(invalidColumnCases(columns, goVarName, createCall))

	var updateCases strings.Builder
	// Valid case
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:    false,\n\t\t\texpectedErrors: nil,\n\t\t},\n", modelName, goVarName, capitalizedModelName, strings.Join(validArgs(columns, true), ", "))
	// invalid uuid case -> expect two errors
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"invalid uuid\",\n\t\t\t%s: makeEdit%sProto(\"invalid-uuid\", %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"id\", Tag: \"uuid\", Message: \"ID must be a valid UUID\"},\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", goVarName, capitalizedModelName, strings.Join(validArgs(columns, false), ", "))
	// nil uuid case -> required only
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"nil uuid\",\n\t\t\t%s: makeEdit%sProto(uuid.Nil.String(), %s),\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", goVarName, capitalizedModelName, strings.Join(validArgs(columns, false), ", "))
	updateCases.WriteString(invalidColumnCases(columns, goVarName, editCall))

	content, err = replaceTestCases(content, "TestValidateAndBuildInsertParams", header+insertCases.String()+footer)
	if err != nil {
		return "", err
	}
	content, err = replaceTestCases(content, "TestValidateAndBuildUpdateParams", header+updateCases.String()+footer)
	if err != nil {
		return "", err
	}

	return content, nil
}

// renderValidationFixtures fills the GF_FIXTURES region with makeCreate/makeEdit proto helpers.
func renderValidationFixtures(template, capitalizedModelName string, columns []Column) string {
	// Build params signature and body fields for create and edit helpers.
	// Edit's first param is the id.
	createParams := make([]string, 0, len(columns))
	editParams := make([]string, 0, len(columns)+1)
	editParams = append(editParams, "id string")
	fields := make([]string, 0, len(columns))
	for _, c := range columns {
		field := toCamelCase(c.Name)
		varType := "string"
		if c.Type == "bool" {
			varType = "bool"
		}
		vn := strings.ToLower(field[:1]) + field[1:]
		createParams = append(createParams, fmt.Sprintf("%s %s", vn, varType))
		editParams = append(editParams, fmt.Sprintf("%s %s", vn, varType))
		fields = append(fields, fmt.Sprintf("%s: %s,", field, vn))
	}

	lines := strings.Split(template, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) != "// GF_FIXTURES_START" {
			out = append(out, line)
			continue
		}
		out = append(out, line)
		indent := strings.Repeat("\t", strings.Count(line, "\t"))
		// makeCreate<Model>Proto
		out = append(out, indent+fmt.Sprintf("func makeCreate%sProto(%s) *proto.%s {", capitalizedModelName, strings.Join(createParams, ", "), capitalizedModelName))
		out = append(out, indent+"\treturn &proto."+capitalizedModelName+"{")
		out = append(out, indent+"\t\tId: \"\",")
		out = append(out, indent+"\t\tCreated: \"\",")
		out = append(out, indent+"\t\tUpdated: \"\",")
		for _, f := range fields {
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
		for _, f := range fields {
			out = append(out, indent+"\t\t"+f)
		}
		out = append(out, indent+"\t}")
		out = append(out, indent+"}")

		// Skip the template's fixtures up to GF_FIXTURES_END, which the next iteration keeps
		for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "// GF_FIXTURES_END" {
			i++
		}
	}
	return strings.Join(out, "\n")
}

// validArgs returns one valid literal per column. Bools are all true or all false.
func validArgs(columns []Column, boolValue bool) []string {
	args := make([]string, 0, len(columns))
	for _, c := range columns {
		switch c.Type {
		case "string":
			args = append(args, "\"Valid\"")
		case "number":
			args = append(args, "\"10\"")
		case "date":
			args = append(args, "\"2025-01-01\"")
		case "bool":
			args = append(args, strconv.FormatBool(boolValue))
		default:
			args = append(args, "\"\"")
		}
	}
	return args
}

// invalidColumnCases renders one failing test case per validation rule of each column.
// protoCall wraps an argument list in the create or edit proto constructor.
func invalidColumnCases(columns []Column, goVarName string, protoCall func(args string) string) string {
	withArg := func(idx int, value string) string {
		args := validArgs(columns, false)
		args[idx] = value
		return protoCall(strings.Join(args, ", "))
	}
	var b strings.Builder
	for idx, c := range columns {
		fieldCamel := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\t\t{\n\t\t\tname: \"%s too short\",\n\t\t\t%s: %s,\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"},\n\t\t\t},\n\t\t},\n", c.Name, goVarName, withArg(idx, "\"ab\""), c.Name, fieldCamel)
		case "number":
			fmt.Fprintf(&b, "\t\t{\n\t\t\tname: \"%s is not a number\",\n\t\t\t%s: %s,\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"},\n\t\t\t},\n\t\t},\n", c.Name, goVarName, withArg(idx, "\"ten\""), c.Name, fieldCamel)
			fmt.Fprintf(&b, "\t\t{\n\t\t\tname: \"%s less than 1\",\n\t\t\t%s: %s,\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"},\n\t\t\t},\n\t\t},\n", c.Name, goVarName, withArg(idx, "\"0\""), c.Name, fieldCamel)
		case "date":
			fmt.Fprintf(&b, "\t\t{\n\t\t\tname: \"invalid %s date\",\n\t\t\t%s: %s,\n\t\t\texpectError:    true,\n\t\t\texpectedErrors: []pkg.ValidationError{\n\t\t\t\t{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"},\n\t\t\t},\n\t\t},\n", c.Name, goVarName, withArg(idx, "\"invalid-date\""), c.Name, fieldCamel)
		}
	}
	return b.String()
}

// replaceTestCases swaps the testCases slice literal inside funcName for block.
func replaceTestCases(src, funcName, block string) (string, error) {
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
	// The slice literal closes at the last '}' before the for loop
	beforeFor := src[tcStart : tcStart+forIdx]
	closeIdx := strings.LastIndex(beforeFor, "}\n")
	if closeIdx == -1 {
		closeIdx = strings.LastIndex(beforeFor, "}")
	}
	if closeIdx == -1 {
		return src, fmt.Errorf("cannot locate end of testCases in %s", funcName)
	}
	endPos := tcStart + closeIdx + 1
	return src[:tcStart] + block + src[endPos:], nil
}

// removeCreateValidationErrorTest removes the "Failure - Validation Error" t.Run block
// from TestService_CreateSkeleton. Used when model has only bool columns (no validation).
func removeCreateValidationErrorTest(content string) string {
	// Find TestService_CreateSkeleton function
	funcStart := strings.Index(content, "func TestService_CreateSkeleton(t *testing.T)")
	if funcStart == -1 {
		return content
	}

	// Find the validation error test block within this function
	searchStart := funcStart
	marker := `t.Run("Failure - Validation Error", func(t *testing.T) {`
	blockStart := strings.Index(content[searchStart:], marker)
	if blockStart == -1 {
		return content
	}
	blockStart += searchStart

	// Find the start of this line (including leading whitespace)
	lineStart := strings.LastIndex(content[:blockStart], "\n") + 1

	// Find the matching closing brace by counting braces
	braceCount := 0
	pos := blockStart
	foundOpen := false
	for pos < len(content) {
		if content[pos] == '{' {
			braceCount++
			foundOpen = true
		} else if content[pos] == '}' {
			braceCount--
			if foundOpen && braceCount == 0 {
				break
			}
		}
		pos++
	}

	// pos now points to the closing brace, find end of line
	lineEnd := strings.Index(content[pos:], "\n")
	if lineEnd == -1 {
		lineEnd = len(content) - pos
	}
	blockEnd := pos + lineEnd + 1

	// Remove the block
	return content[:lineStart] + content[blockEnd:]
}
