package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/svelte"
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
		_, _, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}

		// Ensure we are inside a valid gofast project (has gofast.json)
		if _, err := config.ParseConfig(); err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		modelName := args[0]
		columnStrings := args[1:]

		var columns []Column
		seenNames := map[string]bool{}
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

			// Ensure column names are unique
			if seenNames[parts[0]] {
				cmd.Printf("Error: Duplicate column name '%s'. Column names must be unique.\n", parts[0])
				return
			}
			seenNames[parts[0]] = true

			columns = append(columns, Column{
				Name: parts[0],
				Type: colType,
			})
		}

		configColumns := make([]config.Column, len(columns))
		for i, col := range columns {
			configColumns[i] = config.Column{
				Name: col.Name,
				Type: col.Type,
			}
		}
		err = config.AddModel(modelName, configColumns)
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

		// Add model-specific auth permissions before generating service layer
		err = generateAuthAccessFlags(modelName)
		if err != nil {
			cmd.Printf("Error updating auth permissions: %v.\n", err)
			return
		}

		err = generateServiceLayer(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating service layer: %v.\n", err)
			return
		}

		// Wire new model into core transport server (imports, handler fields, routes)
		err = wireCoreTransportServer(modelName)
		if err != nil {
			cmd.Printf("Error wiring core transport server: %v.\n", err)
			return
		}

		// Wire new model into main.go (imports, service init, handler args)
		err = wireCoreMain(modelName)
		if err != nil {
			cmd.Printf("Error wiring core main.go: %v.\n", err)
			return
		}

		// Generate ConnectRPC transport layer from skeleton template
		err = generateTransportLayer(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating transport layer: %v.\n", err)
			return
		}

		if config.IsSvelte() {
			svelteColumns := make([]svelte.Column, len(columns))
			for i, col := range columns {
				svelteColumns[i] = svelte.Column{
					Name: col.Name,
					Type: col.Type,
				}
			}
			err = svelte.GenerateSvelteScaffolding(modelName, svelteColumns)
			if err != nil {
				cmd.Printf("Error generating Svelte client pages: %v.\n", err)
				return
			}
		}

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

		cmd.Printf("\nProtobuf definitions generated in: %s\n", config.SuccessStyle.Render("./proto/v1/"+modelName+".proto"))
		cmd.Printf("Schema generated in: %s\n", config.SuccessStyle.Render("./app/service-core/storage/schema.sql"))
		cmd.Printf("Queries generated in: %s\n", config.SuccessStyle.Render("./app/service-core/storage/query.sql"))
		cmd.Printf("Service layer generated in: %s\n", config.SuccessStyle.Render("./app/service-core/domain/"+modelName))
		cmd.Printf("Transport layer generated in: %s\n", config.SuccessStyle.Render("./app/service-core/transport/"+modelName))
		if config.IsSvelte() {
			cmd.Printf("Client pages generated in: %s\n", config.SuccessStyle.Render("./app/client/src/pages/"+pluralizeClient.Plural(modelName)))
		}

		cmd.Printf("\nDon't forget to run %s to apply migrations.\n", config.SuccessStyle.Render("scripts/run_atlas.sh"))

		cmd.Printf("\nIf you already created a user, remember to update permissions for the new model.\n")
		cmd.Printf("\nYou can also run %s to update all users with admin permissions.\n\n", config.SuccessStyle.Render("scripts/update_permissions.sh"))
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
		"number": "string",
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
	columnDefs = append(columnDefs, "    user_id uuid not null references users(id) on delete cascade")

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

	// For insert
	var insertColNames []string = []string{"user_id"}
	var placeholders []string = []string{"$1"}
	for i, col := range columns {
		insertColNames = append(insertColNames, col.Name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
	}
	insertColNamesStr := strings.Join(insertColNames, ", ")
	placeholdersStr := strings.Join(placeholders, ", ")

	// For update
	var updatePairs []string
	for i, col := range columns {
		updatePairs = append(updatePairs, fmt.Sprintf("%s = $%d", col.Name, i+1))
	}
	updatePairsStr := strings.Join(updatePairs, ",\n    ")

	queries := fmt.Sprintf(`
-- %s --

-- name: SelectAll%s :many
select * from %s where user_id = $1 order by created desc;

-- name: Select%sByID :one
select * from %s where id = $1 and user_id = $2;

-- name: Insert%s :one
insert into %s (%s) values (%s) returning *;

-- name: Update%s :one
update %s set
    %s,
    updated = current_timestamp
where id = $%d and user_id = $%d returning *;

-- name: Delete%s :exec
delete from %s where id = $1 and user_id = $2;
`, modelNamePlural, modelNamePlural, tableName, modelNameSingular, tableName, modelNameSingular, tableName, insertColNamesStr, placeholdersStr, modelNameSingular, tableName, updatePairsStr, len(columns)+1, len(columns)+2, modelNameSingular, tableName)

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
			newContentStr = strings.ReplaceAll(string(content), "skeleton", modelName)
			newContentStr = strings.ReplaceAll(newContentStr, "Skeleton", capitalizedModelName)
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
	destDir := "app/service-core/transport/" + modelName

	capitalizedModelName := capitalize(modelName)
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)

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
			s = strings.ReplaceAll(s, "Skeletons", pluralCap)
			s = strings.ReplaceAll(s, "skeletons", pluralLower)
			s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
			s = strings.ReplaceAll(s, "skeleton", modelName)
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
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)
	// Rename leftover template-local variable names
	s = strings.ReplaceAll(s, "skelProto", modelName+"Proto")

	// Build dynamic queryToProto mapping based on columns
	// Numeric columns are strings in proto; no strconv needed

	// Construct struct fields
	var b strings.Builder
	b.WriteString("\t\tId:      " + modelName + ".ID.String(),\n")
	b.WriteString("\t\tCreated: " + modelName + ".Created.Format(time.RFC3339),\n")
	b.WriteString("\t\tUpdated: " + modelName + ".Updated.Format(time.RFC3339),\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			b.WriteString("\t\t" + field + ": " + modelName + "." + field + ",\n")
		case "time":
			b.WriteString("\t\t" + field + ": " + modelName + "." + field + ".Format(time.RFC3339),\n")
		case "bool":
			b.WriteString("\t\t" + field + ": " + modelName + "." + field + ",\n")
		case "number":
			// sqlc numeric is string; proto is also string
			b.WriteString("\t\t" + field + ": " + modelName + "." + field + ",\n")
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
	newFn := "func queryToProto(" + modelName + " *query." + capitalizedModelName + ") *proto." + capitalizedModelName + " {\n\treturn &proto." + capitalizedModelName + "{\n" + fields + "\t}\n}\n"
	s = s[:fnStart] + newFn + s[end:]
	return s, nil
}

