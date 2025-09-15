package svelte

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gertd/go-pluralize"
)

type Column struct {
	Name string // column name in snake_case
	Type string // "string", "number", "time", "bool"
}

var pluralizeClient = pluralize.NewClient()

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func GenerateSvelteScaffolding(modelName string, columns []Column) error {
	if err := addModelToNavigation(modelName); err != nil {
		return fmt.Errorf("adding model to navigation: %w", err)
	}
	if err := generateClientConnect(modelName); err != nil {
		return fmt.Errorf("generating client connect.ts: %w", err)
	}
	if err := generateClientListPage(modelName, columns); err != nil {
		return fmt.Errorf("generating client list page: %w", err)
	}
	if err := generateClientDetailPage(modelName, columns); err != nil {
		return fmt.Errorf("generating client detail page: %w", err)
	}
	if err := generateClientListPageSpec(modelName, columns); err != nil {
		return fmt.Errorf("generating client list page spec: %w", err)
	}
	if err := generateClientDetailPageSpec(modelName, columns); err != nil {
		return fmt.Errorf("generating client detail page spec: %w", err)
	}

	// run npm i && npm run format in the service-client directory
	cmd := "cd ./app/service-client && npm i && npm run format"
	execCmd := exec.Command("bash", "-c", cmd)
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running npm commands: %w\nOutput: %s", err, string(out))
	}
	return nil
}

// addModelToNavigation adds a new navigation item for the generated model
// to the main Svelte layout file.
func addModelToNavigation(modelName string) error {
	layoutPath := "./app/service-client/src/routes/(app)/+layout.svelte"
	contentBytes, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("reading layout file %s: %w", layoutPath, err)
	}
	content := string(contentBytes)

	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)

	// Check if entry already exists to ensure idempotency.
	if strings.Contains(content, fmt.Sprintf(`href: "/models/%s"`, pluralLower)) {
		return nil
	}

	navEntry := fmt.Sprintf(`        {
            name: "%s",
            href: "/models/%s",
            icon: Bone,
        },`, pluralCap, pluralLower)

	// Insert the new nav item before the closing bracket of the nav array.
	navArrayEndMarker := `    ];`
	newContent := strings.Replace(content, navArrayEndMarker, navEntry+"\n"+navArrayEndMarker, 1)

	if newContent == content {
		return fmt.Errorf("failed to add model to navigation: insertion point '%s' not found in %s", navArrayEndMarker, layoutPath)
	}

	return os.WriteFile(layoutPath, []byte(newContent), 0644)
}

// generateClientConnect updates the client-side ConnectRPC wiring by adding the
// <Model>Service import and exporting a typed client instance in connect.ts.
func generateClientConnect(modelName string) error {
	path := "./app/service-client/src/lib/connect.ts"
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading connect.ts: %w", err)
	}
	s := string(b)

	capitalized := capitalize(modelName)
	serviceToken := capitalized + "Service"
	clientExport := "export const " + modelName + "_client = createClient(" + serviceToken + ", transport);"

	// Ensure the service is imported from main_pb
	if !strings.Contains(s, serviceToken) {
		// Assume double quotes in imports
		marker := "from \"$lib/gen/proto/v1/main_pb\""
		idx := strings.Index(s, marker)
		if idx == -1 {
			return fmt.Errorf("main_pb import not found in connect.ts")
		}
		// Find the opening brace for the import list
		pre := s[:idx]
		braceOpen := strings.LastIndex(pre, "{")
		braceClose := strings.LastIndex(pre, "}")
		if braceOpen == -1 || braceClose == -1 || braceClose < braceOpen {
			return fmt.Errorf("malformed main_pb import in connect.ts")
		}
		importList := pre[braceOpen+1 : braceClose]
		importList = strings.TrimSpace(importList)
		if importList == "" {
			importList = serviceToken
		} else {
			if !strings.HasSuffix(importList, ",") {
				importList += ","
			}
			importList += " " + serviceToken
		}
		// Rebuild the string with updated import list
		s = s[:braceOpen+1] + importList + s[braceClose:]
	}

	// Ensure the client export exists
	if !strings.Contains(s, clientExport) {
		// Insert after the transport or after last existing client export
		insertAfter := "export const skeleton_client = createClient(SkeletonService, transport);"
		pos := strings.Index(s, insertAfter)
		if pos == -1 {
			// Fallback: append at end
			if !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
			s += clientExport + "\n"
		} else {
			// Find end of that line
			lineEnd := pos + len(insertAfter)
			// Insert a newline and the new export after
			s = s[:lineEnd] + "\n" + clientExport + s[lineEnd:]
		}
	}

	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing connect.ts: %w", err)
	}
	return nil
}

