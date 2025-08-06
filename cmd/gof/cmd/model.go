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

		cmd.Print("Model created successfully!\n")
		cmd.Printf("Model Name: %s\n", modelName)
		cmd.Printf("Columns:\n")
		for _, col := range columns {
			cmd.Printf("  - Name: %s, Type: %s\n", col.Name, col.Type)
		}
	},
}

func generateSchema(modelName string, columns []Column) error {
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

	filePath := "./app/service-core/storage/schema.sql"
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

	_, err = f.WriteString(schemaContent)
	if err != nil {
		return err
	}
	return nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func generateQueries(modelName string, columns []Column) error {
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

	filePath := "./app/service-core/storage/query.sql"
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
	if _, err := f.WriteString(queries); err != nil {
		return err
	}
	return nil
}

func generateServiceLayer(modelName string, columns []Column) error {
	sourceDir := "./app/service-core/domain/skeleton"
	destDir := "app/service-core/domain/" + modelName
	capitalizedModelName := capitalize(modelName)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := strings.Replace(path, sourceDir, destDir, 1)
		destPath = strings.ReplaceAll(destPath, "skeleton", modelName)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		var newContentStr string
		if info.Name() == "dto.go" {
			newContentStr = generateDTO(modelName, columns)
		} else if info.Name() == "service_test.go" {
			newContentStr = generateServiceTestContent(modelName, capitalizedModelName, columns)
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContentStr = strings.ReplaceAll(string(content), "skeleton", modelName)
			newContentStr = strings.ReplaceAll(newContentStr, "Skeleton", capitalizedModelName)
		}

		return os.WriteFile(destPath, []byte(newContentStr), info.Mode())
	})

	return err
}

func generateDTO(modelName string, columns []Column) string {
	capitalizedModelName := capitalize(modelName)
	var createFields, updateFields []string
	usesTime := false
	usesUUID := true // For the ID in the update DTO

	typeMap := map[string]string{
		"string": "string",
		"number": "float64",
		"time":   "time.Time",
		"bool":   "bool",
	}

	for _, col := range columns {
		if col.Type == "time" {
			usesTime = true
		}
		goType := typeMap[col.Type]
		fieldName := capitalize(col.Name)
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

	return content.String()
}

func generateServiceTestContent(modelName, capitalizedModelName string, columns []Column) string {
	// TODO
	return ""
}