func generateTransportTestContent(modelName, capitalizedModelName string, columns []Column, pluralLower, pluralCap string) (string, error) {
	templatePath := "./app/service-core/transport/skeleton/route_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	// Helpers similar to service test generation
	toFieldName := func(col string) string { return toCamelCase(col) }
	zeroQueryVal := func(colType string) string {
		switch colType {
		case "string", "number":
			return "\"\""
		case "time":
			return "time.Time{}"
		case "bool":
			return "false"
		default:
			return "\"\""
		}
	}
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
		case "time":
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
		case "time":
			return "\"\""
		case "bool":
			return "false"
		default:
			return "\"\""
		}
	}

	buildQueryFieldsWithI := func(zero bool) string {
		parts := []string{
			"ID: uuid.New()",
			"Created: time.Now()",
			"Updated: time.Now()",
		}
		if zero {
			parts = []string{
				"ID: uuid.Nil",
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
				// Generate: <Field>: fmt.Sprintf("Test %d", i)
				parts = append(parts, fmt.Sprintf("%s: fmt.Sprintf(\"Test %s\", i)", name, "%d"))
			case "number":
				// sqlc maps numeric to string; use a quoted literal
				parts = append(parts, fmt.Sprintf("%s: \"100\"", name))
			case "time":
				parts = append(parts, fmt.Sprintf("%s: time.Now()", name))
			case "bool":
				parts = append(parts, fmt.Sprintf("%s: i%%2 == 1", name))
			default:
				parts = append(parts, fmt.Sprintf("%s: \"\"", name))
			}
		}
		return strings.Join(parts, ",\n")
	}

	buildProtoFields := func(zero bool, useEditID bool) string {
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
		switch trimmed {
		case "// GF_FIXTURES_START":
			out = append(out, line)
			indent := strings.Repeat("\t", strings.Count(line, "\t"))

			out = append(out, indent+fmt.Sprintf("func makeQuery%s(i int) query.%s {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn query."+capitalizedModelName+"{")
			fields := buildQueryFieldsWithI(false)
			fields = strings.ReplaceAll(fields, "\n", "\n"+indent+"\t\t")
			out = append(out, indent+"\t\t"+fields+",")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeQuery%sPtr(i int) *query.%s {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\tv := makeQuery"+capitalizedModelName+"(i)")
			out = append(out, indent+"\treturn &v")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeCreate%sReq() *proto.Create%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Create"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(false, false)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeEdit%sReq(id uuid.UUID) *proto.Edit%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Edit"+capitalizedModelName+"Request{")
			out = append(out, indent+"\t\t"+capitalizedModelName+": &proto."+capitalizedModelName+"{")
			out = append(out, indent+"\t\t\t"+buildProtoFields(false, true)+",")
			out = append(out, indent+"\t\t},")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			// Minimal request helpers used by tests

			// Nil-skeleton helpers
			out = append(out, indent+fmt.Sprintf("func makeNilCreate%sReq() *proto.Create%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Create"+capitalizedModelName+"Request{ "+capitalizedModelName+": nil }")
			out = append(out, indent+"}")
			out = append(out, "")
			out = append(out, indent+fmt.Sprintf("func makeNilEdit%sReq() *proto.Edit%sRequest {", capitalizedModelName, capitalizedModelName))
			out = append(out, indent+"\treturn &proto.Edit"+capitalizedModelName+"Request{ "+capitalizedModelName+": nil }")
			out = append(out, indent+"}")

			// skip existing until END
			for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "// GF_FIXTURES_END" {
				i++
			}
		default:
			out = append(out, line)
		}
	}
	content := strings.Join(out, "\n")
	// Replace identifiers
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "skeletons", pluralLower)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	// Adjust template-local variable names
	content = strings.ReplaceAll(content, "protoSkel", "proto"+capitalizedModelName)

	// Column-aware test adjustments
	// Map first occurrence per type for assertions/modifications
	var firstStr, firstNum, firstTime, firstBool string
	for _, c := range columns {
		switch c.Type {
		case "string":
			if firstStr == "" {
				firstStr = toFieldName(c.Name)
			}
		case "number":
			if firstNum == "" {
				firstNum = toFieldName(c.Name)
			}
		case "time":
			if firstTime == "" {
				firstTime = toFieldName(c.Name)
			}
		case "bool":
			if firstBool == "" {
				firstBool = toFieldName(c.Name)
			}
		}
	}

	// Replace Name usages with the first string field
	if firstStr != "" {
		content = strings.ReplaceAll(content, ".GetName()", ".Get"+firstStr+"()")
		content = strings.ReplaceAll(content, ".Name", "."+firstStr)
	} else {
		// Remove lines that assert on Name when no string fields exist
		filtered := []string{}
		for ln := range strings.SplitSeq(content, "\n") {
			t := strings.TrimSpace(ln)
			if strings.Contains(t, ".GetName()") || strings.Contains(t, ".Name") {
				continue
			}
			filtered = append(filtered, ln)
		}
		content = strings.Join(filtered, "\n")
	}

	// Remove or adapt numeric/time/bool asserts that come from skeleton template
	if firstNum != "" {
		content = strings.ReplaceAll(content, ".GetAge()", ".Get"+firstNum+"()")
		content = strings.ReplaceAll(content, ".Age", "."+firstNum)
	} else {
		filtered := []string{}
		for ln := range strings.SplitSeq(content, "\n") {
			t := strings.TrimSpace(ln)
			if strings.Contains(t, ".GetAge()") || strings.Contains(t, ".Age") {
				continue
			}
			filtered = append(filtered, ln)
		}
		content = strings.Join(filtered, "\n")
	}

	if firstTime != "" {
		content = strings.ReplaceAll(content, ".GetDeath()", ".Get"+firstTime+"()")
		content = strings.ReplaceAll(content, ".Death", "."+firstTime)
	} else {
		filtered := []string{}
		for ln := range strings.SplitSeq(content, "\n") {
			t := strings.TrimSpace(ln)
			if strings.Contains(t, ".GetDeath()") || strings.Contains(t, ".Death") {
				continue
			}
			filtered = append(filtered, ln)
		}
		content = strings.Join(filtered, "\n")
		// Also drop the unused local variable `death := ...` from the template
		filtered2 := []string{}
		for ln := range strings.SplitSeq(content, "\n") {
			t := strings.TrimSpace(ln)
			if strings.HasPrefix(t, "death :=") || strings.HasPrefix(t, "death:=") {
				continue
			}
			filtered2 = append(filtered2, ln)
		}
		content = strings.Join(filtered2, "\n")
	}

	if firstBool != "" {
		content = strings.ReplaceAll(content, ".GetZombie()", ".Get"+firstBool+"()")
		content = strings.ReplaceAll(content, ".Zombie", "."+firstBool)
	} else {
		filtered := []string{}
		for ln := range strings.SplitSeq(content, "\n") {
			t := strings.TrimSpace(ln)
			if strings.Contains(t, ".GetZombie()") || strings.Contains(t, ".Zombie") {
				continue
			}
			filtered = append(filtered, ln)
		}
		content = strings.Join(filtered, "\n")
	}

	return content, nil
}