// generateClientListPage scaffolds a client list page by copying the
// skeleton list Svelte file and performing token replacements for
// singular/plural model variants. Columns are not yet expanded; this
// is a straight token-based clone of the skeleton UI.
func generateClientListPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/+page.svelte"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists
	destDir := filepath.Join("app/service-client/src/routes/(app)/models", pluralLower)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "+page.svelte")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	// Token replacements (plural/title before singular to avoid partial stomps)
	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Helper: Title-case a label from snake_case
	toTitle := func(name string) string {
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, " ")
	}

	// Build headers: per model columns + Created/Updated
	var h strings.Builder
	for _, c := range columns {
		h.WriteString("                <th role=\"columnheader\">")
		h.WriteString(toTitle(c.Name))
		h.WriteString("</th>\n")
	}
	h.WriteString("                <th role=\"columnheader\">Created</th>\n")
	h.WriteString("                <th role=\"columnheader\">Updated</th>\n")
	headers := h.String()

	// Build cells: per model columns + Created/Updated
	var b strings.Builder
	for _, c := range columns {
		field := c.Name
		switch c.Type {
		case "time":
			b.WriteString("                        <td>{new Date(" + modelName + "." + field + ").toLocaleDateString()}</td>\n")
		case "bool":
			b.WriteString("                        <td>{" + modelName + "." + field + " ? \"Yes\" : \"No\"}</td>\n")
		default:
			b.WriteString("                        <td>{" + modelName + "." + field + "}</td>\n")
		}
	}
	b.WriteString("                        <td>{new Date(" + modelName + ".created).toLocaleDateString()}</td>\n")
	b.WriteString("                        <td>{new Date(" + modelName + ".updated).toLocaleDateString()}</td>\n")
	cells := b.String()

	// Replace regions delimited by markers
	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		start += len(startMarker)
		return content[:start] + "\n" + replacement + content[end:], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "<!-- GF_LIST_HEADERS_START -->", "<!-- GF_LIST_HEADERS_END -->", headers)
	if rErr != nil {
		return fmt.Errorf("replacing headers: %w", rErr)
	}
	s, rErr = replaceRegion(s, "<!-- GF_LIST_CELLS_START -->", "<!-- GF_LIST_CELLS_END -->", cells)
	if rErr != nil {
		return fmt.Errorf("replacing cells: %w", rErr)
	}

	// Remove lines that contain marker comments to avoid extra spacing
	markers := []string{
		"<!-- GF_LIST_HEADERS_START -->",
		"<!-- GF_LIST_HEADERS_END -->",
		"<!-- GF_LIST_CELLS_START -->",
		"<!-- GF_LIST_CELLS_END -->",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, m := range markers {
			if strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	// Write out the generated file
	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client list page %s: %w", destPath, err)
	}
	return nil
}

