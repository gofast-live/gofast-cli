package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(modelCmd)
}

type Column struct {
	Name string
	Type string
}

var typeMap = map[string]string{
	"string": "text",
	"number": "numeric",
	"time":   "timestamptz",
	"bool":   "boolean",
}

var modelCmd = &cobra.Command{
	Use:   "model [model_name] [columns...]",
	Short: "Create a new model",
	Long: `Create a new model including database migrations, query generation, validation, API endpoints and UI views.

Columns are defined as name:type.

Valid column types are:
  - string  (PostgreSQL: text)
  - number  (PostgreSQL: numeric)
  - time    (PostgreSQL: timestamptz)
  - bool    (PostgreSQL: boolean)

Example:
  gof model post title:string content:string views:number published_at:time is_published:bool
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]
		columnStrings := args[1:]

		var columns []Column
		validTypes := map[string]bool{
			"string": true,
			"number": true,
			"time":   true,
			"bool":   true,
		}

		for _, colStr := range columnStrings {
			parts := strings.Split(colStr, ":")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				cmd.Printf("Error: Invalid column format '%s'. Use name:type.\n", colStr)
				return
			}

			colType := strings.ToLower(parts[1])
			if !validTypes[colType] {
				cmd.Printf("Error: Invalid type '%s' for column '%s'.\n", parts[1], parts[0])
				cmd.Println("Valid types are: string, number, time, bool.")
				return
			}

			columns = append(columns, Column{
				Name: parts[0],
				Type: colType,
			})
		}

		err := config.AddModel(modelName)
		if err != nil {
			cmd.Printf("Error adding model: %v.\n", err)
			return
		}

		err = generateProto(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating proto: %v.\n", err)
			return
		}

		err = generateSchema(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating schema: %v.\n", err)
			return
		}

		err = generateQueries(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating queries: %v.\n", err)
			return
		}

		err = generateServiceLayer(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating service layer: %v.\n", err)
			return
		}

		// err = generateRestLayer(modelName, columns)
		// if err != nil {
		// 	cmd.Printf("Error generating REST layer: %v.\n", err)
		// 	return
		// }

		// err = generateAPIEndpoints(modelName)
		// if err != nil {
		// 	cmd.Printf("Error adding model to transport and rest: %v.\n", err)
		// 	return
		// }

		cmdExec := exec.Command("sh", "scripts/run_sqlc.sh")
		output, err := cmdExec.CombinedOutput()
		if err != nil {
			cmd.Printf("Error running SQLC script: %v\nOutput: %s\n", err, output)
			return
		}

		cmd.Print("Model created successfully!\n")
		cmd.Printf("Model Name: %s\n", config.SuccessStyle.Render(modelName))
		cmd.Printf("Columns:\n")
		for _, col := range columns {
			cmd.Printf("  - Name: %s, Type: %v\n", col.Name, typeMap[col.Type])
		}

		cmd.Printf("\nSchema generated in: %s\n", config.SuccessStyle.Render("./app/service-core/storage/schema.sql"))
		cmd.Printf("Queries generated in: %s\n", config.SuccessStyle.Render("./app/service-core/storage/query.sql"))
		cmd.Printf("Service layer generated in: %s\n", config.SuccessStyle.Render("./app/service-core/domain/"+modelName))
		cmd.Printf("REST layer generated in: %s\n\n", config.SuccessStyle.Render("./app/service-core/transport/rest/"+modelName))

		cmd.Printf("Don't forget to run %s to apply migrations.\n", config.SuccessStyle.Render("scripts/run_atlas.sh"))

	},
}

func appendToFile(filePath, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

var pluralizeClient = pluralize.NewClient()

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func generateProto(modelName string, columns []Column) error {
	protoDir := "./proto/v1"

	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		return err
	}

	capitalizedModelName := capitalize(modelName)
	pluralModelName := pluralizeClient.Plural(modelName)

	typeMapProto := map[string]string{
		"string": "string",
		"number": "int64",
		"time":   "string",
		"bool":   "bool",
	}

	// 1) Create model proto file if missing
	modelProtoPath := filepath.Join(protoDir, modelName+".proto")
	if _, err := os.Stat(modelProtoPath); err != nil {
		var b strings.Builder
		b.WriteString("syntax = \"proto3\";\n")
		b.WriteString("option go_package = \"gofast/gen/proto/v1\";\n")
		b.WriteString("package proto.v1;\n\n")
		b.WriteString("message " + capitalizedModelName + " {\n")
		b.WriteString("    string id = 1;\n")
		b.WriteString("    string created = 2;\n")
		b.WriteString("    string updated = 3;\n\n")

		fieldNo := 4
		for _, col := range columns {
			ptype, ok := typeMapProto[col.Type]
			if !ok {
				ptype = "string"
			}
			b.WriteString(fmt.Sprintf("    %s %s = %d;\n", ptype, col.Name, fieldNo))
			fieldNo++
		}
		b.WriteString("}\n")

		if err := os.WriteFile(modelProtoPath, []byte(b.String()), 0o644); err != nil {
			return err
		}
	}

	// 2) Update main.proto: add import, messages, and service
	mainProtoPath := filepath.Join(protoDir, "main.proto")
	mainBytes, err := os.ReadFile(mainProtoPath)
	if err != nil {
		return err
	}
	mainContent := string(mainBytes)

	importLine := fmt.Sprintf("import \"proto/v1/%s.proto\";", modelName)
	if !strings.Contains(mainContent, importLine) {
		lines := strings.Split(mainContent, "\n")
		insertIdx := 0
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "import ") {
				insertIdx = i + 1
			}
		}
		if insertIdx == 0 {
			insertIdx = len(lines)
		}
		lines = append(lines[:insertIdx], append([]string{importLine}, lines[insertIdx:]...)...)
		mainContent = strings.Join(lines, "\n")
	}

	serviceMarker := fmt.Sprintf("service %sService", capitalizedModelName)
	if !strings.Contains(mainContent, serviceMarker) {
		var sb strings.Builder
		sb.WriteString("\n// --- " + capitalizedModelName + " Service ---\n\n")

		// Messages
		pluralCap := capitalize(pluralModelName)
		// GetAll
		sb.WriteString(fmt.Sprintf("// GetAll%s\n", pluralCap))
		sb.WriteString(fmt.Sprintf("message GetAll%sRequest {}\n", pluralCap))
		sb.WriteString(fmt.Sprintf("message GetAll%sResponse {\n", pluralCap))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n\n")

		// GetByID
		sb.WriteString(fmt.Sprintf("// Get%sByID\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("message Get%sByIDRequest {\n", capitalizedModelName))
		sb.WriteString("    string id = 1;\n")
		sb.WriteString("}\n")
		sb.WriteString(fmt.Sprintf("message Get%sByIDResponse {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n\n")

		// Create
		sb.WriteString(fmt.Sprintf("// Create%s\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("message Create%sRequest {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n")
		sb.WriteString(fmt.Sprintf("message Create%sResponse {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n\n")

		// Edit
		sb.WriteString(fmt.Sprintf("// Edit%s\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("message Edit%sRequest {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n")
		sb.WriteString(fmt.Sprintf("message Edit%sResponse {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    %s %s = 1;\n", capitalizedModelName, modelName))
		sb.WriteString("}\n\n")

		// Remove
		sb.WriteString(fmt.Sprintf("// Remove%s\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("message Remove%sRequest {\n", capitalizedModelName))
		sb.WriteString("    string id = 1;\n")
		sb.WriteString("}\n")
		sb.WriteString(fmt.Sprintf("message Remove%sResponse {}\n\n", capitalizedModelName))

		// Service
		sb.WriteString(fmt.Sprintf("service %sService {\n", capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    rpc GetAll%s(GetAll%sRequest) returns (stream GetAll%sResponse) {}\n", pluralCap, pluralCap, pluralCap))
		sb.WriteString(fmt.Sprintf("    rpc Get%sByID(Get%sByIDRequest) returns (Get%sByIDResponse) {}\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    rpc Create%s(Create%sRequest) returns (Create%sResponse) {}\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    rpc Edit%s(Edit%sRequest) returns (Edit%sResponse) {}\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
		sb.WriteString(fmt.Sprintf("    rpc Remove%s(Remove%sRequest) returns (Remove%sResponse) {}\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
		sb.WriteString("}\n")

		mainContent = mainContent + sb.String()
	}

	if err := os.WriteFile(mainProtoPath, []byte(mainContent), 0o644); err != nil {
		return err
	}

	// Generate protobuf stubs via Buf
	bufCmd := exec.Command("sh", "scripts/run_buf.sh")
	bufOut, err := bufCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running Buf script: %v\nOutput: %s", err, bufOut)
	}
	return nil
}

func generateSchema(modelName string, columns []Column) error {
	tableName := pluralizeClient.Plural(modelName)
	var columnDefs []string
	columnDefs = append(columnDefs, "    id uuid primary key default gen_random_uuid()")
	columnDefs = append(columnDefs, "    created timestamptz not null default current_timestamp")
	columnDefs = append(columnDefs, "    updated timestamptz not null default current_timestamp")

	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("    %s %s not null", col.Name, typeMap[col.Type]))
	}

	schemaContent := fmt.Sprintf(`
-- create "%s" table
create table if not exists %s (
%s
);`,
		tableName, tableName, strings.Join(columnDefs, ",\n"))

	err := appendToFile("./app/service-core/storage/schema.sql", schemaContent)
	if err != nil {
		return fmt.Errorf("appending to schema.sql: %w", err)
	}
	return nil
}

func generateQueries(modelName string, columns []Column) error {
	tableName := pluralizeClient.Plural(modelName)
	modelNameSingular := capitalize(modelName)
	modelNamePlural := capitalize(tableName)

	var colNames, placeholders, updatePairs []string
	for i, col := range columns {
		colNames = append(colNames, col.Name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		updatePairs = append(updatePairs, fmt.Sprintf("%s = $%d", col.Name, i+1))
	}
	colNamesStr := strings.Join(colNames, ", ")
	placeholdersStr := strings.Join(placeholders, ", ")
	updatePairsStr := strings.Join(updatePairs, ",\n    ")

	queries := fmt.Sprintf(`
-- %s --

-- name: SelectAll%s :many
select * from %s;

-- name: Select%sByID :one
select * from %s where id = $1;

-- name: Insert%s :one
insert into %s (%s) values (%s) returning *;

-- name: Update%s :one
update %s set
    %s,
    updated = current_timestamp
where id = $%d returning *;

-- name: Delete%s :exec
delete from %s where id = $1;
`, modelNamePlural, modelNamePlural, tableName, modelNameSingular, tableName, modelNameSingular, tableName, colNamesStr, placeholdersStr, modelNameSingular, tableName, updatePairsStr, len(columns)+1, modelNameSingular, tableName)

	err := appendToFile("./app/service-core/storage/query.sql", queries)
	if err != nil {
		return fmt.Errorf("appending to query.sql: %w", err)
	}
	return nil
}

func generateServiceLayer(modelName string, columns []Column) error {
	sourceDir := "./app/service-core/domain/skeleton"
	destDir := "app/service-core/domain/" + modelName
	capitalizedModelName := capitalize(modelName)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := strings.Replace(path, sourceDir, destDir, 1)
		destPath = strings.ReplaceAll(destPath, "skeleton", modelName)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		var newContentStr string
		var genErr error
		if info.Name() == "service.go" {
			newContentStr, genErr = generateServiceContent(modelName, capitalizedModelName, columns)
		} else if info.Name() == "service_test.go" {
			newContentStr, genErr = generateServiceTestContent(modelName, capitalizedModelName, columns)
		} else if info.Name() == "validation.go" {
			// TODO
		} else if info.Name() == "validation_test.go" {
			// TODO
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContentStr = strings.ReplaceAll(string(content), "skeleton", modelName)
			newContentStr = strings.ReplaceAll(newContentStr, "Skeleton", capitalizedModelName)
		}

		if genErr != nil {
			return fmt.Errorf("generating content for %s: %w", destPath, genErr)
		}

		return os.WriteFile(destPath, []byte(newContentStr), info.Mode())
	})
}

func generateServiceContent(modelName string, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}
	content := string(contentBytes)

	var insertFieldParts []string
	for _, col := range columns {
		fieldName := toCamelCase(col.Name)
		insertFieldParts = append(insertFieldParts, fmt.Sprintf("%s:  dto.%s", fieldName, fieldName))
	}
	insertParamsContent := "\t\t" + strings.Join(insertFieldParts, ",\n\t\t") + ","

	var updateFieldParts []string
	updateFieldParts = append(updateFieldParts, "ID:    dto.ID")
	for _, col := range columns {
		fieldName := toCamelCase(col.Name)
		updateFieldParts = append(updateFieldParts, fmt.Sprintf("%s:  dto.%s", fieldName, fieldName))
	}
	updateParamsContent := "\t\t" + strings.Join(updateFieldParts, ",\n\t\t") + ","

	oldInsertBlock := `		Name:   dto.Name,
		Age:    dto.Age,
		Death:  dto.Death,
		Zombie: dto.Zombie,`
	content = strings.Replace(content, oldInsertBlock, insertParamsContent, 1)

	oldUpdateBlock := `		ID:     dto.ID,
		Name:   dto.Name,
		Age:    dto.Age,
		Death:  dto.Death,
		Zombie: dto.Zombie,`
	content = strings.Replace(content, oldUpdateBlock, updateParamsContent, 1)

	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}

func generateServiceTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	// --- Field generation helpers ---
	getMockValue := func(colType string, index int, colName string) string {
		switch colType {
		case "string":
			return fmt.Sprintf("\"Test %s %d\"", toCamelCase(colName), index)
		case "number":
			return fmt.Sprintf("\"%d.0\"", 100*index)
		case "time":
			return "time.Now()"
		case "bool":
			// Use `true` for success-path tests to pass the "required" validation.
			return "true"
		default:
			return "\"\""
		}
	}

	getEmptyValue := func(colType string) string {
		switch colType {
		case "string", "number":
			return "\"\""
		case "time":
			return "time.Time{}"
		case "bool":
			// Use `false` for validation tests to trigger the "required" tag (zero value).
			return "false"
		default:
			return "\"\""
		}
	}

	genFields := func(index int, empty bool, fromDto bool) string {
		var fieldParts []string
		for _, col := range columns {
			fieldName := toCamelCase(col.Name)
			var val string
			if empty {
				val = getEmptyValue(col.Type)
			} else if fromDto {
				val = fmt.Sprintf("dto.%s", fieldName)
			} else {
				val = getMockValue(col.Type, index, col.Name)
			}
			fieldParts = append(fieldParts, fmt.Sprintf("%s: %s", fieldName, val))
		}
		return strings.Join(fieldParts, ", ")
	}

	getIndent := func(s string) string {
		for i, r := range s {
			if r != ' ' && r != '\t' {
				return s[:i]
			}
		}
		return ""
	}

	// --- Define replacements ---
	replacements := map[string][]string{
		"// QUERY": {
			"ID: uuid.New(), Created: time.Now(), Updated: time.Now(), " + genFields(1, false, false) + ",",
			"ID: uuid.New(), Created: time.Now(), Updated: time.Now(), " + genFields(2, false, false) + ",",
			"ID: uuid.New(), Created: time.Now(), Updated: time.Now(), " + genFields(1, false, false) + ",",
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(1, false, false) + ",",
			"ID: uuid.New(), Created: time.Now(), Updated: time.Now(), " + genFields(1, false, true) + ",",
		},
		"// INSERT DTO": {
			genFields(1, false, false) + ",",
			genFields(1, false, false) + ",",
		},
		"// EMPTY INSERT DTO": {
			genFields(0, true, false) + ",",
		},
		"// INSERT PARAMS": {
			genFields(1, false, true) + ",",
			genFields(1, false, true) + ",",
		},
		"// UPDATE DTO": {
			"ID: uuid.New(), " + genFields(2, false, false) + ",",
			"ID: uuid.New(), " + genFields(2, false, false) + ",",
		},
		"// EMPTY UPDATE DTO": {
			"ID: uuid.New(), " + genFields(0, true, false) + ",",
		},
		"// UPDATE PARAMS": {
			"ID: dto.ID, " + genFields(2, false, true) + ",",
			"ID: dto.ID, " + genFields(2, false, true) + ",",
		},
		"// QUERY PARAMS": {
			"ID: dto.ID, Created: time.Now(), Updated: time.Now(), " + genFields(2, false, true) + ",",
		},
	}
	counters := make(map[string]int)

	// --- Process lines ---
	lines := strings.Split(string(contentBytes), "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if replacementList, ok := replacements[trimmedLine]; ok {
			newLines = append(newLines, line) // Keep the comment line

			counter := counters[trimmedLine]
			if counter < len(replacementList) {
				replacement := replacementList[counter]
				if i+1 < len(lines) {
					indent := getIndent(lines[i+1])
					newLines = append(newLines, indent+replacement)
					i++ // Skip the next line from the template
				}
				counters[trimmedLine]++
			} else {
				if i+1 < len(lines) {
					newLines = append(newLines, lines[i+1])
					i++
				}
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	content := strings.Join(newLines, "\n")

	// --- Final replacements ---
	// The empty DTOs generate zero values for all fields, so all should fail "required" validation.
	// The number of errors should always be the number of columns.
	content = strings.Replace(content, "assert.Len(t, target, 3)", fmt.Sprintf("assert.Len(t, target, %d)", len(columns)), 1)
	content = strings.Replace(content, "assert.Len(t, target, 4)", fmt.Sprintf("assert.Len(t, target, %d)", len(columns)), 1)

	// Model names
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}

func generateRestLayer(modelName string, columns []Column) error {
	sourceDir := "./app/service-core/transport/rest/skeleton"
	destDir := "app/service-core/transport/rest/" + modelName
	capitalizedModelName := capitalize(modelName)
	pluralModelName := pluralizeClient.Plural(modelName)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := strings.Replace(path, sourceDir, destDir, 1)
		destPath = strings.ReplaceAll(destPath, "skeleton", modelName)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		var newContentStr string
		var genErr error
		if info.Name() == "route.go" {
			newContentStr, genErr = generateRouteContent(modelName, capitalizedModelName, pluralModelName)
		} else if info.Name() == "route_test.go" {
			newContentStr, genErr = generateRouteTestContent(modelName, capitalizedModelName, columns)
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContentStr = strings.ReplaceAll(string(content), "skeletons", pluralModelName)
			newContentStr = strings.ReplaceAll(newContentStr, "skeleton", modelName)
			newContentStr = strings.ReplaceAll(newContentStr, "Skeleton", capitalizedModelName)
		}

		if genErr != nil {
			return fmt.Errorf("generating content for %s: %w", destPath, genErr)
		}

		return os.WriteFile(destPath, []byte(newContentStr), info.Mode())
	})
}

func generateRouteContent(modelName, capitalizedModelName, pluralModelName string) (string, error) {
	templatePath := "./app/service-core/transport/rest/skeleton/route.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}
	content := string(contentBytes)

	content = strings.ReplaceAll(content, "skeletons", pluralModelName)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}

func generateRouteTestContent(modelName, capitalizedModelName string, columns []Column) (string, error) {
	templatePath := "./app/service-core/transport/rest/skeleton/route_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	// --- Field generation helpers ---
	getMockValue := func(colType string, index int, colName string) string {
		switch colType {
		case "string":
			return fmt.Sprintf("\"Test %s %d\"", toCamelCase(colName), index)
		case "number":
			return fmt.Sprintf("\"%d.0\"", 100*index)
		case "time":
			return "time.Now()"
		case "bool":
			return "true"
		default:
			return "\"\""
		}
	}

	genFields := func(index int, fromDto bool) string {
		var fieldParts []string
		for _, col := range columns {
			fieldName := toCamelCase(col.Name)
			var val string
			if fromDto {
				val = fmt.Sprintf("dto.%s", fieldName)
			} else {
				val = getMockValue(col.Type, index, col.Name)
			}
			fieldParts = append(fieldParts, fmt.Sprintf("%s: %s", fieldName, val))
		}
		return strings.Join(fieldParts, ", ")
	}

	getIndent := func(s string) string {
		for i, r := range s {
			if r != ' ' && r != '\t' {
				return s[:i]
			}
		}
		return ""
	}

	// --- Define replacements ---
	replacements := map[string][]string{
		"// QUERY": {
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(1, false) + ",",
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(1, false) + ",",
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(2, false) + ",",
		},
		"// INSERT DTO": {
			genFields(1, false) + ",",
			genFields(1, false) + ",",
		},
		"// UPDATE DTO": {
			"ID: id, " + genFields(2, false) + ",",
			"ID: id, " + genFields(2, false) + ",",
		},
		"// QUERY DTO": {
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(1, true) + ",",
			"ID: id, Created: time.Now(), Updated: time.Now(), " + genFields(2, true) + ",",
		},
	}
	counters := make(map[string]int)

	// --- Process lines ---
	lines := strings.Split(string(contentBytes), "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if replacementList, ok := replacements[trimmedLine]; ok {
			newLines = append(newLines, line) // Keep the comment line

			counter := counters[trimmedLine]
			if counter < len(replacementList) {
				replacement := replacementList[counter]
				if i+1 < len(lines) {
					indent := getIndent(lines[i+1])
					newLines = append(newLines, indent+replacement)
					i++ // Skip the next line from the template
				}
				counters[trimmedLine]++
			} else {
				if i+1 < len(lines) {
					newLines = append(newLines, lines[i+1])
					i++
				}
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	content := strings.Join(newLines, "\n")

	// --- Smart assertion replacement ---
	var assertColName string
	for _, col := range columns {
		if col.Type == "string" {
			assertColName = toCamelCase(col.Name)
			break
		}
	}

	if assertColName != "" {
		// TestRegisterRoutes and TestGetAllSkeletons use mocks with index 1
		assertValue1 := fmt.Sprintf("Test %s 1", assertColName)
		content = strings.Replace(content, `assert.Contains(t, rr.Body.String(), "Skelly")`, fmt.Sprintf(`assert.Contains(t, rr.Body.String(), "%s")`, assertValue1), 2)

		// TestGetSkeletonByID and TestEditSkeleton use mocks with index 2
		assertValue2 := fmt.Sprintf("Test %s 2", assertColName)
		content = strings.Replace(content, `assert.Contains(t, rr.Body.String(), "Skelly")`, fmt.Sprintf(`assert.Contains(t, rr.Body.String(), "%s")`, assertValue2), 2)
	} else {
		// Fallback for models without string fields: check for the ID.
		content = strings.ReplaceAll(content, `assert.Contains(t, rr.Body.String(), "Skelly")`, `assert.Contains(t, rr.Body.String(), id.String())`)
	}

	pluralModelName := pluralizeClient.Plural(modelName)

	// Final replacements
	content = strings.ReplaceAll(content, "skeletons", pluralModelName)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}

func generateAPIEndpoints(modelName string) error {
	capitalizedModelName := capitalize(modelName)
	pluralModelName := pluralizeClient.Plural(modelName)

	// =================================================================
	// Modify ./app/service-core/transport/transport.go
	// =================================================================
	transportGoPath := "./app/service-core/transport/transport.go"
	transportGoContentBytes, err := os.ReadFile(transportGoPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", transportGoPath, err)
	}
	transportGoContent := string(transportGoContentBytes)

	// Add import
	importHook := `import (`
	importAddition := fmt.Sprintf("\n\t%[1]sSvc \"service-core/domain/%[1]s\"", modelName)
	transportGoContent = strings.Replace(transportGoContent, importHook, importHook+importAddition, 1)

	// Add service to Transport struct
	structHook := `SkeletonService *skeletonSvc.Service`
	structAddition := fmt.Sprintf("\n\t%sService *%sSvc.Service", capitalizedModelName, modelName)
	transportGoContent = strings.Replace(transportGoContent, structHook, structHook+structAddition, 1)

	// Instantiate service in New()
	newFuncHook := `skeletonService := skeletonSvc.NewService(cfg, store)`
	newFuncAddition := fmt.Sprintf("\n\t%sService := %sSvc.NewService(cfg, store)", modelName, modelName)
	transportGoContent = strings.Replace(transportGoContent, newFuncHook, newFuncHook+newFuncAddition, 1)

	// Pass service to rest.NewHandler
	restHandlerHook := `rest.NewHandler(cfg, store, authService, skeletonService)`
	restHandlerReplacement := fmt.Sprintf(`rest.NewHandler(cfg, store, authService, skeletonService, %sService)`, modelName)
	transportGoContent = strings.Replace(transportGoContent, restHandlerHook, restHandlerReplacement, 1)

	// Add service to returned struct
	returnHook := `SkeletonService: skeletonService,`
	returnAddition := fmt.Sprintf("\n\t\t%sService: %sService,", capitalizedModelName, modelName)
	transportGoContent = strings.Replace(transportGoContent, returnHook, returnHook+returnAddition, 1)

	err = os.WriteFile(transportGoPath, []byte(transportGoContent), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", transportGoPath, err)
	}

	// =================================================================
	// Modify ./app/service-core/transport/rest/server.go
	// =================================================================
	serverGoPath := "./app/service-core/transport/rest/server.go"
	serverGoContentBytes, err := os.ReadFile(serverGoPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", serverGoPath, err)
	}
	serverGoContent := string(serverGoContentBytes)

	// Add imports
	serverImportHook := `import (`
	serverImportAddition := fmt.Sprintf("\n\t%[1]sSvc \"service-core/domain/%[1]s\"\n\t%[1]sRoute \"service-core/transport/rest/%[1]s\"", modelName)
	serverGoContent = strings.Replace(serverGoContent, serverImportHook, serverImportHook+serverImportAddition, 1)

	// Add service to Handler struct
	serverStructHook := `skeletonService *skeletonSvc.Service`
	serverStructAddition := fmt.Sprintf("\n\t%sService *%sSvc.Service", modelName, modelName)
	serverGoContent = strings.Replace(serverGoContent, serverStructHook, serverStructHook+serverStructAddition, 1)

	// Add service to NewHandler signature
	newHandlerSigHook := `skeletonService *skeletonSvc.Service,`
	newHandlerSigAddition := fmt.Sprintf("\n\t%sService *%sSvc.Service,", modelName, modelName)
	serverGoContent = strings.Replace(serverGoContent, newHandlerSigHook, newHandlerSigHook+newHandlerSigAddition, 1)

	// Add service to Handler literal
	newHandlerBodyHook := `skeletonService: skeletonService,`
	newHandlerBodyAddition := fmt.Sprintf("\n\t\t%sService: %sService,", modelName, modelName)
	serverGoContent = strings.Replace(serverGoContent, newHandlerBodyHook, newHandlerBodyHook+newHandlerBodyAddition, 1)

	// Add routes in NewServer
	serverRoutesHook := `mux.Handle("/skeleton/", h.Authn(skeletonMux))`
	serverRoutesAddition := fmt.Sprintf("\n\n\t// %s\n\t%sMux := http.NewServeMux()\n\t%sHandler := %sRoute.NewHandler(h.%sService)\n\t%sRoute.RegisterRoutes(%sMux, %sHandler)\n\tmux.Handle(\"/%s/\", h.Authn(%sMux))",
		capitalize(pluralModelName),
		modelName,
		modelName, modelName, modelName,
		modelName, modelName, modelName,
		pluralModelName, modelName)
	serverGoContent = strings.Replace(serverGoContent, serverRoutesHook, serverRoutesHook+serverRoutesAddition, 1)

	err = os.WriteFile(serverGoPath, []byte(serverGoContent), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", serverGoPath, err)
	}

	// =================================================================
	// Modify ./app/service-core/transport/rest/server_test.go
	// =================================================================
	serverTestGoPath := "./app/service-core/transport/rest/server_test.go"
	serverTestGoContentBytes, err := os.ReadFile(serverTestGoPath)
	if err != nil {
		// If the file doesn't exist, we don't need to do anything.
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading %s: %w", serverTestGoPath, err)
	}
	serverTestGoContent := string(serverTestGoContentBytes)

	lines := strings.Split(serverTestGoContent, "\n")
	var newLines []string
	for _, line := range lines {
		if strings.Contains(line, "rest.NewHandler(") {
			line = strings.Replace(line, ")", ", nil)", 1)
		}
		newLines = append(newLines, line)
	}
	newServerTestGoContent := strings.Join(newLines, "\n")

	err = os.WriteFile(serverTestGoPath, []byte(newServerTestGoContent), 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", serverTestGoPath, err)
	}

	return nil
}
