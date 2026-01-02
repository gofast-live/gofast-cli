package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateServiceContent(modelName string, capitalizedModelName string) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}
	content := string(contentBytes)

	// Go naming conversions
	goPackageName := toGoPackageName(modelName)
	goVarName := toGoVarName(modelName)
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	pluralVarName := toGoVarName(pluralLower)

	// Template already follows ConnectRPC and builds params via validation helpers.
	// Replace in order: PascalCase types first, then lowercase (order matters to avoid partial replacements)
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)
	// Replace package declaration specifically
	content = strings.Replace(content, "package skeleton", "package "+goPackageName, 1)
	// Replace variable names (skeletons -> pluralVarName, skeleton -> goVarName)
	content = strings.ReplaceAll(content, "skeletons", pluralVarName)
	content = strings.ReplaceAll(content, "skeleton", goVarName)

	return content, nil
}

func generateServiceLayer(modelName string, columns []Column) error {
	sourceDir := "./app/service-core/domain/skeleton"
	goPackageName := toGoPackageName(modelName)
	destDir := "app/service-core/domain/" + goPackageName
	capitalizedModelName := capitalize(modelName)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := strings.Replace(path, sourceDir, destDir, 1)
		destPath = strings.ReplaceAll(destPath, "skeleton", goPackageName)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		var newContentStr string
		var genErr error
		if info.Name() == "service.go" {
			newContentStr, genErr = generateServiceContent(modelName, capitalizedModelName)
		} else if info.Name() == "service_test.go" {
			newContentStr, genErr = generateServiceTestContent(modelName, capitalizedModelName, columns)
		} else if info.Name() == "validation.go" {
			newContentStr, genErr = generateValidationContent(modelName, capitalizedModelName, columns)
		} else if info.Name() == "validation_test.go" {
			newContentStr, genErr = generateValidationTestContent(modelName, capitalizedModelName, columns)
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			pluralLower := pluralizeClient.Plural(modelName)
			pluralCap := capitalize(pluralLower)
			pluralVarName := toGoVarName(pluralLower)
			newContentStr = strings.ReplaceAll(string(content), "Skeletons", pluralCap)
			newContentStr = strings.ReplaceAll(newContentStr, "Skeleton", capitalizedModelName)
			newContentStr = strings.Replace(newContentStr, "package skeleton", "package "+goPackageName, 1)
			newContentStr = strings.ReplaceAll(newContentStr, "skeletons", pluralVarName)
			newContentStr = strings.ReplaceAll(newContentStr, "skeleton", toGoVarName(modelName))
		}

		if genErr != nil {
			return fmt.Errorf("generating content for %s: %w", destPath, genErr)
		}

		return os.WriteFile(destPath, []byte(newContentStr), info.Mode())
	})
}

// generateTransportLayer scaffolds ConnectRPC handlers by copying the transport
// skeleton and performing token replacements for singular/plural variants.
func generateTransportLayer(modelName string, columns []Column) error {
	sourceDir := "./app/service-core/transport/skeleton"
	goPackageName := toGoPackageName(modelName)
	destDir := "app/service-core/transport/" + goPackageName

	capitalizedModelName := capitalize(modelName)
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := strings.Replace(path, sourceDir, destDir, 1)
		destPath = strings.ReplaceAll(destPath, "skeleton", goPackageName)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		var newContentStr string
		var genErr error
		switch info.Name() {
		case "route.go":
			newContentStr, genErr = generateTransportRouteContent(modelName, capitalizedModelName, pluralLower, pluralCap, columns)
		case "route_test.go":
			newContentStr, genErr = generateTransportTestContent(modelName, capitalizedModelName, columns, pluralLower, pluralCap)
		default:
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			s := string(content)
			goVarName := toGoVarName(modelName)
			pluralVarName := toGoVarName(pluralLower)
			s = strings.ReplaceAll(s, "Skeletons", pluralCap)
			s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
			s = strings.Replace(s, "package skeleton", "package "+goPackageName, 1)
			s = strings.ReplaceAll(s, "skeletons", pluralVarName)
			s = strings.ReplaceAll(s, "skeleton", goVarName)
			newContentStr = s
		}
		if genErr != nil {
			return fmt.Errorf("generating transport content for %s: %w", destPath, genErr)
		}
		return os.WriteFile(destPath, []byte(newContentStr), info.Mode())
	})
}