func generateServiceContent(modelName string, capitalizedModelName string) (string, error) {
	templatePath := "./app/service-core/domain/skeleton/service.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}
	content := string(contentBytes)

	// Template already follows ConnectRPC and builds params via validation helpers.
	// We only need to rename occurrences of the model token.
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
		case "time":
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
		case "time":
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
		case "time":
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
		case "time":
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
			"Created: time.Now()",
			"Updated: time.Now()",
		}
		if zero {
			parts = []string{
				"ID: uuid.Nil",
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
			"Created: time.Now()",
			"Updated: time.Now()",
		}
		if zero {
			parts = []string{
				"ID: uuid.Nil",
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
			case "time":
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
			case "time":
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
			case "time":
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

			out = append(out, indent+fmt.Sprintf("func makeQuery%s(i int) query.%s {", capitalizedModelName, capitalizedModelName))
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

			out = append(out, indent+fmt.Sprintf("func makeInsert%sParams() query.Insert%sParams {", capitalizedModelName, capitalizedModelName))
			pre, fieldsIns := buildInsertParams()
			if strings.TrimSpace(pre) != "" {
				for pl := range strings.SplitSeq(pre, "\n") {
					out = append(out, indent+"\t"+pl)
				}
			}
			out = append(out, indent+"\treturn query.Insert"+capitalizedModelName+"Params{")
			out = append(out, indent+"\t\t"+fieldsIns+",")
			out = append(out, indent+"\t}")
			out = append(out, indent+"}")
			out = append(out, "")

			out = append(out, indent+fmt.Sprintf("func makeUpdate%sParams(id uuid.UUID) query.Update%sParams {", capitalizedModelName, capitalizedModelName))
			preU, fieldsUpd := buildUpdateParams()
			if strings.TrimSpace(preU) != "" {
				for pl := range strings.SplitSeq(preU, "\n") {
					out = append(out, indent+"\t"+pl)
				}
			}
			out = append(out, indent+"\treturn query.Update"+capitalizedModelName+"Params{")
			out = append(out, indent+"\t\t"+fieldsUpd+",")
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

func generateValidationContent(modelName string, capitalizedModelName string, columns []Column) (string, error) {
	// Determine which imports are needed based on column types
	needStrconv := false
	needStr := false
	for _, c := range columns {
		switch c.Type {
		case "number":
			needStrconv = true
		case "time":
			needStr = true
		}
	}

	// Build imports
	imports := []string{
		"\"gofast/pkg\"",
		"\"gofast/service-core/storage/query\"",
		"\n\tproto \"gofast/gen/proto/v1\"",
		"\n\t\"github.com/google/uuid\"",
	}
	if needStr {
		imports = append([]string{"\"gofast/pkg/str\""}, imports...)
	}
	if needStrconv {
		imports = append(imports, "\n\t\"strconv\"")
	}

	// Helpers
	toFieldName := func(name string) string { return toCamelCase(name) }
	toVarName := func(camel string) string {
		if camel == "" {
			return camel
		}
		return strings.ToLower(camel[:1]) + camel[1:]
	}

	// Begin file content
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", modelName)
	b.WriteString("import (\n\t")
	b.WriteString(strings.Join(imports, "\n\t"))
	b.WriteString("\n)\n\n")

	// ValidateAndBuildInsertParams
	fmt.Fprintf(&b, "func ValidateAndBuildInsertParams(%s *proto.%s) (*query.Insert%sParams, error) {\n", modelName, capitalizedModelName, capitalizedModelName)
	b.WriteString("\terrors := make([]pkg.ValidationError, 0)\n")

	// Per-column validations (insert)
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\tif %s.Get%s() == \"\" {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s is required\"})\n\t}\n", modelName, field, c.Name, toFieldName(c.Name))
			fmt.Fprintf(&b, "\tif %s.Get%s() != \"\" && len(%s.Get%s()) < 3 {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"})\n\t}\n", modelName, field, modelName, field, c.Name, toFieldName(c.Name))
		case "number":
			v := toVarName(field) + "Float"
			fmt.Fprintf(&b, "\t%s, err := strconv.ParseFloat(%s.Get%s(), 64)\n", v, modelName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t} else if " + v + " < 1 {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		case "time":
			v := toVarName(field)
			fmt.Fprintf(&b, "\t%s, err := str.ParseDate(%s.Get%s())\n", v, modelName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		}
	}

	b.WriteString("\tif len(errors) > 0 {\n\t\treturn nil, pkg.ValidationErrors(errors)\n\t}\n\n")

	// Build Insert params
	fmt.Fprintf(&b, "\treturn &query.Insert%sParams{\n", capitalizedModelName)
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		case "number":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		case "time":
			v := toVarName(field)
			fmt.Fprintf(&b, "\t\t%s: %s,\n", field, v)
		case "bool":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		}
	}
	b.WriteString("\t}, nil\n}\n\n")

	// ValidateAndBuildUpdateParams
	fmt.Fprintf(&b, "func ValidateAndBuildUpdateParams(%s *proto.%s) (*query.Update%sParams, error) {\n", modelName, capitalizedModelName, capitalizedModelName)
	b.WriteString("\terrors := make([]pkg.ValidationError, 0)\n")
	fmt.Fprintf(&b, "\tid, err := uuid.Parse(%s.GetId())\n", modelName)
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
			fmt.Fprintf(&b, "\tif %s.Get%s() == \"\" {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s is required\"})\n\t}\n", modelName, field, c.Name, toFieldName(c.Name))
			fmt.Fprintf(&b, "\tif %s.Get%s() != \"\" && len(%s.Get%s()) < 3 {\n\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"})\n\t}\n", modelName, field, modelName, field, c.Name, toFieldName(c.Name))
		case "number":
			v := toVarName(field) + "Float"
			fmt.Fprintf(&b, "\t%s, err := strconv.ParseFloat(%s.Get%s(), 64)\n", v, modelName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"number\", Message: \"%s must be a number\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t} else if " + v + " < 1 {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		case "time":
			v := toVarName(field)
			fmt.Fprintf(&b, "\t%s, err := str.ParseDate(%s.Get%s())\n", v, modelName, field)
			b.WriteString("\tif err != nil {\n")
			fmt.Fprintf(&b, "\t\terrors = append(errors, pkg.ValidationError{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"})\n", c.Name, toFieldName(c.Name))
			b.WriteString("\t}\n")
		}
	}

	b.WriteString("\tif len(errors) > 0 {\n\t\treturn nil, pkg.ValidationErrors(errors)\n\t}\n\n")

	// Build Update params
	fmt.Fprintf(&b, "\treturn &query.Update%sParams{\n", capitalizedModelName)
	b.WriteString("\t\tID: id,\n")
	for _, c := range columns {
		field := toFieldName(c.Name)
		switch c.Type {
		case "string":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		case "number":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		case "time":
			v := toVarName(field)
			fmt.Fprintf(&b, "\t\t%s: %s,\n", field, v)
		case "bool":
			fmt.Fprintf(&b, "\t\t%s: %s.Get%s(),\n", field, modelName, field)
		}
	}
	b.WriteString("\t}, nil\n}\n")

	return b.String(), nil
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
		case "time":
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
			case "time":
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
	insertHeader := "\ttestCases := []struct {\n\t\tname          string\n\t\t" + modelName + "      *proto." + capitalizedModelName + "\n\t\texpectError   bool\n\t\texpectedError pkg.ValidationErrors\n\t}{\n"

	var insertCases strings.Builder
	// Valid case (bools true)
	fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError:   false,\n\t\t\texpectedError: nil,\n\t\t},\n", modelName, modelName, capitalizedModelName, strings.Join(buildValidArgs(true), ", "))

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
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"%s too short\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "number":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"0\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"%s less than 1\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "time":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"invalid-date\""
				}
			}
			fmt.Fprintf(&insertCases, "\t\t{\n\t\t\tname: \"invalid %s date\",\n\t\t\t%s: makeCreate%sProto(%s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		}
	}
	insertFooter := "\t}\n"

	// Update testCases generation
	updateHeader := "\ttestCases := []struct {\n\t\tname          string\n\t\t" + modelName + "      *proto." + capitalizedModelName + "\n\t\texpectError   bool\n\t\texpectedError pkg.ValidationErrors\n\t}{\n"
	var updateCases strings.Builder
	// Valid case
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"valid %s\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError:   false,\n\t\t\texpectedError: nil,\n\t\t},\n", modelName, modelName, capitalizedModelName, strings.Join(buildValidArgs(true), ", "))
	// invalid uuid case -> expect two errors
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"invalid uuid\",\n\t\t\t%s: makeEdit%sProto(\"invalid-uuid\", %s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"id\", Tag: \"uuid\", Message: \"ID must be a valid UUID\"},\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", modelName, capitalizedModelName, strings.Join(buildValidArgs(false), ", "))
	// nil uuid case -> required only
	fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"nil uuid\",\n\t\t\t%s: makeEdit%sProto(uuid.Nil.String(), %s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"id\", Tag: \"required\", Message: \"ID is required\"},\n\t\t\t},\n\t\t},\n", modelName, capitalizedModelName, strings.Join(buildValidArgs(false), ", "))
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
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"%s too short\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"minlength\", Message: \"%s must be at least 3 characters long\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "number":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"0\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"%s less than 1\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"gte\", Message: \"%s must be greater than or equal to 1\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
		case "time":
			for i := range columns {
				if columns[i].Name == c.Name {
					args[i] = "\"invalid-date\""
				}
			}
			fmt.Fprintf(&updateCases, "\t\t{\n\t\t\tname: \"invalid %s date\",\n\t\t\t%s: makeEdit%sProto(uuid.New().String(), %s),\n\t\t\texpectError: true,\n\t\t\texpectedError: pkg.ValidationErrors{\n\t\t\t\t{Field: \"%s\", Tag: \"required\", Message: \"%s date is required and must be in YYYY-MM-DD or RFC3339 format\"},\n\t\t\t},\n\t\t},\n", c.Name, modelName, capitalizedModelName, strings.Join(args, ", "), c.Name, fieldCamel)
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

