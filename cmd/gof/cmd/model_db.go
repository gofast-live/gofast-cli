package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

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
		"date":   "string",
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
	bufCmd := exec.Command("sh", "scripts/run_proto.sh")
	bufOut, err := bufCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running Buf script: %v\nOutput: %s", err, bufOut)
	}
	return nil
}

func generateSchema(modelName string, columns []Column) (string, error) {
	tableName := pluralizeClient.Plural(modelName)
	migrationsDir := "./app/service-core/storage/migrations"

	err := os.MkdirAll(migrationsDir, 0o755)
	if err != nil {
		return "", fmt.Errorf("creating migrations directory %s: %w", migrationsDir, err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("reading migrations directory %s: %w", migrationsDir, err)
	}

	var maxNumber int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		underscoreIndex := strings.Index(name, "_")
		if underscoreIndex == -1 {
			continue
		}

		numberPart := name[:underscoreIndex]
		parsedNumber, parseErr := strconv.Atoi(numberPart)
		if parseErr != nil {
			continue
		}

		if parsedNumber > maxNumber {
			maxNumber = parsedNumber
		}
	}

	nextNumber := maxNumber + 1
	migrationFileName := fmt.Sprintf("%05d_create_%s.sql", nextNumber, tableName)
	migrationPath := filepath.Join(migrationsDir, migrationFileName)

	_, err = os.Stat(migrationPath)
	if err == nil {
		return "", fmt.Errorf("migration file already exists: %s", migrationPath)
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking migration file %s: %w", migrationPath, err)
	}

	columnDefs := []string{
		"    id uuid primary key default gen_random_uuid()",
		"    created timestamptz not null default current_timestamp",
		"    updated timestamptz not null default current_timestamp",
		"    user_id uuid not null references users(id) on delete cascade",
	}

	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("    %s %s not null", col.Name, typeMap[col.Type]))
	}

	migrationContent := fmt.Sprintf(`-- +goose Up
-- create "%s" table
create table if not exists %s (
%s
);

-- +goose Down
drop table if exists %s;
`, tableName, tableName, strings.Join(columnDefs, ",\n"), tableName)

	err = os.WriteFile(migrationPath, []byte(migrationContent), 0o644)
	if err != nil {
		return "", fmt.Errorf("writing migration file %s: %w", migrationPath, err)
	}

	return migrationPath, nil
}

func generateQueries(modelName string, columns []Column) error {
	tableName := pluralizeClient.Plural(modelName)
	modelNameSingular := capitalize(modelName)
	modelNamePlural := capitalize(tableName)

	// For insert
	var insertColNames = []string{"user_id"}
	var placeholders = []string{"$1"}
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