func generateTransportRouteContent(modelName, capitalizedModelName, pluralLower, pluralCap string, columns []Column) (string, error) {
	templatePath := "./app/service-core/transport/skeleton/route.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}
	s := string(contentBytes)

	// Go naming conversions
	goPackageName := toGoPackageName(modelName)
	goVarName := toGoVarName(modelName)
	pluralVarName := toGoVarName(pluralLower)

	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.Replace(s, "package skeleton", "package "+goPackageName, 1)
	// Replace import path with alias: goVarName "gofast/service-core/domain/goPackageName"
	s = strings.Replace(s, `"gofast/service-core/domain/skeleton"`, goVarName+` "gofast/service-core/domain/`+goPackageName+`"`, 1)
	s = strings.ReplaceAll(s, "skeletons", pluralVarName)
	s = strings.ReplaceAll(s, "skeleton", goVarName)
	// Rename leftover template-local variable names
	s = strings.ReplaceAll(s, "skelProto", goVarName+"Proto")

	// Build dynamic queryToProto mapping based on columns
	// Numeric columns are strings in proto; no strconv needed

	// Construct struct fields
	var b strings.Builder
	b.WriteString("\t\tId:      " + goVarName + ".ID.String(),\n")
	b.WriteString("\t\tCreated: " + goVarName + ".Created.Format(time.RFC3339),\n")
	b.WriteString("\t\tUpdated: " + goVarName + ".Updated.Format(time.RFC3339),\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			b.WriteString("\t\t" + field + ": " + goVarName + "." + field + ",\n")
		case "date":
			b.WriteString("\t\t" + field + ": " + goVarName + "." + field + ".Format(time.RFC3339),\n")
		case "bool":
			b.WriteString("\t\t" + field + ": " + goVarName + "." + field + ",\n")
		case "number":
			// sqlc numeric is string; proto is also string
			b.WriteString("\t\t" + field + ": " + goVarName + "." + field + ",\n")
		}
	}
	fields := b.String()

	// Replace the queryToProto function body
	fnStart := strings.Index(s, "func queryToProto(")
	if fnStart == -1 {
		return "", fmt.Errorf("queryToProto function not found in transport template")
	}
	// Find function end by counting braces
	braceIdx := strings.Index(s[fnStart:], "{")
	if braceIdx == -1 {
		return "", fmt.Errorf("malformed queryToProto: no opening brace")
	}
	absOpen := fnStart + braceIdx
	depth := 0
	end := absOpen
	for i := absOpen; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				i = len(s)
			}
		}
	}
	newFn := "func queryToProto(" + goVarName + " *query." + capitalizedModelName + ") *proto." + capitalizedModelName + " {\n\treturn &proto." + capitalizedModelName + "{\n" + fields + "\t}\n}\n"
	s = s[:fnStart] + newFn + s[end:]
	return s, nil
}

