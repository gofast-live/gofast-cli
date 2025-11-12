package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
		columnStrings := args[1:]

		var columns []Column
		seenNames := map[string]bool{}
		validTypes := map[string]bool{
			"string": true,
			"number": true,
			"date":   true,
			"bool":   true,
		}

		var counter int
		for _, colStr := range columnStrings {
			parts := strings.Split(colStr, ":")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				cmd.Printf("Error: Invalid column format '%s'. Use name:type.\n", colStr)
				return
			}

			colType := strings.ToLower(parts[1])
			if !validTypes[colType] {
				cmd.Printf("Error: Invalid type '%s' for column '%s'.\n", parts[1], parts[0])
				cmd.Println("Valid types are: string, number, date, bool.")
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
		case "date":
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

	buildQueryFieldsWithI := func(zero bool) string {
		parts := []string{
			"ID: uuid.New()",
			"UserID: uuid.New()",
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
				// Generate: <Field>: fmt.Sprintf("Test %d", i)
				parts = append(parts, fmt.Sprintf("%s: fmt.Sprintf(\"Test %s\", i)", name, "%d"))
			case "number":
				// sqlc maps numeric to string; use a quoted literal
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
		case "date":
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

	// Remove or adapt numeric/date/bool asserts that come from skeleton template
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