// generateClientDetailPage scaffolds a client detail/create page by copying the
// skeleton detail Svelte file and performing token replacements for
// singular/plural model variants. It also expands the column-aware regions for
// empty model defaults, form-data extraction, request payload fields, and form inputs.
func generateClientDetailPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/[skeleton_id]/+page.svelte"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists: /(app)/models/<plural>/[<model>_id]
	destDir := filepath.Join(
		"app/service-client/src/routes/(app)/models",
		pluralLower,
		"["+modelName+"_id]",
	)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "+page.svelte")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	// Token replacements (plural/title before singular to avoid partial stomps)
	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Build replacement snippets
	// 1) Empty model defaults inside empty<Model>
	var emptyB strings.Builder
	emptyIndent := "        "
	emptyB.WriteString(emptyIndent + "created: \"\",\n")
	emptyB.WriteString(emptyIndent + "updated: \"\",\n")
	emptyB.WriteString(emptyIndent + "id: \"\",\n")
	for _, c := range columns {
		switch c.Type {
		case "bool":
			emptyB.WriteString(emptyIndent + c.Name + ": false,\n")
		default:
			emptyB.WriteString(emptyIndent + c.Name + ": \"\",\n")
		}
	}
	emptySnippet := strings.TrimRight(emptyB.String(), "\n")

	// 2) FormData extraction
	var fdB strings.Builder
	fdIndent := "        "
	for _, c := range columns {
		if c.Type == "bool" {
			fdB.WriteString(fdIndent + "const " + c.Name + " = formData.get(\"" + c.Name + "\") === \"on\";\n")
		} else {
			fdB.WriteString(fdIndent + "const " + c.Name + " = formData.get(\"" + c.Name + "\")?.toString() ?? \"\";\n")
		}
	}
	formDataSnippet := strings.TrimRight(fdB.String(), "\n")

	// 3) Request payload fields for create/edit
	var reqB strings.Builder
	for _, c := range columns {
		reqB.WriteString("                        " + c.Name + ",\n")
	}
	payloadFields := strings.TrimRight(reqB.String(), "\n")

	// 4) Form input fields markup
	toTitle := func(name string) string {
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, " ")
	}
	var uiB strings.Builder
	for _, c := range columns {
		label := toTitle(c.Name)
		switch c.Type {
		case "string":
			uiB.WriteString("        <label class=\"label\" for=\"" + c.Name + "\">" + label + "</label>\n")
			uiB.WriteString("        <div>\n")
			uiB.WriteString("            <input\n")
			uiB.WriteString("                type=\"text\"\n")
			uiB.WriteString("                id=\"" + c.Name + "\"\n")
			uiB.WriteString("                name=\"" + c.Name + "\"\n")
			uiB.WriteString("                required\n")
			uiB.WriteString("                class=\"input input-bordered validator w-full\"\n")
			uiB.WriteString("                value={" + modelName + "." + c.Name + "}\n")
			uiB.WriteString("            />\n")
			uiB.WriteString("            <div class=\"validator-hint\">Enter at least 3 characters</div>\n")
			uiB.WriteString("        </div>\n\n")
		case "number":
			uiB.WriteString("        <label class=\"label\" for=\"" + c.Name + "\">" + label + "</label>\n")
			uiB.WriteString("        <div>\n")
			uiB.WriteString("            <input\n")
			uiB.WriteString("                type=\"number\"\n")
			uiB.WriteString("                id=\"" + c.Name + "\"\n")
			uiB.WriteString("                name=\"" + c.Name + "\"\n")
			uiB.WriteString("                required\n")
			uiB.WriteString("                class=\"input input-bordered validator w-full\"\n")
			uiB.WriteString("                value={" + modelName + "." + c.Name + "}\n")
			uiB.WriteString("            />\n")
			uiB.WriteString("            <div class=\"validator-hint\">Enter a positive number</div>\n")
			uiB.WriteString("        </div>\n\n")
		case "time":
			uiB.WriteString("        <label class=\"label\" for=\"" + c.Name + "\">" + label + "</label>\n")
			uiB.WriteString("        <div>\n")
			uiB.WriteString("            <input\n")
			uiB.WriteString("                type=\"date\"\n")
			uiB.WriteString("                id=\"" + c.Name + "\"\n")
			uiB.WriteString("                required\n")
			uiB.WriteString("                name=\"" + c.Name + "\"\n")
			uiB.WriteString("                class=\"input input-bordered validator w-full\"\n")
			uiB.WriteString("                value={formatDate(" + modelName + "." + c.Name + ")}\n")
			uiB.WriteString("            />\n")
			uiB.WriteString("            <div class=\"validator-hint\">Select a valid date</div>\n")
			uiB.WriteString("        </div>\n\n")
		case "bool":
			uiB.WriteString("        <label class=\"label cursor-pointer my-2\">\n")
			uiB.WriteString("            <span class=\"label-text\">" + label + "</span>\n")
			uiB.WriteString("            <input\n")
			uiB.WriteString("                name=\"" + c.Name + "\"\n")
			uiB.WriteString("                type=\"checkbox\"\n")
			uiB.WriteString("                class=\"toggle\"\n")
			uiB.WriteString("                checked={" + modelName + "." + c.Name + "}\n")
			uiB.WriteString("            />\n")
			uiB.WriteString("        </label>\n\n")
		}
	}
	fieldsSnippet := strings.TrimRight(uiB.String(), "\n")

	// Replace regions delimited by markers
	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + "\n" + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_DETAIL_EMPTY_START", "// GF_DETAIL_EMPTY_END", emptySnippet)
	if rErr != nil {
		return fmt.Errorf("replacing empty defaults: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_FORMDATA_START", "// GF_DETAIL_FORMDATA_END", formDataSnippet)
	if rErr != nil {
		return fmt.Errorf("replacing form data: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_CREATE_FIELDS_START", "// GF_DETAIL_CREATE_FIELDS_END", payloadFields)
	if rErr != nil {
		return fmt.Errorf("replacing create fields: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_FIELDS_START", "// GF_DETAIL_EDIT_FIELDS_END", payloadFields)
	if rErr != nil {
		return fmt.Errorf("replacing edit fields: %w", rErr)
	}
	s, rErr = replaceRegion(s, "<!-- GF_DETAIL_FIELDS_START -->", "<!-- GF_DETAIL_FIELDS_END -->", fieldsSnippet)
	if rErr != nil {
		return fmt.Errorf("replacing UI fields: %w", rErr)
	}

	// Remove lines that contain marker comments to avoid extra spacing
	markers := []string{
		"// GF_DETAIL_EMPTY_START", "// GF_DETAIL_EMPTY_END",
		"// GF_DETAIL_FORMDATA_START", "// GF_DETAIL_FORMDATA_END",
		"// GF_DETAIL_CREATE_FIELDS_START", "// GF_DETAIL_CREATE_FIELDS_END",
		"// GF_DETAIL_EDIT_FIELDS_START", "// GF_DETAIL_EDIT_FIELDS_END",
		"<!-- GF_DETAIL_FIELDS_START -->", "<!-- GF_DETAIL_FIELDS_END -->",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, m := range markers {
			if strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	// Write out the generated file
	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client detail page %s: %w", destPath, err)
	}
	return nil
}

// generateClientListPageSpec scaffolds a client list page test by copying the
// skeleton spec and performing token and marker-based replacements for
// singular/plural model variants and column-aware assertions.
func generateClientListPageSpec(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/page.svelte.spec.ts"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := capitalize(pluralLower)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists
	destDir := filepath.Join("app/service-client/src/routes/(app)/models", pluralLower)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "page.svelte.spec.ts")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	// Token replacements (plural/title before singular to avoid partial stomps)
	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Helper: Title-case a label from snake_case
	toTitle := func(name string) string {
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, " ")
	}

	// Build mock data creation function
	var mockB strings.Builder
	mockB.WriteString(fmt.Sprintf("function createMock%s(overrides: Partial<%s> = {}): %s {\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
	mockB.WriteString("    return {\n")
	mockB.WriteString(fmt.Sprintf("        $typeName: \"proto.v1.%s\" as const,\n", capitalizedModelName))
	mockB.WriteString("        id: \"123\",\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			mockB.WriteString(fmt.Sprintf("        %s: \"Test %s %d\",\n", c.Name, capitalizedModelName, i))
		case "number":
			mockB.WriteString(fmt.Sprintf("        %s: \"%d\",\n", c.Name, 25+i))
		case "time":
			mockB.WriteString(fmt.Sprintf("        %s: \"2023-01-%02dT00:00:00Z\",\n", c.Name, 15+i))
		case "bool":
			mockB.WriteString(fmt.Sprintf("        %s: true,\n", c.Name))
		}
	}
	mockB.WriteString("        created: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        updated: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        ...overrides,\n")
	mockB.WriteString("    };\n")
	mockB.WriteString("}\n")
	mockFields := mockB.String()

	// Build column header assertions
	var headersB strings.Builder
	for _, c := range columns {
		title := toTitle(c.Name)
		headersB.WriteString("        const " + c.Name + "Header = page.getByRole(\"columnheader\", { name: \"" + title + "\" });\n")
		headersB.WriteString("        await expect.element(" + c.Name + "Header).toBeInTheDocument();\n")
	}
	headersAssert := strings.TrimRight(headersB.String(), "\n")

	// Build row assertion for first test case
	var rowAssertB strings.Builder
	rowIndent := "            "
	for i, c := range columns {
		switch c.Type {
		case "string":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"Test %s %d\"))\n", capitalizedModelName, i))
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "number":
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"%d\")).toBeInTheDocument();\n", 25+i))
		case "time":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"1/%d/2023\"))\n", 15+i))
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "bool":
			rowAssertB.WriteString(rowIndent + "await expect\n")
			rowAssertB.WriteString(rowIndent + "    .element(row.getByRole(\"cell\", { name: \"Yes\", exact: true }))\n")
			rowAssertB.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		}
	}
	rowAssert := strings.TrimRight(rowAssertB.String(), "\n")

	// Build row assertion for second test case (different values)
	var rowAssertB2 strings.Builder
	for i, c := range columns {
		switch c.Type {
		case "string":
			rowAssertB2.WriteString(rowIndent + "await expect\n")
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"    .element(row.getByText(\"Another %s %d\"))\n", capitalizedModelName, i))
			rowAssertB2.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		case "number":
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"%d\")).toBeInTheDocument();\n", 30+i))
		case "time":
			rowAssertB2.WriteString(fmt.Sprintf(rowIndent+"await expect.element(row.getByText(\"6/%d/2024\")).toBeInTheDocument();\n", 1+i))
		case "bool":
			rowAssertB2.WriteString(rowIndent + "await expect\n")
			rowAssertB2.WriteString(rowIndent + "    .element(row.getByRole(\"cell\", { name: \"No\", exact: true }))\n")
			rowAssertB2.WriteString(rowIndent + "    .toBeInTheDocument();\n")
		}
	}
	rowAssert2 := strings.TrimRight(rowAssertB2.String(), "\n")

	// Replace regions delimited by markers
	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + "\n" + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END", mockFields)
	if rErr != nil {
		return fmt.Errorf("replacing mock fields: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END", headersAssert)
	if rErr != nil {
		return fmt.Errorf("replacing headers assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END", rowAssert)
	if rErr != nil {
		return fmt.Errorf("replacing row assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_ROW_ASSERT_2_START", "// GF_ROW_ASSERT_2_END", rowAssert2)
	if rErr != nil {
		return fmt.Errorf("replacing second row assert: %w", rErr)
	}

	// Update second test case in testData array with different mock values
	var overridesB strings.Builder
	overridesB.WriteString("                id: \"456\",\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			overridesB.WriteString(fmt.Sprintf("                %s: \"Another %s %d\",\n", c.Name, capitalizedModelName, i))
		case "number":
			overridesB.WriteString(fmt.Sprintf("                %s: \"%d\",\n", c.Name, 30+i))
		case "time":
			overridesB.WriteString(fmt.Sprintf("                %s: \"2024-06-%02dT00:00:00Z\",\n", c.Name, 1+i))
		case "bool":
			overridesB.WriteString(fmt.Sprintf("                %s: false,\n", c.Name))
		}
	}
	overridesText := strings.TrimRight(overridesB.String(), ",\n")

	startMarker := fmt.Sprintf("createMock%s({", capitalizedModelName)
	endMarker := "}),"
	startIndex := strings.Index(s, startMarker)
	if startIndex != -1 {
		endIndex := strings.Index(s[startIndex:], endMarker)
		if endIndex != -1 {
			endIndex += startIndex
			// Find the opening brace of the overrides object
			openBraceIndex := strings.Index(s[startIndex:], "{") + startIndex
			// We want to replace everything between the braces
			oldOverrides := s[openBraceIndex+1 : endIndex]
			s = strings.Replace(s, oldOverrides, "\n"+overridesText+"\n            ", 1)
		}
	}

	// Remove lines that contain marker comments
	markers := []string{
		"// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END",
		"// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END",
		"// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END",
		"// GF_ROW_ASSERT_2_START", "// GF_ROW_ASSERT_2_END",
		"// GF_ROW_SELECT_START", "// GF_ROW_SELECT_END",
		"// GF_ROW_SELECT_DELETE_START", "// GF_ROW_SELECT_DELETE_END",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, m := range markers {
			if strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	// Write out the generated file
	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client list page spec %s: %w", destPath, err)
	}
	return nil
}