func generateAuthAccessFlags(modelName string) error {
	path := "./app/pkg/auth/auth.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading auth file %s: %w", path, err)
	}
	content := string(contentBytes)

	// Build new flags and access list entries
	modelCap := capitalize(modelName)
	modelPlural := pluralizeClient.Plural(modelName)
	modelPluralCap := capitalize(modelPlural)

	flagsSnippet := fmt.Sprintf("\n\tGet%[1]s   int64 = 1 << iota\n\tCreate%[2]s int64 = 1 << iota\n\tEdit%[2]s   int64 = 1 << iota\n\tRemove%[2]s int64 = 1 << iota\n", modelPluralCap, modelCap)
	userListSnippet := fmt.Sprintf("Get%[1]s | Create%[2]s | Edit%[2]s | Remove%[2]s", modelPluralCap, modelCap)

	// Insert flags inside GF_ACCESS_FLAGS markers (before END) unless already present
	const flagsStart = "GF_ACCESS_FLAGS_START"
	const flagsEnd = "GF_ACCESS_FLAGS_END"
	{
		s := strings.Index(content, flagsStart)
		e := strings.Index(content, flagsEnd)
		if s == -1 || e == -1 || e <= s {
			return fmt.Errorf("auth markers for access flags not found")
		}
		// Compute exact region bounds between end of START line and start of END line
		startLineEndRel := strings.Index(content[s:], "\n")
		if startLineEndRel == -1 {
			return fmt.Errorf("cannot locate end of start marker line")
		}
		regionStart := s + startLineEndRel + 1
		endLineStart := strings.LastIndex(content[:e], "\n") + 1

		region := content[regionStart:endLineStart]
		// Clean out stray blank comment-only lines
		cleanedLines := make([]string, 0)
		for ln := range strings.SplitSeq(region, "\n") {
			t := strings.TrimSpace(ln)
			if t == "" || t == "//" {
				continue
			}
			cleanedLines = append(cleanedLines, ln)
		}
		region = strings.Join(cleanedLines, "\n")

		// Append flags only if not present already
		if !strings.Contains(region, "Create"+modelCap) && !strings.Contains(region, "Get"+modelPluralCap) {
			// Ensure region ends with a newline if not empty
			if region != "" && !strings.HasSuffix(region, "\n") {
				region += "\n"
			}
			region += strings.TrimPrefix(flagsSnippet, "\n")
		}
		content = content[:regionStart] + region + content[endLineStart:]
	}

	// Helper to append to a list region while ignoring commented placeholders
	appendToRegion := func(c, startMarker, endMarker, addition string) (string, error) {
		s := strings.Index(c, startMarker)
		e := strings.Index(c, endMarker)
		if s == -1 || e == -1 || e <= s {
			return c, fmt.Errorf("auth markers %s/%s not found", startMarker, endMarker)
		}
		// Compute region bounds: after START line to start of END line
		startLineEndRel := strings.Index(c[s:], "\n")
		if startLineEndRel == -1 {
			return c, fmt.Errorf("cannot locate end of start marker line for %s", startMarker)
		}
		regionStart := s + startLineEndRel + 1
		endLineStart := strings.LastIndex(c[:e], "\n") + 1
		region := c[regionStart:endLineStart]

		// Parse existing non-comment lines (each line is a group for a model)
		tokens := []string{}
		for ln := range strings.SplitSeq(region, "\n") {
			t := strings.TrimSpace(ln)
			if t == "" || strings.HasPrefix(t, "//") {
				continue
			}
			t = strings.TrimSuffix(t, "|")
			t = strings.TrimSpace(t)
			if after, ok := strings.CutPrefix(t, "|"); ok {
				t = strings.TrimSpace(after)
			}
			if t != "" {
				tokens = append(tokens, t)
			}
		}
		// Add the new addition if not already present
		joined := strings.Join(tokens, " ")
		if !strings.Contains(joined, addition) {
			tokens = append(tokens, addition)
		}

		// One model group per line; each line except the last ends with a trailing '|'
		rebuilt := ""
		if len(tokens) > 0 {
			for i, t := range tokens {
				line := "\t" + t
				if i < len(tokens)-1 {
					line += " |"
				}
				rebuilt += line + "\n"
			}
		}
		return c[:regionStart] + rebuilt + c[endLineStart:], nil
	}

	var uErr error
	content, uErr = appendToRegion(content, "GF_USER_ACCESS_START", "GF_USER_ACCESS_END", userListSnippet)
	if uErr != nil {
		return uErr
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing auth file: %w", err)
	}
	return nil

}

