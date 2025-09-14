package svelte

import (
	"fmt"
	"os"
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
		marker := "from '$lib/gen/proto/v1/main_pb'"
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
		start += len(startMarker)
		return content[:start] + "\n" + replacement + "\n" + content[end:], nil
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
    // Adjust type import path proto file name
    s = strings.ReplaceAll(s, "skeleton_pb", modelName+"_pb")

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

    // Build header assertions for each column + Created/Updated
    var h strings.Builder
    for _, c := range columns {
        title := toTitle(c.Name)
        varName := c.Name + "Header"
        h.WriteString("        const " + varName + " = page.getByRole(\"columnheader\", { name: \"" + title + "\" });\n")
        h.WriteString("        await expect.element(" + varName + ").toBeInTheDocument();\n\n")
    }
    h.WriteString("        const createdHeader = page.getByRole(\"columnheader\", { name: \"Created\" });\n")
    h.WriteString("        await expect.element(createdHeader).toBeInTheDocument();\n\n")
    h.WriteString("        const updatedHeader = page.getByRole(\"columnheader\", { name: \"Updated\" });\n")
    h.WriteString("        await expect.element(updatedHeader).toBeInTheDocument();\n")
    headers := strings.TrimRight(h.String(), "\n")

    // Determine first occurrences of each type for row assertions and tests
    firstStr := ""
    firstNum := ""
    firstTime := ""
    firstBool := ""
    for _, c := range columns {
        switch c.Type {
        case "string":
            if firstStr == "" { firstStr = c.Name }
        case "number":
            if firstNum == "" { firstNum = c.Name }
        case "time":
            if firstTime == "" { firstTime = c.Name }
        case "bool":
            if firstBool == "" { firstBool = c.Name }
        }
    }

    // Build mock object fields for createMock<Model>
    var mf strings.Builder
    mf.WriteString("function createMock" + capitalizedModelName + "(overrides: Partial<" + capitalizedModelName + "> = {}): " + capitalizedModelName + " {\n")
    mf.WriteString("    return {\n")
    mf.WriteString("        '$typeName': 'proto.v1." + capitalizedModelName + "' as const,\n")
    mf.WriteString("        id: \"123\",\n")
    didFirstStr := false
    for _, c := range columns {
        switch c.Type {
        case "string":
            if !didFirstStr {
                mf.WriteString("        " + c.Name + ": \"Test " + capitalizedModelName + "\",\n")
                didFirstStr = true
            } else {
                mf.WriteString("        " + c.Name + ": \"Other " + toTitle(c.Name) + "\",\n")
            }
        case "number":
            mf.WriteString("        " + c.Name + ": \"25\",\n")
        case "time":
            mf.WriteString("        " + c.Name + ": \"2023-01-15T00:00:00Z\",\n")
        case "bool":
            mf.WriteString("        " + c.Name + ": true,\n")
        }
    }
    mf.WriteString("        created: \"2022-01-01T00:00:00Z\",\n")
    mf.WriteString("        updated: \"2022-01-01T00:00:00Z\",\n")
    mf.WriteString("        ...overrides\n")
    mf.WriteString("    };\n")
    mf.WriteString("}\n")
    mockFunc := mf.String()

    // Build row selection for data-available test
    var rowSelect strings.Builder
    if firstStr != "" {
        rowSelect.WriteString("        const row = page.getByRole(\"row\", { name: /test " + modelName + "/i });\n")
        rowSelect.WriteString("        await expect.element(row).toBeInTheDocument();")
    } else {
        rowSelect.WriteString("        const row = page.getByRole(\"row\", { name: /Edit/i });\n")
        rowSelect.WriteString("        await expect.element(row).toBeInTheDocument();")
    }
    rowSelectBlock := rowSelect.String()

    // Build row assertions for cells (use row-scoped if we selected by name, else page-scoped)
    var ra strings.Builder
    if firstStr != "" {
        ra.WriteString("        await expect.element(row.getByText(\"Test " + capitalizedModelName + "\")).toBeInTheDocument();\n")
    }
    if firstNum != "" {
        if firstStr != "" {
            ra.WriteString("        await expect.element(row.getByText(\"25\")).toBeInTheDocument();\n")
        } else {
            ra.WriteString("        await expect.element(page.getByText(\"25\")).toBeInTheDocument();\n")
        }
    }
    if firstTime != "" {
        if firstStr != "" {
            ra.WriteString("        await expect.element(row.getByText(\"1/15/2023\")).toBeInTheDocument();\n")
        } else {
            ra.WriteString("        await expect.element(page.getByText(\"1/15/2023\")).toBeInTheDocument();\n")
        }
    }
    if firstBool != "" {
        if firstStr != "" {
            ra.WriteString("        await expect.element(row.getByRole('cell', { name: 'Yes' })).toBeInTheDocument();\n")
        } else {
            ra.WriteString("        await expect.element(page.getByRole('cell', { name: 'Yes' })).toBeInTheDocument();\n")
        }
    }
    rowAsserts := strings.TrimRight(ra.String(), "\n")

    // Build delete test row selection (prefer string field, fallback to 'Edit')
    var delSel strings.Builder
    rowVar := modelName + "Row"
    if firstStr != "" {
        delSel.WriteString("        const " + rowVar + " = page.getByRole(\"row\", { name: /test " + modelName + "/i });\n")
        delSel.WriteString("        await expect.element(" + rowVar + ").toBeInTheDocument();")
    } else {
        delSel.WriteString("        const " + rowVar + " = page.getByRole(\"row\", { name: /Edit/i });\n")
        delSel.WriteString("        await expect.element(" + rowVar + ").toBeInTheDocument();")
    }
    deleteRowSelect := delSel.String()

    // Replace regions delimited by markers
    replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
        start := strings.Index(content, startMarker)
        end := strings.Index(content, endMarker)
        if start == -1 || end == -1 || end < start {
            return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
        }
        start += len(startMarker)
        return content[:start] + "\n" + replacement + "\n" + content[end:], nil
    }

    var rErr error
    s, rErr = replaceRegion(s, "// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END", mockFunc)
    if rErr != nil { return fmt.Errorf("replacing mock fields: %w", rErr) }
    s, rErr = replaceRegion(s, "// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END", headers)
    if rErr != nil { return fmt.Errorf("replacing header asserts: %w", rErr) }
    s, rErr = replaceRegion(s, "// GF_ROW_SELECT_START", "// GF_ROW_SELECT_END", rowSelectBlock)
    if rErr != nil { return fmt.Errorf("replacing row select: %w", rErr) }
    s, rErr = replaceRegion(s, "// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END", rowAsserts)
    if rErr != nil { return fmt.Errorf("replacing row asserts: %w", rErr) }
    s, rErr = replaceRegion(s, "// GF_ROW_SELECT_DELETE_START", "// GF_ROW_SELECT_DELETE_END", deleteRowSelect)
    if rErr != nil { return fmt.Errorf("replacing delete row select: %w", rErr) }

    // If we keep the bool test, update its row selection similarly
    if firstBool != "" {
        bStart := strings.Index(s, "// GF_BOOL_TEST_START")
        bEnd := strings.Index(s, "// GF_BOOL_TEST_END")
        if bStart != -1 && bEnd != -1 && bEnd > bStart {
            region := s[bStart:bEnd]
            // Replace the standard two-line selection with our rowSelectBlock
            // Look for the first occurrence of "const row =" and the subsequent expect line
            selIdx := strings.Index(region, "const row =")
            if selIdx != -1 {
                expNeedle := "await expect.element(row).toBeInTheDocument();"
                expIdx := strings.Index(region[selIdx:], expNeedle)
                if expIdx != -1 {
                    expEnd := selIdx + expIdx + len(expNeedle)
                    newRegion := region[:selIdx] + rowSelectBlock + region[expEnd:]
                    s = s[:bStart] + newRegion + s[bEnd:]
                }
            }
        }
    }

    // Remove the non-bool test if model has no boolean columns
    if firstBool == "" {
        start := strings.Index(s, "// GF_BOOL_TEST_START")
        end := strings.Index(s, "// GF_BOOL_TEST_END")
        if start != -1 && end != -1 && end > start {
            // Remove from start to end line inclusive
            s = s[:start] + s[end+len("// GF_BOOL_TEST_END"):]
        }
    } else {
        // Keep; nothing to do
    }

    // Finally, strip marker lines from output
    markers := []string{
        "// GF_MOCK_FIELDS_START", "// GF_MOCK_FIELDS_END",
        "// GF_HEADERS_ASSERT_START", "// GF_HEADERS_ASSERT_END",
        "// GF_ROW_SELECT_START", "// GF_ROW_SELECT_END",
        "// GF_ROW_ASSERT_START", "// GF_ROW_ASSERT_END",
        "// GF_ROW_SELECT_DELETE_START", "// GF_ROW_SELECT_DELETE_END",
        "// GF_BOOL_TEST_START", "// GF_BOOL_TEST_END",
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
