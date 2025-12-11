package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
	"date":   "timestamptz",
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
  - date    (PostgreSQL: timestamptz)
  - bool    (PostgreSQL: boolean)

Example:
  gof model post title:string content:string views:number published_at:date is_published:bool
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

		// Validate model name: must be lowercase letters and underscores only
		validModelName := regexp.MustCompile(`^[a-z][a-z_]*$`)
		if !validModelName.MatchString(modelName) {
			cmd.Println("Error: Invalid model name. Must start with a lowercase letter and contain only lowercase letters and underscores.")
			cmd.Println("Example: gof model note title:string content:string")
			return
		}

		// Reject plural model names to avoid generation issues
		if pluralizeClient.IsPlural(modelName) {
			singular := pluralizeClient.Singular(modelName)
			cmd.Printf("Error: Model name '%s' appears to be plural. Use the singular form instead.\n", modelName)
			cmd.Printf("Suggestion: gof model %s ...\n", singular)
			return
		}

		columnStrings := args[1:]

		var columns []Column
		seenNames := map[string]bool{}
		validTypes := map[string]bool{
			"string": true,
			"number": true,
			"date":   true,
			"bool":   true,
		}

		// Reserved column names that conflict with auto-generated fields
		reservedColumns := map[string]bool{
			"id": true, "user_id": true, "created": true, "updated": true,
		}

		// Go reserved keywords that would cause compilation errors
		goKeywords := map[string]bool{
			"break": true, "case": true, "chan": true, "const": true, "continue": true,
			"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
			"func": true, "go": true, "goto": true, "if": true, "import": true,
			"interface": true, "map": true, "package": true, "range": true, "return": true,
			"select": true, "struct": true, "switch": true, "type": true, "var": true,
		}

		// Column name format: same as model name (lowercase + underscores)
		validColName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

		var counter int
		for _, colStr := range columnStrings {
			parts := strings.Split(colStr, ":")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				cmd.Printf("Error: Invalid column format '%s'. Use name:type.\n", colStr)
				return
			}

			colName := parts[0]

			// Validate column name format
			if !validColName.MatchString(colName) {
				cmd.Printf("Error: Invalid column name '%s'. Must start with a lowercase letter and contain only lowercase letters, numbers, and underscores.\n", colName)
				return
			}

			// Check for reserved column names
			if reservedColumns[colName] {
				cmd.Printf("Error: Column name '%s' is reserved (auto-generated). Choose a different name.\n", colName)
				return
			}

			// Check for Go keywords
			if goKeywords[colName] {
				cmd.Printf("Error: Column name '%s' is a Go reserved keyword. Choose a different name.\n", colName)
				return
			}

			colType := strings.ToLower(parts[1])
			if !validTypes[colType] {
				cmd.Printf("Error: Invalid type '%s' for column '%s'.\n", parts[1], colName)
				cmd.Println("Valid types are: string, number, date, bool.")
				return
			}

			// Ensure column names are unique
			if seenNames[colName] {
				cmd.Printf("Error: Duplicate column name '%s'. Column names must be unique.\n", colName)
				return
			}
			seenNames[colName] = true

			columns = append(columns, Column{
				Name: colName,
				Type: colType,
			})
			counter++
		}

		// min 2 columns
		if counter < 2 {
			cmd.Printf("Error: At least 2 columns are required, got %d.\n", counter)
			return
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

		migrationPath, err := generateSchema(modelName, columns)
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

		// Generate ConnectRPC transport layer from skeleton template
		err = generateTransportLayer(modelName, columns)
		if err != nil {
			cmd.Printf("Error generating transport layer: %v.\n", err)
			return
		}

		// Wire new model into main.go (imports, deps init, route mounting)
		err = wireCoreMain(modelName)
		if err != nil {
			cmd.Printf("Error wiring core main.go: %v.\n", err)
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

		cmdExec := exec.Command("sh", "scripts/run_queries.sh")
		output, err := cmdExec.CombinedOutput()
		if err != nil {
			cmd.Printf("Error running scripts/run_queries.sh: %v\nOutput: %s\n", err, output)
			return
		}

		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = "app"
		gofmtOutput, err := gofmtCmd.CombinedOutput()
		if err != nil {
			cmd.Printf("Error running go fmt in ./app: %v\nOutput: %s\n", err, gofmtOutput)
			return
		}

		cmd.Println("")
		cmd.Print("Model created successfully!\n")
		cmd.Println("")
		cmd.Printf("Model Name: %s\n", config.SuccessStyle.Render(modelName))
		cmd.Printf("Columns:\n")
		for _, col := range columns {
			cmd.Printf("  - Name: %s, Type: %v\n", col.Name, typeMap[col.Type])
		}

		cmd.Printf("\nProtobuf definitions generated in: %s\n", config.SuccessStyle.Render("proto/v1/"+modelName+".proto"))
		cmd.Printf("Migration generated in: %s\n", config.SuccessStyle.Render(migrationPath))
		cmd.Printf("Queries generated in: %s\n", config.SuccessStyle.Render("app/service-core/storage/query.sql"))
		cmd.Printf("Service layer generated in: %s\n", config.SuccessStyle.Render("app/service-core/domain/"+modelName))
		cmd.Printf("Transport layer generated in: %s\n", config.SuccessStyle.Render("app/service-core/transport/"+modelName))
		if config.IsSvelte() {
			cmd.Printf("Client pages generated in: %s\n", config.SuccessStyle.Render("app/client/src/pages/"+pluralizeClient.Plural(modelName)))
		}

		cmd.Printf("\nDon't forget to run %s to apply migrations.\n", config.SuccessStyle.Render("scripts/run_migrations.sh"))

		cmd.Printf("\nIf you already created a user, remember to update permissions for the new model (check %s).\n", config.SuccessStyle.Render("pkg/auth/auth.go"))
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
	return toCamelCase(s)
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

// generateTransportTestContent generates transport test file by copying skeleton and replacing markers
func generateTransportTestContent(modelName, capitalizedModelName string, columns []Column, pluralLower, pluralCap string) (string, error) {
	templatePath := "./app/service-core/transport/skeleton/route_test.go"
	contentBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	content := string(contentBytes)

	// Build replacement content for each marker type (reuse helpers from model_test_gen.go)
	entityFields := buildEntityFields(columns)
	createFields := buildCreateProtoFields(columns, capitalizedModelName)
	editFields := buildEditProtoFields(columns)

	// Replace marker regions
	content = replaceMarkerRegion(content, "GF_TP_TEST_ENTITY_FIELDS_START", "GF_TP_TEST_ENTITY_FIELDS_END", entityFields)
	content = replaceMarkerRegion(content, "GF_TP_TEST_CREATE_FIELDS_START", "GF_TP_TEST_CREATE_FIELDS_END", createFields)
	content = replaceMarkerRegion(content, "GF_TP_TEST_EDIT_FIELDS_START", "GF_TP_TEST_EDIT_FIELDS_END", editFields)

	// Token replacement
	content = strings.ReplaceAll(content, "Skeletons", pluralCap)
	content = strings.ReplaceAll(content, "skeletons", pluralLower)
	content = strings.ReplaceAll(content, "skeleton", modelName)
	content = strings.ReplaceAll(content, "Skeleton", capitalizedModelName)

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

// wireCoreMain injects imports, deps initialization, and route mounting for
// a new model into ./app/service-core/main.go using marker regions.
func wireCoreMain(modelName string) error {
	path := "./app/service-core/main.go"
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading core main.go: %w", err)
	}
	s := string(b)

	cap := capitalize(modelName)
	svcAlias := modelName + "Svc"
	routeAlias := modelName + "Route"

	// Import lines
	svcImportLine := "\t" + svcAlias + " \"gofast/service-core/domain/" + modelName + "\""
	routeImportLine := "\t" + routeAlias + " \"gofast/service-core/transport/" + modelName + "\""

	// Deps initialization
	depsInitLine := "\t" + modelName + "Deps := " + svcAlias + ".Deps{Store: store}"

	// Route mounting (3 lines)
	routeMountLines := strings.Join([]string{
		"\t" + modelName + "Server := " + routeAlias + ".New" + cap + "Server(" + modelName + "Deps)",
		"\tpath, handler = v1connect.New" + cap + "ServiceHandler(" + modelName + "Server, server.Interceptors())",
		"\tserver.Mount(path, handler)",
	}, "\n")

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
		if strings.Contains(region, strings.TrimSpace(strings.Split(line, "\n")[0])) {
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
	s, aerr = appendUniqueLine(s, "GF_MAIN_IMPORT_SERVICES_START", "GF_MAIN_IMPORT_SERVICES_END", svcImportLine)
	if aerr != nil {
		return fmt.Errorf("adding service import: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_MAIN_IMPORT_ROUTES_START", "GF_MAIN_IMPORT_ROUTES_END", routeImportLine)
	if aerr != nil {
		return fmt.Errorf("adding route import: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_MAIN_INIT_SERVICES_START", "GF_MAIN_INIT_SERVICES_END", depsInitLine)
	if aerr != nil {
		return fmt.Errorf("adding deps init: %w", aerr)
	}
	s, aerr = appendUniqueLine(s, "GF_MAIN_MOUNT_ROUTES_START", "GF_MAIN_MOUNT_ROUTES_END", routeMountLines)
	if aerr != nil {
		return fmt.Errorf("adding route mount: %w", aerr)
	}

	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing core main.go: %w", err)
	}
	return nil
}