// wireCoreTransportServer injects imports, handler fields, constructor params/assignments,
// and route registrations for a new model into the core transport server using
// explicit marker regions. It expects markers to already exist in
// ./app/service-core/transport/server.go and will append entries if they are not present.
func wireCoreTransportServer(modelName string) error {
	path := "./app/service-core/transport/server.go"
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading transport server: %w", err)
	}
	s := string(b)

	cap := capitalize(modelName)
	svcAlias := modelName + "Svc"
	routeAlias := modelName + "Route"

	// Build snippets
	svcImport := "\t" + svcAlias + " \"gofast/service-core/domain/" + modelName + "\""
	routeImport := "\t" + routeAlias + " \"gofast/service-core/transport/" + modelName + "\""
	handlerField := "\t" + modelName + "Service *" + svcAlias + ".Service"
	param := "\t" + modelName + "Service *" + svcAlias + ".Service,"
	assign := "\t\t" + modelName + "Service: " + modelName + "Service,"
	route := strings.Join([]string{
		"\t" + modelName + "Server := " + routeAlias + ".New" + cap + "Server(h." + modelName + "Service)",
		"\tpath, handler = v1connect.New" + cap + "ServiceHandler(" + modelName + "Server, interceptors)",
		"\tmux.Handle(path, withCORS(h.cfg, handler))",
	}, "\n")

	// Helper: append a line into a region delimited by start/end markers.
	appendUniqueLine := func(content, startMarker, endMarker, line string) (string, error) {
		sidx := strings.Index(content, startMarker)
		eidx := strings.Index(content, endMarker)
		if sidx == -1 || eidx == -1 || eidx <= sidx {
			return content, fmt.Errorf("markers %q..%q not found", startMarker, endMarker)
		}
		// Find region bounds: after start line to start of end line
		// Move to end of start line
		startLineEndRel := strings.Index(content[sidx:], "\n")
		if startLineEndRel == -1 {
			return content, fmt.Errorf("cannot locate end of start marker line for %s", startMarker)
		}
		regionStart := sidx + startLineEndRel + 1
		endLineStart := strings.LastIndex(content[:eidx], "\n") + 1
		region := content[regionStart:endLineStart]
		// Check if line (trimmed) already present (loose contains)
		if strings.Contains(region, strings.TrimSpace(line)) {
			return content, nil
		}
		if region != "" && !strings.HasSuffix(region, "\n") {
			region += "\n"
		}
		if !strings.HasSuffix(line, "\n") {
			line += "\n"
		}
		region += line
		return content[:regionStart] + region + content[endLineStart:], nil
	}

	var aerr error
	s, aerr = appendUniqueLine(s, "GF_TP_IMPORT_SERVICES_START", "GF_TP_IMPORT_SERVICES_END", svcImport)
	if aerr != nil {
		return fmt.Errorf("adding service import: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_TP_IMPORT_ROUTES_START", "GF_TP_IMPORT_ROUTES_END", routeImport)
	if aerr != nil {
		return fmt.Errorf("adding route import: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_TP_HANDLER_FIELDS_START", "GF_TP_HANDLER_FIELDS_END", handlerField)
	if aerr != nil {
		return fmt.Errorf("adding handler field: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_TP_HANDLER_PARAMS_START", "GF_TP_HANDLER_PARAMS_END", param)
	if aerr != nil {
		return fmt.Errorf("adding handler param: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_TP_HANDLER_ASSIGN_START", "GF_TP_HANDLER_ASSIGN_END", assign)
	if aerr != nil {
		return fmt.Errorf("adding handler assignment: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_TP_ROUTES_START", "GF_TP_ROUTES_END", route)
	if aerr != nil {
		return fmt.Errorf("adding route registration: %w", aerr)
	}

	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing transport server: %w", err)
	}
	return nil
}

// wireCoreMain injects imports, service initialization, and handler arguments for
// a new model into ./app/service-core/main.go using marker regions.
func wireCoreMain(modelName string) error {
	path := "./app/service-core/main.go"
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading core main.go: %w", err)
	}
	s := string(b)

	svcAlias := modelName + "Svc"
	importLine := "\t" + svcAlias + " \"gofast/service-core/domain/" + modelName + "\""
	initLine := "\t" + modelName + "Service := " + svcAlias + ".NewService(cfg, store, authService)"
	argLine := "\t" + modelName + "Service,"

	// Append unique lines to regions
	appendUniqueLine := func(content, startMarker, endMarker, line string) (string, error) {
		sidx := strings.Index(content, startMarker)
		eidx := strings.Index(content, endMarker)
		if sidx == -1 || eidx == -1 || eidx <= sidx {
			return content, fmt.Errorf("markers %q..%q not found", startMarker, endMarker)
		}
		// Find region bounds: after start line to start of end line
		startLineEndRel := strings.Index(content[sidx:], "\n")
		if startLineEndRel == -1 {
			return content, fmt.Errorf("cannot locate end of start marker line for %s", startMarker)
		}
		regionStart := sidx + startLineEndRel + 1
		endLineStart := strings.LastIndex(content[:eidx], "\n") + 1
		region := content[regionStart:endLineStart]
		if strings.Contains(region, strings.TrimSpace(line)) {
			return content, nil
		}
		if region != "" && !strings.HasSuffix(region, "\n") {
			region += "\n"
		}
		if !strings.HasSuffix(line, "\n") {
			line += "\n"
		}
		region += line
		return content[:regionStart] + region + content[endLineStart:], nil
	}

	var aerr error
	s, aerr = appendUniqueLine(s, "GF_MAIN_IMPORT_SERVICES_START", "GF_MAIN_IMPORT_SERVICES_END", importLine)
	if aerr != nil {
		return fmt.Errorf("adding main import: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_MAIN_INIT_SERVICES_START", "GF_MAIN_INIT_SERVICES_END", initLine)
	if aerr != nil {
		return fmt.Errorf("adding main init service: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_MAIN_HANDLER_ARGS_START", "GF_MAIN_HANDLER_ARGS_END", argLine)
	if aerr != nil {
		return fmt.Errorf("adding main handler arg: %w", aerr)
	}

	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing core main.go: %w", err)
	}
	return nil
}
