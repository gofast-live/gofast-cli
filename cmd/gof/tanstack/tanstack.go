package tanstack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/gertd/go-pluralize"
)

type Column struct {
	Name string
	Type string
}

func GetModelPath(modelName string) string {
	return "/models/" + pluralize.NewClient().Plural(modelName)
}

func FormatProject(ctx context.Context) error {
	cmd := `cd ./app/service-tanstack && npm ci && node --input-type=module -e "import { Generator } from '@tanstack/router-generator'; import { getConfig } from '@tanstack/router-plugin'; const root = process.cwd(); const generator = new Generator({ config: getConfig({}, root), root }); await generator.run();" && npm run format`
	execCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running npm commands: %w\nOutput: %s", err, string(out))
	}
	return nil
}

func GenerateTanstackScaffolding(modelName string, columns []Column) error {
	err := generateClientConnect(modelName)
	if err != nil {
		return fmt.Errorf("generating client connect.ts: %w", err)
	}
	err = generateClientListPage(modelName, columns)
	if err != nil {
		return fmt.Errorf("generating client list page: %w", err)
	}
	err = generateClientDetailPage(modelName, columns)
	if err != nil {
		return fmt.Errorf("generating client detail page: %w", err)
	}
	return nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

func replaceProtoFieldAccess(content, fieldName, replacement string) string {
	pattern := regexp.MustCompile(`\.` + regexp.QuoteMeta(fieldName) + `\b`)
	return pattern.ReplaceAllString(content, "."+replacement)
}

func generateClientConnect(modelName string) error {
	path := "./app/service-tanstack/src/lib/connect.ts"
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading connect.ts: %w", err)
	}
	s := string(b)

	pascalName := toPascalCase(modelName)
	serviceToken := pascalName + "Service"
	clientExport := "export const " + modelName + "_client = createClient(" + serviceToken + ", transport)"

	if !strings.Contains(s, serviceToken) {
		s, err = addMainPbImport(s, serviceToken)
		if err != nil {
			return err
		}
	}

	if !strings.Contains(s, clientExport) {
		insertAfter := "export const skeleton_client = createClient(SkeletonService, transport)"
		pos := strings.Index(s, insertAfter)
		if pos == -1 {
			if !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
			s += clientExport + "\n"
		} else {
			lineEnd := pos + len(insertAfter)
			s = s[:lineEnd] + "\n" + clientExport + s[lineEnd:]
		}
	}

	err = os.WriteFile(path, []byte(s), 0o644)
	if err != nil {
		return fmt.Errorf("writing connect.ts: %w", err)
	}
	return nil
}

func generateClientListPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-tanstack/src/routes/_layout/models/skeletons/index.tsx"
	pluralLower := pluralize.NewClient().Plural(modelName)
	pluralCap := toPascalCase(pluralLower)
	capitalizedModelName := toPascalCase(modelName)

	destDir := filepath.Join("app/service-tanstack/src/routes/_layout/models", pluralLower)
	err := os.MkdirAll(destDir, 0o755)
	if err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "index.tsx")

	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	camelName := toCamelCase(modelName)
	s = replaceProtoFieldAccess(s, "skeleton", camelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

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

	var headersBuilder strings.Builder
	for _, c := range columns {
		headersBuilder.WriteString("                <th role=\"columnheader\">")
		headersBuilder.WriteString(toTitle(c.Name))
		headersBuilder.WriteString("</th>\n")
	}
	headersBuilder.WriteString("                <th role=\"columnheader\">Created</th>\n")
	headersBuilder.WriteString("                <th role=\"columnheader\">Updated</th>\n")

	var cellsBuilder strings.Builder
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "date":
			cellsBuilder.WriteString("                    <td>{new Date(" + modelName + "." + field + ").toLocaleDateString()}</td>\n")
		case "bool":
			cellsBuilder.WriteString("                    <td>{" + modelName + "." + field + " ? 'Yes' : 'No'}</td>\n")
		default:
			cellsBuilder.WriteString("                    <td>{" + modelName + "." + field + "}</td>\n")
		}
	}
	cellsBuilder.WriteString("                    <td>{new Date(" + modelName + ".created).toLocaleDateString()}</td>\n")
	cellsBuilder.WriteString("                    <td>{new Date(" + modelName + ".updated).toLocaleDateString()}</td>\n")

	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		start += len(startMarker)
		return content[:start] + "\n" + replacement + content[end:], nil
	}

	var replaceErr error
	s, replaceErr = replaceRegion(s, "{/* GF_LIST_HEADERS_START */}", "{/* GF_LIST_HEADERS_END */}", headersBuilder.String())
	if replaceErr != nil {
		return fmt.Errorf("replacing headers: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "{/* GF_LIST_CELLS_START */}", "{/* GF_LIST_CELLS_END */}", cellsBuilder.String())
	if replaceErr != nil {
		return fmt.Errorf("replacing cells: %w", replaceErr)
	}

	markers := []string{
		"{/* GF_LIST_HEADERS_START */}",
		"{/* GF_LIST_HEADERS_END */}",
		"{/* GF_LIST_CELLS_START */}",
		"{/* GF_LIST_CELLS_END */}",
	}
	var outLines []string
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, marker := range markers {
			if strings.Contains(line, marker) {
				skip = true
				break
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	err = os.WriteFile(destPath, []byte(s), 0o644)
	if err != nil {
		return fmt.Errorf("writing client list page %s: %w", destPath, err)
	}
	return nil
}

func generateClientDetailPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-tanstack/src/routes/_layout/models/skeletons/$skeleton_id.tsx"
	pluralLower := pluralize.NewClient().Plural(modelName)
	pluralCap := toPascalCase(pluralLower)
	capitalizedModelName := toPascalCase(modelName)

	destDir := filepath.Join("app/service-tanstack/src/routes/_layout/models", pluralLower)
	err := os.MkdirAll(destDir, 0o755)
	if err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}
	destPath := filepath.Join(destDir, "$"+modelName+"_id.tsx")

	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading template file %s: %w", sourcePath, err)
	}

	s := string(contentBytes)
	s = strings.ReplaceAll(s, "Skeletons", pluralCap)
	s = strings.ReplaceAll(s, "skeletons", pluralLower)
	s = strings.ReplaceAll(s, "Skeleton", capitalizedModelName)
	camelName := toCamelCase(modelName)
	s = replaceProtoFieldAccess(s, "skeleton", camelName)
	s = strings.ReplaceAll(s, "skeleton: {", camelName+": {")
	s = strings.ReplaceAll(s, "skeleton", modelName)

	emptySnippet := detailEmptySnippet(columns)
	formDataSnippet := detailFormDataSnippet(columns)
	payloadSnippet := detailPayloadFields(columns)

	fieldsSnippet := detailFieldsMarkup(modelName, columns)

	replaceRegion := func(content, startMarker, endMarker, replacement string) (string, error) {
		start := strings.Index(content, startMarker)
		end := strings.Index(content, endMarker)
		if start == -1 || end == -1 || end < start {
			return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
		}
		return content[:start] + replacement + "\n" + content[end+len(endMarker):], nil
	}

	var replaceErr error
	s, replaceErr = replaceRegion(s, "// GF_DETAIL_EMPTY_START", "// GF_DETAIL_EMPTY_END", emptySnippet)
	if replaceErr != nil {
		return fmt.Errorf("replacing empty defaults: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "// GF_DETAIL_FORMDATA_START", "// GF_DETAIL_FORMDATA_END", formDataSnippet)
	if replaceErr != nil {
		return fmt.Errorf("replacing form data: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "// GF_DETAIL_CREATE_FIELDS_START", "// GF_DETAIL_CREATE_FIELDS_END", payloadSnippet)
	if replaceErr != nil {
		return fmt.Errorf("replacing create fields: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "// GF_DETAIL_EDIT_FIELDS_START", "// GF_DETAIL_EDIT_FIELDS_END", payloadSnippet)
	if replaceErr != nil {
		return fmt.Errorf("replacing edit fields: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "{/* GF_DETAIL_FIELDS_START */}", "{/* GF_DETAIL_FIELDS_END */}", fieldsSnippet)
	if replaceErr != nil {
		return fmt.Errorf("replacing UI fields: %w", replaceErr)
	}

	s = stripDetailMarkers(s, slices.ContainsFunc(columns, func(c Column) bool { return c.Type == "date" }))

	err = os.WriteFile(destPath, []byte(s), 0o644)
	if err != nil {
		return fmt.Errorf("writing client detail page %s: %w", destPath, err)
	}
	return nil
}

// detailEmptySnippet builds the empty<Model> defaults, using camelCase proto field names.
func detailEmptySnippet(columns []Column) string {
	var emptyBuilder strings.Builder
	emptyIndent := "  "
	emptyBuilder.WriteString(emptyIndent + "created: '',\n")
	emptyBuilder.WriteString(emptyIndent + "updated: '',\n")
	emptyBuilder.WriteString(emptyIndent + "id: '',\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		if c.Type == "bool" {
			emptyBuilder.WriteString(emptyIndent + field + ": false,\n")
			continue
		}
		emptyBuilder.WriteString(emptyIndent + field + ": '',\n")
	}
	return strings.TrimRight(emptyBuilder.String(), "\n")
}

// detailFormDataSnippet reads form fields (snake_case names) into camelCase variables.
func detailFormDataSnippet(columns []Column) string {
	var formDataBuilder strings.Builder
	formDataIndent := "    "
	for _, c := range columns {
		field := toCamelCase(c.Name)
		if c.Type == "bool" {
			formDataBuilder.WriteString(formDataIndent + "const " + field + " = formData.get('" + c.Name + "') === 'on'\n")
			continue
		}
		formDataBuilder.WriteString(formDataIndent + "const " + field + " = formData.get('" + c.Name + "')?.toString() ?? ''\n")
	}
	return strings.TrimRight(formDataBuilder.String(), "\n")
}

// detailPayloadFields lists the camelCase request fields for create and edit.
func detailPayloadFields(columns []Column) string {
	var payloadBuilder strings.Builder
	for _, c := range columns {
		field := toCamelCase(c.Name)
		payloadBuilder.WriteString("            " + field + ",\n")
	}
	return strings.TrimRight(payloadBuilder.String(), "\n")
}

// detailFieldsMarkup renders one form input per column.
func detailFieldsMarkup(modelName string, columns []Column) string {
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

	var fieldsBuilder strings.Builder
	for _, c := range columns {
		label := toTitle(c.Name)
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			fieldsBuilder.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsBuilder.WriteString("            " + label + "\n")
			fieldsBuilder.WriteString("          </label>\n")
			fieldsBuilder.WriteString("          <div>\n")
			fieldsBuilder.WriteString("            <input\n")
			fieldsBuilder.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              type=\"text\"\n")
			fieldsBuilder.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              required\n")
			fieldsBuilder.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsBuilder.WriteString("              defaultValue={" + modelName + "." + field + "}\n")
			fieldsBuilder.WriteString("            />\n")
			fieldsBuilder.WriteString("            <div className=\"validator-hint\">Enter at least 3 characters</div>\n")
			fieldsBuilder.WriteString("          </div>\n")
		case "number":
			fieldsBuilder.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsBuilder.WriteString("            " + label + "\n")
			fieldsBuilder.WriteString("          </label>\n")
			fieldsBuilder.WriteString("          <div>\n")
			fieldsBuilder.WriteString("            <input\n")
			fieldsBuilder.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              type=\"number\"\n")
			fieldsBuilder.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              required\n")
			fieldsBuilder.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsBuilder.WriteString("              defaultValue={" + modelName + "." + field + "}\n")
			fieldsBuilder.WriteString("            />\n")
			fieldsBuilder.WriteString("            <div className=\"validator-hint\">Enter a positive number</div>\n")
			fieldsBuilder.WriteString("          </div>\n")
		case "date":
			fieldsBuilder.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsBuilder.WriteString("            " + label + "\n")
			fieldsBuilder.WriteString("          </label>\n")
			fieldsBuilder.WriteString("          <div>\n")
			fieldsBuilder.WriteString("            <input\n")
			fieldsBuilder.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              type=\"date\"\n")
			fieldsBuilder.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              required\n")
			fieldsBuilder.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsBuilder.WriteString("              defaultValue={formatDate(" + modelName + "." + field + ")}\n")
			fieldsBuilder.WriteString("            />\n")
			fieldsBuilder.WriteString("            <div className=\"validator-hint\">Select a valid date</div>\n")
			fieldsBuilder.WriteString("          </div>\n")
		case "bool":
			fieldsBuilder.WriteString("          <label className=\"label my-2 cursor-pointer\" htmlFor=\"" + c.Name + "\">\n")
			fieldsBuilder.WriteString("            <span className=\"label-text\">" + label + "</span>\n")
			fieldsBuilder.WriteString("            <input\n")
			fieldsBuilder.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsBuilder.WriteString("              type=\"checkbox\"\n")
			fieldsBuilder.WriteString("              className=\"toggle\"\n")
			fieldsBuilder.WriteString("              defaultChecked={" + modelName + "." + field + "}\n")
			fieldsBuilder.WriteString("            />\n")
			fieldsBuilder.WriteString("          </label>\n")
		}
	}
	return strings.TrimRight(fieldsBuilder.String(), "\n")
}

// stripDetailMarkers drops marker comment lines, and the formatDate helper when no column is a date.
func stripDetailMarkers(s string, hasDateColumn bool) string {
	markers := []string{
		"// GF_DETAIL_EMPTY_START", "// GF_DETAIL_EMPTY_END",
		"// GF_DETAIL_FORMDATA_START", "// GF_DETAIL_FORMDATA_END",
		"// GF_DETAIL_CREATE_FIELDS_START", "// GF_DETAIL_CREATE_FIELDS_END",
		"// GF_DETAIL_EDIT_FIELDS_START", "// GF_DETAIL_EDIT_FIELDS_END",
		"{/* GF_DETAIL_FIELDS_START */}", "{/* GF_DETAIL_FIELDS_END */}",
	}
	var outLines []string
	inFormatDateFunc := false
	braceDepth := 0
	for line := range strings.SplitSeq(s, "\n") {
		skip := false
		for _, marker := range markers {
			if strings.Contains(line, marker) {
				skip = true
				break
			}
		}
		// Remove formatDate function if no date columns
		if !hasDateColumn && strings.HasPrefix(strings.TrimSpace(line), "function formatDate(") {
			inFormatDateFunc = true
			braceDepth = 0
		}
		if !hasDateColumn && inFormatDateFunc {
			skip = true
			braceDepth, inFormatDateFunc = trackFunctionEnd(line, braceDepth)
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	return strings.Join(outLines, "\n")
}

// trackFunctionEnd updates the brace depth for one line and reports whether the function is still open.
func trackFunctionEnd(line string, braceDepth int) (int, bool) {
	for _, ch := range line {
		switch ch {
		case '{':
			braceDepth++
		case '}':
			braceDepth--
			if braceDepth == 0 {
				return braceDepth, false
			}
		}
	}
	return braceDepth, true
}

// addMainPbImport appends serviceToken to the main_pb import list in connect.ts, one import per line.
func addMainPbImport(s, serviceToken string) (string, error) {
	marker := "from './gen/proto/v1/main_pb'"
	pre, _, found := strings.Cut(s, marker)
	if !found {
		return "", errors.New("main_pb import not found in connect.ts")
	}
	braceOpen := strings.LastIndex(pre, "{")
	braceClose := strings.LastIndex(pre, "}")
	if braceOpen == -1 || braceClose == -1 || braceClose < braceOpen {
		return "", errors.New("malformed main_pb import in connect.ts")
	}
	importList := strings.TrimSpace(pre[braceOpen+1 : braceClose])
	switch {
	case importList == "":
		importList = serviceToken
	case strings.HasSuffix(importList, ","):
		importList += "\n  " + serviceToken
	default:
		importList += ",\n  " + serviceToken
	}
	return s[:braceOpen+1] + "\n  " + importList + "\n" + s[braceClose:], nil
}