func generateValidationContent(modelName string, capitalizedModelName string, columns []Column) (string, error) {
	// Determine which imports are needed based on column types
	needStrconv := false
	needStr := false
	for _, c := range columns {
		switch c.Type {
		case "number":
			needStrconv = true
		case "date":
			needStr = true
		}
	}

	// Build imports
	imports := make([]string, 0, 5)
	if needStr {
		imports = append(imports, "\"gofast/pkg/str\"")
	}
	imports = append(imports,
		"\"gofast/pkg\"",
		"\"gofast/service-core/storage/query\"",
		"proto \"gofast/gen/proto/v1\"",
		"\"github.com/google/uuid\"",
	)
	if needStrconv {
		imports = append(imports, "\"strconv\"")
	}

	// Go naming conversions
	goPackageName := toGoPackageName(modelName)
	goVarName := toGoVarName(modelName)

	// Helpers
	toFieldName := func(name string) string { return toCamelCase(name) }
	toLocalVarName := func(camel string) string {
		if camel == "" {
			return camel
		}
		return strings.ToLower(camel[:1]) + camel[1:]
	}

	// Begin file content
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", goPackageName)
	b.WriteString("import (\n")
	for _, imp := range imports {
		b.WriteString("\t" + imp + "\n")
	}
	b.WriteString(")\n\n")

	// ValidateAndBuildInsertParams
	fmt.Fprintf(&b, "func ValidateAndBuildInsertParams(userID uuid.UUID, %s *proto.%s) (*query.Insert%sParams, []pkg.ValidationError) {\n", goVarName, capitalizedModelName, capitalizedModelName)
	b.WriteString("\terrors := make([]pkg.ValidationError, 0)\n")

	// Per-column validations (insert)
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\tif %s.Get%s() == \"\" {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s is required\"})\n\t}\n", goVarName, field, c.Name, toFieldName(c.Name))
			fmt.Fprintf(&b, "\tif %s.Get%s() != \"\" && len(%s.Get%s()) < 3 {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"})\n\t}\n", goVarName, field, goVarName, field, c.Name, toFieldName(c.Name))
		case "number":
			v := toLocalVarName(field) + "Float"
			fmt.Fprintf(&b, "\t%s, err := strconv.ParseFloat(%s.Get%s(), 64)\n", v, goVarName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t} else if " + v + " < 1 {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		case "date":
			v := toLocalVarName(field)
			fmt.Fprintf(&b, "\t%s, err := str.ParseDate(%s.Get%s())\n", v, goVarName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		}
	}

	b.WriteString("\tif len(errors) > 0 {\n\t\treturn nil, errors\n\t}\n\n")

	// Build Insert params
	fmt.Fprintf(&b, "\treturn &query.Insert%sParams{\n", capitalizedModelName)
	b.WriteString("\t\tUserID: userID,\n")
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		case "number":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		case "date":
			v := toLocalVarName(field)
			fmt.Fprintf(&b, "\t\t%s: %s,\n", field, v)
		case "bool":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		}
	}
	b.WriteString("\t}, nil\n}\n\n")

	// ValidateAndBuildUpdateParams
	fmt.Fprintf(&b, "func ValidateAndBuildUpdateParams(userID uuid.UUID, %s *proto.%s) (*query.Update%sParams, []pkg.ValidationError) {\n", goVarName, capitalizedModelName, capitalizedModelName)
	b.WriteString("\terrors := make([]pkg.ValidationError, 0)\n")
	fmt.Fprintf(&b, "\tid, err := uuid.Parse(%s.GetId())\n", goVarName)
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\terrors = append(errors, pkg.ValidationError{Field: \"id\", Tag: \"uuid\", Message: \"ID must be a valid UUID\"})\n")
	b.WriteString("\t}\n")
	b.WriteString("\tif id == uuid.Nil {\n")
	b.WriteString("\t\terrors = append(errors, pkg.ValidationError{Field: \"id\", Tag: \"required\", Message: \"ID is required\"})\n")
	b.WriteString("\t}\n")

	// Per-column validations (update)
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\tif %s.Get%s() == \"\" {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s is required\"})\n\t}\n", goVarName, field, c.Name, toFieldName(c.Name))
			fmt.Fprintf(&b, "\tif %s.Get%s() != \"\" && len(%s.Get%s()) < 3 {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"})\n\t}\n", goVarName, field, goVarName, field, c.Name, toFieldName(c.Name))
		case "number":
			v := toLocalVarName(field) + "Float"
			fmt.Fprintf(&b, "\t%s, err := strconv.ParseFloat(%s.Get%s(), 64)\n", v, goVarName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t} else if " + v + " < 1 {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		case "date":
			v := toLocalVarName(field)
			fmt.Fprintf(&b, "\t%s, err := str.ParseDate(%s.Get%s())\n", v, goVarName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		}
	}

	b.WriteString("\tif len(errors) > 0 {\n\t\treturn nil, errors\n\t}\n\n")

	// Build Update params
	fmt.Fprintf(&b, "\treturn &query.Update%sParams{\n", capitalizedModelName)
	b.WriteString("\t\tID: id,\n")
	b.WriteString("\t\tUserID: userID,\n")
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		case "number":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		case "date":
			v := toLocalVarName(field)
			fmt.Fprintf(&b, "\t\t%s: %s,\n", field, v)
		case "bool":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, goVarName, field)
		}
	}
	b.WriteString("\t}, nil\n}\n")

	return b.String(), nil
}
