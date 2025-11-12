package cmd

import (
	"fmt"
	"os"
	"strings"
)

func generateServiceTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	// Helpers to generate field lists
	toFieldName := func(col string) string { return toCamelCase(col) }
	toVarName := func(camel string) string {
		if camel == "" {
			return camel
		}
		return strings.ToLower(camel[:1]) + camel[1:]
	}
	mockQueryVal := func(colType string, index int) string {
		switch colType {
		case "string":
			return fmt.Sprintf("\"Test %d\"", index)
		case "number":
			// sqlc maps numeric to string by default
			return fmt.Sprintf("\"%d\"", 100*index)
		case "date":
			return "time.Now()"
		case "bool":
			if index%2 == 0 {
				return "false"
			}
			return "true"
		default:
			return "\"\""
		}
	}
	zeroQueryVal := func(colType string) string {
		switch colType {
		case "string", "number":
			return "\"\""
		case "date":
			return "time.Time{}"
		case "bool":
			return "false"
		default:
			return "\"\""
		}
	}
	// proto value helpers: use stable values that match params helpers
	protoVal := func(colType string, isEdit bool) string {
		switch colType {
		case "string":
			if isEdit {
				return "\"Updated\""
			}
			return "\"Test\""
		case "number":
			if isEdit {
				return "\"200\""
			}
			return "\"100\""
		case "date":
			return "\"2023-10-01\""
		case "bool":
			return "true"
		default:
			return "\"\""
		}
	}
	zeroProtoVal := func(colType string) string {
		switch colType {
		case "string":
			return "\"\""
		case "number":
			return "\"\""
		case "date":
			return "\"\""
		case "bool":
			return "false"
		default:
			return "\"\""
		}
	}

	buildQueryFields := func(index int, zero bool) string {
		parts := []string{
			"ID: uuid.New()",
			"UserID: userID",
			"Created: time.Now()",
			"Updated: time.Now()",
		}
		if zero {
			parts = []string{
				"ID: uuid.Nil",
				"UserID: uuid.Nil",
				"Created: time.Time{}",
				"Updated: time.Time{}",
			}
		}
		for _, c := range columns {
			name := toFieldName(c.Name)
			if zero {
				parts = append(parts, fmt.Sprintf("%s: %s", name, zeroQueryVal(c.Type)))
			} else {
				parts = append(parts, fmt.Sprintf("%s: %s", name, mockQueryVal(c.Type, index)))
			}
		}
		return strings.Join(parts, ",\n")
	}
	buildQueryFieldsWithI := func(zero bool) string {
		parts := []string{
			"ID: uuid.New()",
			"UserID: userID",
			"Created: time.Now()",
			"Updated: time.Now()",
		}
		if zero {
			parts = []string{
				"ID: uuid.Nil",
				"UserID: uuid.Nil",
				"Created: time.Time{}",
				"Updated: time.Time{}",
			}
		}
		for _, c := range columns {
			name := toFieldName(c.Name)
			if zero {
				parts = append(parts, fmt.Sprintf("%s: %s", name, zeroQueryVal(c.Type)))
				continue
			}
			switch c.Type {
			case "string":
				parts = append(parts, fmt.Sprintf("%s: fmt.Sprintf(\"Test %s\", i)", name, "%d"))
			case "number":
				parts = append(parts, fmt.Sprintf("%s: \"100\"", name))
			case "date":
				parts = append(parts, fmt.Sprintf("%s: time.Now()", name))
			case "bool":
				parts = append(parts, fmt.Sprintf("%s: i%%2 == 1", name))
			default:
				parts = append(parts, fmt.Sprintf("%s: \"\"", name))
			}
		}
		return strings.Join(parts, ",\n")
	}
	buildInsertParams := func() (pre string, fields string) {
		parts := []string{}
		predecl := []string{}
		for _, c := range columns {
			name := toFieldName(c.Name)
			switch c.Type {
			case "string":
				parts = append(parts, fmt.Sprintf("%s: \"Test\"", name))
			case "number":
				parts = append(parts, fmt.Sprintf("%s: \"100\"", name))
			case "date":
				v := toVarName(name)
				predecl = append(predecl, fmt.Sprintf("%s, _ := time.Parse(\"2006-01-02\", \"2023-10-01\")", v))
				parts = append(parts, fmt.Sprintf("%s: %s", name, v))
			case "bool":
				parts = append(parts, fmt.Sprintf("%s: true", name))
			default:
				parts = append(parts, fmt.Sprintf("%s: \"\"", name))
			}
		}
		return strings.Join(predecl, "\n\t"), strings.Join(parts, ", ")
	}
	buildUpdateParams := func() (pre string, fields string) {
		parts := []string{"ID: id"}
		predecl := []string{}
		for _, c := range columns {
			name := toFieldName(c.Name)
			switch c.Type {
			case "string":
				parts = append(parts, fmt.Sprintf("%s: \"Updated\"", name))
			case "number":
				parts = append(parts, fmt.Sprintf("%s: \"200\"", name))
			case "date":
				v := toVarName(name)
				predecl = append(predecl, fmt.Sprintf("%s, _ := time.Parse(\"2006-01-02\", \"2023-10-01\")", v))
				parts = append(parts, fmt.Sprintf("%s: %s", name, v))
			case "bool":
				parts = append(parts, fmt.Sprintf("%s: true", name))
			default:
				parts = append(parts, fmt.Sprintf("%s: \"\"", name))
			}
		}
		return strings.Join(predecl, "\n\t"), strings.Join(parts, ", ")
	}
	buildProtoFields := func(_ int, zero bool, useEditID bool) string {
		parts := []string{}
		if useEditID {
			parts = append(parts, "Id: id.String()")
		} else {
			parts = append(parts, "Id: \"\"")
		}
		parts = append(parts, "Created: \"\"")
		parts = append(parts, "Updated: \"\"")
		for _, c := range columns {
			name := toFieldName(c.Name)
			if zero {
				parts = append(parts, fmt.Sprintf("%s: %s", name, zeroProtoVal(c.Type)))
			} else {
				parts = append(parts, fmt.Sprintf("%s: %s", name, protoVal(c.Type, useEditID)))
			}
		}
		return strings.Join(parts, ",\n\t\t\t\t")
	}

	lines := strings.Split(string(contentBytes), "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		switch strings.TrimSpace(trimmed) {
		case "// GF_FIXTURES_START":
			out = append(out, line)
			indent := strings.Repeat("\t", strings.Count(line, "\t"))

			out = append(out, indent+fmt.Sprintf("func makeQuery%s(i int, userID uuid.UUID) query.%s {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn query."+capitalizedModelName+"{")
			fields := buildQueryFieldsWithI(false)
			fields = strings.ReplaceAll(fields, "\n", "\n"+indent+"\t\t")
			out = append(out, indent+"\t\t"+fields+",")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func zeroQuery%s() query.%s {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn query."+capitalizedModelName+"{")
			zfields := buildQueryFields(1, true)
			zfields = strings.ReplaceAll(zfields, "\n", "\n"+indent+"\t\t")
			out = append(out, indent+"\t\t"+zfields+",")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeInsert%sParams(userID uuid.UUID) query.Insert%sParams {", capitalizedModelName, capitalizedModelName))
			pre, fieldsIns := buildInsertParams()
			if strings.TrimSpace(pre) != "" {
				for pl := range strings.SplitSeq(pre, "\n") {
					out = append(out, indent+"\t"+pl)
				}
			}
			out = append(out, indent+"\treturn query.Insert"+capitalizedModelName+"Params{")
			out = append(out, indent+"\t\tUserID: userID,")
			out = append(out, indent+"\t\t"+fieldsIns+",")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeUpdate%sParams(id uuid.UUID, userID uuid.UUID) query.Update%sParams {", capitalizedModelName, capitalizedModelName))
			preU, fieldsUpd := buildUpdateParams()
			if strings.TrimSpace(preU) != "" {
				for pl := range strings.SplitSeq(preU, "\n") {
					out = append(out, indent+"\t"+pl)
				}
			}
			out = append(out, indent+"\treturn query.Update"+capitalizedModelName+"Params{")
			out = append(out, indent+"\t\t"+fieldsUpd+",")
			out = append(out, indent+"\t\tUserID: userID,")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeCreate%sReq() *proto.Create%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Create"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(1, false, false)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeEdit%sReq(id uuid.UUID) *proto.Edit%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Edit"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(1, false, true)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			// Zero/Invalid proto helpers
			out = append(out, indent+fmt.Sprintf("func makeZeroCreate%sReq() *proto.Create%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Create"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(0, true, false)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeZeroEdit%sReq() *proto.Edit%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Edit"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(0, true, false)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeInvalidEdit%sReq(id uuid.UUID) *proto.Edit%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Edit"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(0, true, true)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")

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
	// Replace identifiers
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)
	return content, nil
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