// generateClientDetailPageSpec scaffolds a client detail page test by copying
// the skeleton detail spec and performing token and marker-based replacements
// for singular/plural model variants and column-aware assertions.
func generateClientDetailPageSpec(modelName string, columns []Column) error {
	sourcePath := "./app/service-client/src/routes/(app)/models/skeletons/[skeleton_id]/page.svelte.spec.ts"
	pluralLower := pluralizeClient.Plural(modelName)
	capitalizedModelName := capitalize(modelName)

	// Ensure destination directory exists: /(app)/models/<plural>/[<model>_id]
	destDir := filepath.Join(
		"app/service-client/src/routes/(app)/models",
		pluralLower,
		"["+modelName+"_id]",
	)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "page.svelte.spec.ts")

	// Read template
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	// Token replacements
	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	// Helper: Title-case a label from snake_case
	toTitle := func(name string) string {
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, " ")
	}

	// Brittle replacements for hardcoded values from skeleton template
	var firstStringColName string
	for _, c := range columns {
		if c.Type == "string" && firstStringColName == "" {
			firstStringColName = c.Name
		}
	}

	if firstStringColName != "" {
		s = strings.ReplaceAll(s, ".getByLabelText(\"Name\")", ".getByLabelText(\""+toTitle(firstStringColName)+"\")")
	}

	// Conditionally handle date-related tests
	var firstTimeColName string
	hasTimeColumn := false
	for _, c := range columns {
		if c.Type == "time" {
			firstTimeColName = c.Name
			hasTimeColumn = true
			break
		}
	}

	if hasTimeColumn {
		s = strings.ReplaceAll(s, ".getByLabelText(\"Death\")", ".getByLabelText(\""+toTitle(firstTimeColName)+"\")")
		s = strings.ReplaceAll(s, "death: \"2024-07-15T10:00:00Z\"", firstTimeColName+": \"2024-07-15T10:00:00Z\"")
		s = strings.ReplaceAll(s, "death: \"\"", firstTimeColName+": \"\"")
		s = strings.ReplaceAll(s, "death: \"invalid-date\"", firstTimeColName+": \"invalid-date\"")
		s = strings.ReplaceAll(s, "// GF_UTILITIES_TESTS_START", "")
		s = strings.ReplaceAll(s, "// GF_UTILITIES_TESTS_END", "")
	} else {
		// Remove the entire "Utilities" describe block if no time columns are present
		startMarker := "// GF_UTILITIES_TESTS_START"
		endMarker := "// GF_UTILITIES_TESTS_END"
		start := strings.Index(s, startMarker)
		end := strings.Index(s, endMarker)
		if start != -1 && end != -1 {
			s = s[:start] + s[end+len(endMarker):]
		}
	}

	// Build mock data creation function
	var mockB strings.Builder
	mockB.WriteString(fmt.Sprintf("function createMock%s(overrides: Partial<%s> = {}): %s {\n", capitalizedModelName, capitalizedModelName, capitalizedModelName))
	mockB.WriteString("    return {\n")
	mockB.WriteString(fmt.Sprintf("        $typeName: \"proto.v1.%s\" as const,\n", capitalizedModelName))
	mockB.WriteString("        id: \"123\",\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			mockB.WriteString(fmt.Sprintf("        %s: \"Bones %d\",\n", c.Name, i))
		case "number":
			mockB.WriteString(fmt.Sprintf("        %s: \"%d\",\n", c.Name, 150+i))
		case "time":
			mockB.WriteString(fmt.Sprintf("        %s: \"2023-01-%02dT00:00:00Z\",\n", c.Name, 15+i))
		case "bool":
			mockB.WriteString(fmt.Sprintf("        %s: true,\n", c.Name))
		}
	}
	mockB.WriteString("        created: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        updated: \"2022-01-01T00:00:00Z\",\n")
	mockB.WriteString("        ...overrides,\n")
	mockB.WriteString("    };\n")
	mockB.WriteString("}\n")
	mockFields := mockB.String()

	// Build create form fill operations
	var createFillB strings.Builder
	createIndent := "            "
	for i, c := range columns {
		title := toTitle(c.Name)
		switch c.Type {
		case "string":
			createFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"New %s %d\");\n", createIndent, title, capitalizedModelName, i))
		case "number":
			createFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"%d\");\n", createIndent, title, 50+i))
		case "time":
			createFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"2025-05-%02d\");\n", createIndent, title, 10+i))
		case "bool":
			createFillB.WriteString(createIndent + "await page_context.getByLabelText(\"" + title + "\").click();\n")
		}
	}
	createFill := strings.TrimRight(createFillB.String(), "\n")

	// Build create expected payload
	var createExpectB strings.Builder
	createExpectIndent := "            "
	createExpectB.WriteString(createExpectIndent + modelName + ": {\n")
	for i, c := range columns {
		switch c.Type {
		case "string":
			createExpectB.WriteString(fmt.Sprintf("                %s: \"New %s %d\",\n", c.Name, capitalizedModelName, i))
		case "number":
			createExpectB.WriteString(fmt.Sprintf("                %s: \"%d\",\n", c.Name, 50+i))
		case "time":
			createExpectB.WriteString(fmt.Sprintf("                %s: \"2025-05-%02d\",\n", c.Name, 10+i))
		case "bool":
			createExpectB.WriteString("                " + c.Name + ": true,\n")
		}
	}
	createExpectB.WriteString(createExpectIndent + "}, ")
	createExpected := createExpectB.String()

	// Build edit form change operations
	var editChangeB strings.Builder
	for i, c := range columns {
		if i == 0 { // Change the first field only for edit test
			title := toTitle(c.Name)
			switch c.Type {
			case "string":
				editChangeB.WriteString(createIndent + "await page_context.getByLabelText(\"" + title + "\").fill(\"Updated Bones\");\n")
			case "number":
				editChangeB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"200\");\n", createIndent, title))
			case "time":
				editChangeB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"2024-01-01\");\n", createIndent, title))
			case "bool":
				editChangeB.WriteString(createIndent + "await page_context.getByLabelText(\"" + title + "\").click();\n")
			}
			break
		}
	}
	editChange := strings.TrimRight(editChangeB.String(), "\n")

	// Build edit expected payload
	var editExpectB strings.Builder
	editExpectIndent := "            "
	editExpectB.WriteString(editExpectIndent + modelName + ": {\n")
	editExpectB.WriteString("                id: \"123\",\n")
	for i, c := range columns {
		if i == 0 { // Change the first field value in expected payload
			switch c.Type {
			case "string":
				editExpectB.WriteString("                " + c.Name + ": \"Updated Bones\",\n")
			case "number":
				editExpectB.WriteString("                " + c.Name + ": \"200\",\n")
			case "time":
				editExpectB.WriteString("                " + c.Name + ": \"2024-01-01\",\n")
			case "bool":
				editExpectB.WriteString("                " + c.Name + ": false,\n")
			}
		} else {
			// Keep original values for other fields
			switch c.Type {
			case "string":
				editExpectB.WriteString(fmt.Sprintf("                %s: \"Bones %d\",\n", c.Name, i))
			case "number":
				editExpectB.WriteString(fmt.Sprintf("                %s: \"%d\",\n", c.Name, 150+i))
			case "time":
				editExpectB.WriteString(fmt.Sprintf("                %s: \"2023-01-%02d\",\n", c.Name, 15+i))
			case "bool":
				editExpectB.WriteString("                " + c.Name + ": true,\n")
			}
		}
	}
	editExpectB.WriteString(editExpectIndent + "}, ")
	editExpected := editExpectB.String()

	// Build initialization assertions (empty form)
	var initAssertB strings.Builder
	initIndent := "            "
	for i, c := range columns {
		if i == 0 { // Only assert the first field for brevity
			title := toTitle(c.Name)
			initAssertB.WriteString(initIndent + "await expect\n")
			initAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			initAssertB.WriteString(initIndent + "    .toHaveValue(\"\");\n")
			break
		}
	}
	initAssert := strings.TrimRight(initAssertB.String(), "\n")

	// Build empty form assertions
	var emptyAssertB strings.Builder
	for _, c := range columns {
		title := toTitle(c.Name)
		switch c.Type {
		case "string", "time":
			emptyAssertB.WriteString(initIndent + "await expect\n")
			emptyAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			emptyAssertB.WriteString(initIndent + "    .toHaveValue(\"\");\n")
		case "number":
			emptyAssertB.WriteString(initIndent + "await expect\n")
			emptyAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			emptyAssertB.WriteString(initIndent + "    .toHaveValue(null);\n")
		case "bool":
			emptyAssertB.WriteString(initIndent + "await expect\n")
			emptyAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			emptyAssertB.WriteString(initIndent + "    .not.toBeChecked();\n")
		}
	}
	emptyAssert := strings.TrimRight(emptyAssertB.String(), "\n")

	// Build edit form assertions (populated)
	var editAssertB strings.Builder
	for i, c := range columns {
		title := toTitle(c.Name)
		switch c.Type {
		case "string":
			editAssertB.WriteString(initIndent + "await expect\n")
			editAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			editAssertB.WriteString(fmt.Sprintf(initIndent+"    .toHaveValue(\"Bones %d\");\n", i))
		case "number":
			editAssertB.WriteString(initIndent + "await expect\n")
			editAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			editAssertB.WriteString(fmt.Sprintf(initIndent+"    .toHaveValue(%d);\n", 150+i))
		case "time":
			editAssertB.WriteString(initIndent + "await expect\n")
			editAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			editAssertB.WriteString(fmt.Sprintf(initIndent+"    .toHaveValue(\"2023-01-%02d\");\n", 15+i))
		case "bool":
			editAssertB.WriteString(initIndent + "await expect\n")
			editAssertB.WriteString(initIndent + "    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			editAssertB.WriteString(initIndent + "    .toBeChecked();\n")
		}
	}
	editAssert := strings.TrimRight(editAssertB.String(), "\n")

	// Build date test assertions (using first time field if any)
	var dateTestB strings.Builder
	for _, c := range columns {
		if c.Type == "time" {
			title := toTitle(c.Name)
			dateTestB.WriteString("                await expect\n")
			dateTestB.WriteString("                    .element(page_context.getByLabelText(\"" + title + "\"))\n")
			dateTestB.WriteString("                    .toHaveValue(\"2024-07-15\");\n")
			break
		}
	}
	dateTest := strings.TrimRight(dateTestB.String(), "\n")

	// Build form data fill for create error test
	var createFillSimpleB strings.Builder
	for i, c := range columns {
		title := toTitle(c.Name)
		switch c.Type {
		case "string":
			createFillSimpleB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"Test %d\");\n", createIndent, title, i))
		case "number":
			createFillSimpleB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\" %d\");\n", createIndent, title, 10+i))
		case "time":
			createFillSimpleB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"2023-01-%02d\");\n", createIndent, title, 1+i))
		case "bool":
			// Don't click bool fields for simple fill
		}
	}
	createFillSimple := strings.TrimRight(createFillSimpleB.String(), "\n")

	// Build formdata fill for edit test
	var editFormdataFillB strings.Builder
	for i, c := range columns {
		title := toTitle(c.Name)
		switch c.Type {
		case "string":
			editFormdataFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"Test Name %d\");\n", createIndent, title, i))
		case "number":
			editFormdataFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\" %d\");\n", createIndent, title, 25+i))
		case "time":
			editFormdataFillB.WriteString(fmt.Sprintf("%sawait page_context.getByLabelText(\"%s\").fill(\"2024-01-%02d\");\n", createIndent, title, 1+i))
		case "bool":
			editFormdataFillB.WriteString(createIndent + "const " + c.Name + "Checkbox = page_context\n")
			editFormdataFillB.WriteString(createIndent + "    .getByLabelText(\"" + title + "\")\n")
			editFormdataFillB.WriteString(createIndent + "    .query() as HTMLInputElement;\n")
			editFormdataFillB.WriteString(createIndent + "if (" + c.Name + "Checkbox.checked) {\n")
			editFormdataFillB.WriteString(createIndent + "    " + c.Name + "Checkbox.click(); // Uncheck if checked\n")
			editFormdataFillB.WriteString(createIndent + "}\n")
		}
	}
	editFormdataFill := strings.TrimRight(editFormdataFillB.String(), "\n")

	// Build formdata expected payload
	var fields []string
	fields = append(fields, "                    id: \"test-id-123\"")
	for i, c := range columns {
		switch c.Type {
		case "string":
			fields = append(fields, fmt.Sprintf("                    %s: \"Test Name %d\"", c.Name, i))
		case "number":
			fields = append(fields, fmt.Sprintf("                    %s: \"%d\"", c.Name, 25+i))
		case "time":
			fields = append(fields, fmt.Sprintf("                    %s: \"2024-01-%02d\"", c.Name, 1+i))
		case "bool":
			fields = append(fields, "                    "+c.Name+": false")
		}
	}
	fieldsString := strings.Join(fields, ",\n")

	editFormdataExpected := fmt.Sprintf(`            expect(mockEdit%s).toHaveBeenCalledWith({ 
                %s: {
%s
                },
            });`, capitalizedModelName, modelName, fieldsString)

	// Build fallback rename operations
	var fallbackRenameB strings.Builder
	for _, c := range columns {
		fallbackRenameB.WriteString(createIndent + "const " + c.Name + "El = container.querySelector(\n")
		fallbackRenameB.WriteString(createIndent + "    \"#" + c.Name + "\"\n")
		fallbackRenameB.WriteString(createIndent + ") as HTMLInputElement | null;\n")
		fallbackRenameB.WriteString(createIndent + "if (" + c.Name + "El) {\n")
		fallbackRenameB.WriteString(createIndent + "    " + c.Name + "El.setAttribute(\"name\", \"_" + c.Name + "\");\n")
		fallbackRenameB.WriteString(createIndent + "}\n")
	}
	fallbackRename := strings.TrimRight(fallbackRenameB.String(), "\n")

	// Build fallback expected values
	var fallbackExpectB strings.Builder
	for _, c := range columns {
		switch c.Type {
		case "bool":
			fallbackExpectB.WriteString("                    " + c.Name + ": false,\n")
		default:
			fallbackExpectB.WriteString("                    " + c.Name + ": \"\",\n")
		}
	}
	fallbackExpected := strings.TrimRight(fallbackExpectB.String(), "\n")

	// Build edit fallback expected values (for existing data)
	var editFallbackExpectB strings.Builder
	for _, c := range columns {
		editFallbackExpectB.WriteString("                    " + c.Name + ": expect.any(")
		switch c.Type {
		case "bool":
			editFallbackExpectB.WriteString("Boolean")
		default:
			editFallbackExpectB.WriteString("String")
		}
		editFallbackExpectB.WriteString("),\n")
	}
	editFallbackExpected := strings.TrimRight(editFallbackExpectB.String(), "\n")

	// Replace regions delimited by markers
	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + "\n" + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var rErr error
	s, rErr = replaceRegion(s, "// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END", mockFields)
	if rErr != nil {
		return fmt.Errorf("replacing mock fields: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_CREATE_FILL_START", "// GF_DETAIL_CREATE_FILL_END", createFill)
	if rErr != nil {
		return fmt.Errorf("replacing create fill: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_CREATE_EXPECT_START", "// GF_DETAIL_CREATE_EXPECT_END", createExpected)
	if rErr != nil {
		return fmt.Errorf("replacing create expect: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_CHANGE_START", "// GF_DETAIL_EDIT_CHANGE_END", editChange)
	if rErr != nil {
		return fmt.Errorf("replacing edit change: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_EXPECT_START", "// GF_DETAIL_EDIT_EXPECT_END", editExpected)
	if rErr != nil {
		return fmt.Errorf("replacing edit expect: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_INIT_ASSERT_START", "// GF_DETAIL_INIT_ASSERT_END", initAssert)
	if rErr != nil {
		return fmt.Errorf("replacing init assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EMPTY_ASSERT_START", "// GF_DETAIL_EMPTY_ASSERT_END", emptyAssert)
	if rErr != nil {
		return fmt.Errorf("replacing empty assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_ASSERT_START", "// GF_DETAIL_EDIT_ASSERT_END", editAssert)
	if rErr != nil {
		return fmt.Errorf("replacing edit assert: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_DATE_TEST_START", "// GF_DETAIL_DATE_TEST_END", dateTest)
	if rErr != nil {
		return fmt.Errorf("replacing date test: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_FORMDATA_FILL_START", "// GF_DETAIL_EDIT_FORMDATA_FILL_END", editFormdataFill)
	if rErr != nil {
		return fmt.Errorf("replacing edit formdata fill: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_FORMDATA_EXPECT_START", "// GF_DETAIL_EDIT_FORMDATA_EXPECT_END", editFormdataExpected)
	if rErr != nil {
		return fmt.Errorf("replacing edit formdata expect: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_FALLBACK_RENAME_START", "// GF_DETAIL_FALLBACK_RENAME_END", fallbackRename)
	if rErr != nil {
		return fmt.Errorf("replacing fallback rename: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_FALLBACK_EXPECT_START", "// GF_DETAIL_FALLBACK_EXPECT_END", fallbackExpected)
	if rErr != nil {
		return fmt.Errorf("replacing fallback expect: %w", rErr)
	}
	s, rErr = replaceRegion(s, "// GF_DETAIL_EDIT_FALLBACK_EXPECT_START", "// GF_DETAIL_EDIT_FALLBACK_EXPECT_END", editFallbackExpected)
	if rErr != nil {
		return fmt.Errorf("replacing edit fallback expect: %w", rErr)
	}

	// Replace all remaining occurrences of the simple create fill pattern
	for strings.Contains(s, "// GF_DETAIL_CREATE_FILL_START") {
		s, rErr = replaceRegion(s, "// GF_DETAIL_CREATE_FILL_START", "// GF_DETAIL_CREATE_FILL_END", createFillSimple)
		if rErr != nil {
			return fmt.Errorf("replacing create fill simple: %w", rErr)
		}
	}

	// Remove lines that contain marker comments
	markers := []string{
		"// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END",
		"// GF_DETAIL_CREATE_FILL_START", "// GF_DETAIL_CREATE_FILL_END",
		"// GF_DETAIL_CREATE_EXPECT_START", "// GF_DETAIL_CREATE_EXPECT_END",
		"// GF_DETAIL_EDIT_CHANGE_START", "// GF_DETAIL_EDIT_CHANGE_END",
		"// GF_DETAIL_EDIT_EXPECT_START", "// GF_DETAIL_EDIT_EXPECT_END",
		"// GF_DETAIL_INIT_ASSERT_START", "// GF_DETAIL_INIT_ASSERT_END",
		"// GF_DETAIL_EMPTY_ASSERT_START", "// GF_DETAIL_EMPTY_ASSERT_END",
		"// GF_DETAIL_EDIT_ASSERT_START", "// GF_DETAIL_EDIT_ASSERT_END",
		"// GF_DETAIL_DATE_TEST_START", "// GF_DETAIL_DATE_TEST_END",
		"// GF_DETAIL_EDIT_FORMDATA_FILL_START", "// GF_DETAIL_EDIT_FORMDATA_FILL_END",
		"// GF_DETAIL_EDIT_FORMDATA_EXPECT_START", "// GF_DETAIL_EDIT_FORMDATA_EXPECT_END",
		"// GF_DETAIL_FALLBACK_RENAME_START", "// GF_DETAIL_FALLBACK_RENAME_END",
		"// GF_DETAIL_FALLBACK_EXPECT_START", "// GF_DETAIL_FALLBACK_EXPECT_END",
		"// GF_DETAIL_EDIT_FALLBACK_EXPECT_START", "// GF_DETAIL_EDIT_FALLBACK_EXPECT_END",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, m := range markers {
			if strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	// Write out the generated file
	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing client detail page spec %s: %w", destPath, err)
	}
	return nil
}
