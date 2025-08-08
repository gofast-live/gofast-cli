package cmd

import (
	"fmt"
	"os"
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

var modelCmd = &cobra.Command{
	Use:   "model [model_name] [columns...]",
	Short: "Create a new model",
	Long: `Create a new model including database migrations, query generation, validation, API endpoints and UI views.

Columns are defined as name:type.

Valid column types are:
  - string  (PostgreSQL: text)
  - number  (PostgreSQL: numeric)
  - time    (PostgreSQL: timestamptz)
  - bool (PostgreSQL: bool)

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

		schemaContent, err := generateSchema(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating schema: %v.\n", err)
			return
		}
		err = appendToFile("./app/service-core/storage/schema.sql", schemaContent)
		if err != nil {
			cmd.Printf("Error writing schema file: %v.\n", err)
			return
		}

		queriesContent, err := generateQueries(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating queries: %v.\n", err)
			return
		}
		err = appendToFile("./app/service-core/storage/query.sql", queriesContent)
		if err != nil {
			cmd.Printf("Error writing query file: %v.\n", err)
			return
		}

		err = generateServiceLayer(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating service layer: %v.\n", err)
			return
		}

		cmd.Print("Model created successfully!\n")
		cmd.Printf("Model Name: %s\n", modelName)
		cmd.Printf("Columns:\n")
		for _, col := range columns {
			cmd.Printf("  - Name: %s, Type: %s\n", col.Name, col.Type)
		}
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

func generateSchema(modelName string, columns []Column) (string, error) {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(modelName)
	var columnDefs []string
	columnDefs = append(columnDefs, "    id uuid primary key default gen_random_uuid()")
	columnDefs = append(columnDefs, "    created timestamptz not null default current_timestamp")
	columnDefs = append(columnDefs, "    updated timestamptz not null default current_timestamp")

	typeMap := map[string]string{
		"string": "text",
		"number": "numeric",
		"time":   "timestamptz",
		"bool":   "bool",
	}

	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("    %s %s not null", col.Name, typeMap[col.Type]))
	}

	schemaContent := fmt.Sprintf(`

-- create \"%s\" table
create table if not exists %s (
%s
);`,
		tableName, tableName, strings.Join(columnDefs, ",\n"))

	return schemaContent, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func generateQueries(modelName string, columns []Column) (string, error) {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(modelName)
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
`, modelNamePlural, tableName, modelNameSingular, tableName, modelNameSingular, tableName, colNamesStr, placeholdersStr, modelNameSingular, tableName, updatePairsStr, len(columns)+1, modelNameSingular, tableName)

	return queries, nil
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
		if info.Name() == "dto.go" {
			newContentStr, genErr = generateDTO(modelName, columns)
		} else if info.Name() == "service.go" {
			newContentStr, genErr = generateServiceContent(modelName, capitalizedModelName, columns)
		} else if info.Name() == "service_test.go" {
			newContentStr, genErr = generateServiceTestContent(modelName, capitalizedModelName, columns)
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

func generateDTO(modelName string, columns []Column) (string, error) {
	capitalizedModelName := capitalize(modelName)
	var createFields, updateFields []string
	usesTime := false
	usesUUID := true // For the ID in the update DTO

	typeMap := map[string]string{
		"string": "string",
		"number": "string",
		"time":   "time.Time",
		"bool":   "bool",
	}

	for _, col := range columns {
		if col.Type == "time" {
			usesTime = true
		}
		goType := typeMap[col.Type]
		fieldName := toCamelCase(col.Name)
		jsonTag := col.Name
		createFields = append(createFields, fmt.Sprintf("\t%s %s `json:\"%s\" validate:\"required\"`", fieldName, goType, jsonTag))
		updateFields = append(updateFields, fmt.Sprintf("\t%s %s `json:\"%s\" validate:\"required\"`", fieldName, goType, jsonTag))
	}

	var content strings.Builder

	content.WriteString(fmt.Sprintf("package %s\n\n", modelName))

	imports := []string{}
	if usesTime {
		imports = append(imports, "\t\"time\"")
	}
	if usesUUID {
		imports = append(imports, "\t\"github.com/google/uuid\"")
	}

	if len(imports) > 0 {
		content.WriteString("import (\n")
		content.WriteString(strings.Join(imports, "\n"))
		content.WriteString("\n)")
		content.WriteString("\n\n")
	}

	createStructName := fmt.Sprintf("Insert%sDTO", capitalizedModelName)
	createStruct := fmt.Sprintf("type %s struct {\n%s\n}", createStructName, strings.Join(createFields, "\n"))

	updateStructName := fmt.Sprintf("Update%sDTO", capitalizedModelName)
	updateFieldsWithID := append([]string{"\tID    uuid.UUID `json:\"id\" validate:\"required\"`"}, updateFields...)
	updateStruct := fmt.Sprintf("type %s struct {\n%s\n}", updateStructName, strings.Join(updateFieldsWithID, "\n"))

	content.WriteString(createStruct)
	content.WriteString("\n\n")
	content.WriteString(updateStruct)
	content.WriteString("\n")

	return content.String(), nil
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
	insertParamsContent := "\t\t" + strings.Join(insertFieldParts, ",\n\t\t")

	var updateFieldParts []string
	updateFieldParts = append(updateFieldParts, "ID:    dto.ID")
	for _, col := range columns {
		fieldName := toCamelCase(col.Name)
		updateFieldParts = append(updateFieldParts, fmt.Sprintf("%s:  dto.%s", fieldName, fieldName))
	}
	updateParamsContent := "\t\t" + strings.Join(updateFieldParts, ",\n\t\t")

	oldInsertBlock := "\t\tName:  dto.Name,\n\t\tAge:   dto.Age,\n\t\tAlive: dto.Alive"
	content = strings.Replace(content, oldInsertBlock, insertParamsContent, 1)

	oldUpdateBlock := "\t\tID:    dto.ID,\n\t\tName:  dto.Name,\n\t\tAge:   dto.Age,\n\t\tAlive: dto.Alive"
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
	content := string(contentBytes)

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
			return fmt.Sprintf("%t", index%2 != 0)
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
			return "false"
		default:
			return "\"\""
		}
	}

	genFields := func(index int, empty bool) string {
		var fieldParts []string
		for _, col := range columns {
			fieldName := toCamelCase(col.Name)
			var val string
			if empty {
				val = getEmptyValue(col.Type)
			} else {
				val = getMockValue(col.Type, index, col.Name)
			}
			fieldParts = append(fieldParts, fmt.Sprintf("%s: %s", fieldName, val))
		}
		return strings.Join(fieldParts, ", ")
	}

	// --- Generate replacement field strings ---

	newQueryFields1 := genFields(1, false)
	newQueryFields2 := genFields(2, false)
	newInsertDtoFields := genFields(1, false)
	newEmptyInsertDtoFields := genFields(0, true)
	newUpdateDtoFields := genFields(2, false)

	// --- Perform replacements ---

	// TestService_GetAllSkeletons
	content = strings.Replace(content, `Name: "Skelly1", Age: "100", Alive: true`, newQueryFields1, 1)
	content = strings.Replace(content, `Name: "Skelly2", Age: "200", Alive: false`, newQueryFields2, 1)

	// TestService_GetAllSkeletons_ProcessError
	content = strings.Replace(content, `Name: "Skelly1", Age: "100", Alive: true`, newQueryFields1, 1)

	// TestService_GetSkeletonByID
	content = strings.Replace(content, `ID: id, Name: "Test Skeleton", Created: time.Now(), Updated: time.Now(), Age: "50", Alive: true`,
		"ID: id, Created: time.Now(), Updated: time.Now(), "+newQueryFields1, 1)

	// TestService_CreateSkeleton
	content = strings.ReplaceAll(content, `Name: "New Skeleton", Age: "1", Alive: true`, newInsertDtoFields, )
	content = strings.Replace(content, `ID: uuid.New(), Name: "New Skeleton", Created: time.Now(), Updated: time.Now(), Age: "1", Alive: true`,
		"ID: uuid.New(), Created: time.Now(), Updated: time.Now(), "+newInsertDtoFields, 1)

	// TestService_CreateSkeleton_ValidationError
	content = strings.Replace(content, `Name: "", Age: "", Alive: true`, newEmptyInsertDtoFields, 1)
	numValidationErrors := 0
	for _, col := range columns {
		if col.Type == "string" || col.Type == "number" {
			numValidationErrors++
		}
	}
	content = strings.Replace(content, `assert.Len(t, target, 2)`, fmt.Sprintf("assert.Len(t, target, %d)", numValidationErrors), 1)

	// TestService_EditSkeleton
	content = strings.ReplaceAll(content, `ID: uuid.New(), Name: "Updated Skeleton", Age: "2", Alive: false`, "ID: uuid.New(), "+newUpdateDtoFields, )
	content = strings.ReplaceAll(content, `ID: dto.ID, Name: "Updated Skeleton", Age: "2", Alive: false`, "ID: dto.ID, "+newUpdateDtoFields, )
	content = strings.Replace(content, `ID: dto.ID, Name: "Updated Skeleton", Created: time.Now(), Updated: time.Now(), Age: "2", Alive: false`,
		"ID: dto.ID, Created: time.Now(), Updated: time.Now(), "+newUpdateDtoFields, 1)

	// TestService_EditSkeleton_ValidationError
	content = strings.Replace(content, `ID: uuid.New(), Title: "", Content: "", Views: "", Done: false, Deadline: time.Time{}`, "ID: uuid.New(), "+newEmptyInsertDtoFields, 1)

	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

	return content, nil
}


