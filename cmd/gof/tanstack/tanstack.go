package tanstack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
)

type Column struct {
	Name string // column name in snake_case
	Type string // "string", "number", "date", "bool"
}

var pluralizeClient = pluralize.NewClient()

// GetModelPath returns the client-side path for a model (e.g., "/models/notes" for "note")
func GetModelPath(modelName string) string {
	return "/models/" + pluralizeClient.Plural(modelName)
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

func toTitleCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func replaceRegion(content, startMarker, endMarker, replacement string) (string, error) {
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if start == -1 || end == -1 || end < start {
		return content, fmt.Errorf("markers %q .. %q not found", startMarker, endMarker)
	}
	start += len(startMarker)
	return content[:start] + "\n" + replacement + content[end:], nil
}

func replaceRegexOnce(content, pattern, replacement string) (string, error) {
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(content)
	if loc == nil {
		return content, fmt.Errorf("pattern not found: %s", pattern)
	}
	return content[:loc[0]] + replacement + content[loc[1]:], nil
}

// UpdateUserPermissions adds new model permissions to the user management page
// (_layout/users/$user_id.tsx) by inserting entries before GF_PERMISSIONS_END marker.
func UpdateUserPermissions(modelName string) error {
	path := "./app/service-tanstack/src/routes/_layout/users/$user_id.tsx"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading user details page: %w", err)
	}
	content := string(contentBytes)

	modelCap := toPascalCase(modelName)
	modelPlural := pluralizeClient.Plural(modelName)
	modelPluralCap := toPascalCase(modelPlural)

	if strings.Contains(content, "Get"+modelPluralCap) {
		return nil
	}

	const marker = "// GF_PERMISSIONS_END"
	e := strings.Index(content, marker)
	if e == -1 {
		return fmt.Errorf("permissions marker %s not found in user details page", marker)
	}

	beforeMarker := content[:e]
	lastBit := -1
	for i := len(beforeMarker) - 1; i >= 0; i-- {
		if i >= 4 && beforeMarker[i-4:i+1] == "bit: " {
			numStart := i + 1
			numEnd := numStart
			for numEnd < len(beforeMarker) && beforeMarker[numEnd] >= '0' && beforeMarker[numEnd] <= '9' {
				numEnd++
			}
			if numEnd > numStart {
				if n, convErr := strconv.Atoi(beforeMarker[numStart:numEnd]); convErr == nil && n > lastBit {
					lastBit = n
				}
			}
			break
		}
	}
	if lastBit == -1 {
		return fmt.Errorf("could not find last bit number in permissions array")
	}

	nextBit := lastBit + 1
	newPerms := fmt.Sprintf(`  { name: 'Get%[1]s', bit: %[3]d },
  { name: 'Create%[2]s', bit: %[4]d },
  { name: 'Edit%[2]s', bit: %[5]d },
  { name: 'Remove%[2]s', bit: %[6]d },
`, modelPluralCap, modelCap, nextBit, nextBit+1, nextBit+2, nextBit+3)

	markerLineStart := strings.LastIndex(content[:e], "\n") + 1

	prevLineEnd := markerLineStart - 1
	for prevLineEnd > 0 && content[prevLineEnd-1] == '\n' {
		prevLineEnd--
	}
	prevLineStart := strings.LastIndex(content[:prevLineEnd], "\n") + 1
	prevLine := content[prevLineStart:prevLineEnd]
	trimmedPrev := strings.TrimSpace(prevLine)

	if strings.HasSuffix(trimmedPrev, "}") && !strings.HasSuffix(trimmedPrev, "},") {
		bracePos := prevLineStart + strings.LastIndex(prevLine, "}")
		content = content[:bracePos+1] + "," + content[bracePos+1:]
		markerLineStart++
	}

	content = content[:markerLineStart] + newPerms + content[markerLineStart:]
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing user details page: %w", err)
	}
	return nil
}

func GenerateTanstackScaffolding(modelName string, columns []Column) error {
	if err := generateClientConnect(modelName); err != nil {
		return fmt.Errorf("generating tanstack connect.ts: %w", err)
	}
	if err := generateClientListPage(modelName, columns); err != nil {
		return fmt.Errorf("generating tanstack list page: %w", err)
	}
	if err := generateClientDetailPage(modelName, columns); err != nil {
		return fmt.Errorf("generating tanstack detail page: %w", err)
	}

	cmd := "cd ./app/service-tanstack && npm ci && npm run format"
	execCmd := exec.Command("bash", "-c", cmd)
	out, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running tanstack npm commands: %w\nOutput: %s", err, string(out))
	}
	return nil
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
		marker := "from './gen/proto/v1/main_pb'"
		idx := strings.Index(s, marker)
		if idx == -1 {
			return fmt.Errorf("main_pb import not found in tanstack connect.ts")
		}
		pre := s[:idx]
		braceOpen := strings.LastIndex(pre, "{")
		braceClose := strings.LastIndex(pre, "}")
		if braceOpen == -1 || braceClose == -1 || braceClose < braceOpen {
			return fmt.Errorf("malformed main_pb import in tanstack connect.ts")
		}
		importList := strings.TrimSpace(pre[braceOpen+1 : braceClose])
		if importList == "" {
			importList = serviceToken
		} else {
			if !strings.HasSuffix(importList, ",") {
				importList += ","
			}
			importList += "\n  " + serviceToken
		}
		s = s[:braceOpen+1] + "\n  " + importList + "\n" + s[braceClose:]
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

	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing tanstack connect.ts: %w", err)
	}
	return nil
}

func generateClientListPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-tanstack/src/routes/_layout/models/skeletons/index.tsx"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := toPascalCase(pluralLower)
	capitalizedModelName := toPascalCase(modelName)

	destDir := filepath.Join("app/service-tanstack/src/routes/_layout/models", pluralLower)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
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
	s = strings.ReplaceAll(s, "res.skeleton", "res."+camelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	var headersB strings.Builder
	for _, c := range columns {
		headersB.WriteString("                <th role=\"columnheader\">")
		headersB.WriteString(toTitleCase(c.Name))
		headersB.WriteString("</th>\n")
	}
	headersB.WriteString("                <th role=\"columnheader\">Created</th>\n")
	headersB.WriteString("                <th role=\"columnheader\">Updated</th>\n")

	var cellsB strings.Builder
	for _, c := range columns {
		field := toCamelCase(c.Name)
		switch c.Type {
		case "date":
			cellsB.WriteString("                    <td>{new Date(" + modelName + "." + field + ").toLocaleDateString()}</td>\n")
		case "bool":
			cellsB.WriteString("                    <td>{" + modelName + "." + field + " ? 'Yes' : 'No'}</td>\n")
		default:
			cellsB.WriteString("                    <td>{" + modelName + "." + field + "}</td>\n")
		}
	}
	cellsB.WriteString("                    <td>{new Date(" + modelName + ".created).toLocaleDateString()}</td>\n")
	cellsB.WriteString("                    <td>{new Date(" + modelName + ".updated).toLocaleDateString()}</td>\n")

	var replaceErr error
	s, replaceErr = replaceRegion(s, "{/* GF_LIST_HEADERS_START */}", "{/* GF_LIST_HEADERS_END */}", headersB.String())
	if replaceErr != nil {
		return fmt.Errorf("replacing tanstack list headers: %w", replaceErr)
	}
	s, replaceErr = replaceRegion(s, "{/* GF_LIST_CELLS_START */}", "{/* GF_LIST_CELLS_END */}", cellsB.String())
	if replaceErr != nil {
		return fmt.Errorf("replacing tanstack list cells: %w", replaceErr)
	}

	loadingColSpan := len(columns) + 3
	s = strings.ReplaceAll(s, "colSpan={7}", fmt.Sprintf("colSpan={%d}", loadingColSpan))

	markers := []string{
		"{/* GF_LIST_HEADERS_START */}",
		"{/* GF_LIST_HEADERS_END */}",
		"{/* GF_LIST_CELLS_START */}",
		"{/* GF_LIST_CELLS_END */}",
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

	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing tanstack list page %s: %w", destPath, err)
	}
	return nil
}

func generateClientDetailPage(modelName string, columns []Column) error {
	sourcePath := "./app/service-tanstack/src/routes/_layout/models/skeletons/$skeleton_id.tsx"
	pluralLower := pluralizeClient.Plural(modelName)
	pluralCap := toPascalCase(pluralLower)
	capitalizedModelName := toPascalCase(modelName)

	destDir := filepath.Join("app/service-tanstack/src/routes/_layout/models", pluralLower)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
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
	s = strings.ReplaceAll(s, "s.skeleton", "s."+camelName)
	s = strings.ReplaceAll(s, "skeleton", modelName)

	idVar := modelName + "_id"
	paramRe := regexp.MustCompile(`const \{\s*([a-z0-9_]+)\s*\} = Route\.useParams\(\)`)
	m := paramRe.FindStringSubmatch(s)
	if len(m) == 2 {
		idVar = m[1]
	}

	var stateB strings.Builder
	stateB.WriteString("  const [form, setForm] = useState({\n")
	stateB.WriteString("    id: '',\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		if c.Type == "bool" {
			stateB.WriteString("    " + field + ": false,\n")
		} else {
			stateB.WriteString("    " + field + ": '',\n")
		}
	}
	stateB.WriteString("  })")

	var newStateB strings.Builder
	newStateB.WriteString("if (" + idVar + " === 'new') {\n")
	newStateB.WriteString("        setForm({\n")
	newStateB.WriteString("          id: '',\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		if c.Type == "bool" {
			newStateB.WriteString("          " + field + ": false,\n")
		} else {
			newStateB.WriteString("          " + field + ": '',\n")
		}
	}
	newStateB.WriteString("        })\n")
	newStateB.WriteString("        return\n")
	newStateB.WriteString("      }")

	var fetchedStateB strings.Builder
	fetchedStateB.WriteString("setForm({\n")
	fetchedStateB.WriteString("          id: sk.id,\n")
	for _, c := range columns {
		field := toCamelCase(c.Name)
		if c.Type == "date" {
			fetchedStateB.WriteString("          " + field + ": formatDate(sk." + field + "),\n")
		} else {
			fetchedStateB.WriteString("          " + field + ": sk." + field + ",\n")
		}
	}
	fetchedStateB.WriteString("        })")

	var payloadB strings.Builder
	for _, c := range columns {
		field := toCamelCase(c.Name)
		payloadB.WriteString("            " + field + ": form." + field + ",\n")
	}

	var fieldsB strings.Builder
	for _, c := range columns {
		label := toTitleCase(c.Name)
		field := toCamelCase(c.Name)
		switch c.Type {
		case "string":
			fieldsB.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsB.WriteString("            " + label + "\n")
			fieldsB.WriteString("          </label>\n")
			fieldsB.WriteString("          <div>\n")
			fieldsB.WriteString("            <input\n")
			fieldsB.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              type=\"text\"\n")
			fieldsB.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              required\n")
			fieldsB.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsB.WriteString("              value={form." + field + "}\n")
			fieldsB.WriteString("              onChange={(e) => {\n")
			fieldsB.WriteString("                const value = e.currentTarget.value\n")
			fieldsB.WriteString("                setForm((prev) => ({ ...prev, " + field + ": value }))\n")
			fieldsB.WriteString("              }}\n")
			fieldsB.WriteString("            />\n")
			fieldsB.WriteString("            <div className=\"validator-hint\">Enter at least 3 characters</div>\n")
			fieldsB.WriteString("          </div>\n")
		case "number":
			fieldsB.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsB.WriteString("            " + label + "\n")
			fieldsB.WriteString("          </label>\n")
			fieldsB.WriteString("          <div>\n")
			fieldsB.WriteString("            <input\n")
			fieldsB.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              type=\"number\"\n")
			fieldsB.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              required\n")
			fieldsB.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsB.WriteString("              value={form." + field + "}\n")
			fieldsB.WriteString("              onChange={(e) => {\n")
			fieldsB.WriteString("                const value = e.currentTarget.value\n")
			fieldsB.WriteString("                setForm((prev) => ({ ...prev, " + field + ": value }))\n")
			fieldsB.WriteString("              }}\n")
			fieldsB.WriteString("            />\n")
			fieldsB.WriteString("            <div className=\"validator-hint\">Enter a positive number</div>\n")
			fieldsB.WriteString("          </div>\n")
		case "date":
			fieldsB.WriteString("          <label className=\"label\" htmlFor=\"" + c.Name + "\">\n")
			fieldsB.WriteString("            " + label + "\n")
			fieldsB.WriteString("          </label>\n")
			fieldsB.WriteString("          <div>\n")
			fieldsB.WriteString("            <input\n")
			fieldsB.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              type=\"date\"\n")
			fieldsB.WriteString("              required\n")
			fieldsB.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              className=\"input input-bordered validator w-full\"\n")
			fieldsB.WriteString("              value={form." + field + "}\n")
			fieldsB.WriteString("              onChange={(e) => {\n")
			fieldsB.WriteString("                const value = e.currentTarget.value\n")
			fieldsB.WriteString("                setForm((prev) => ({ ...prev, " + field + ": value }))\n")
			fieldsB.WriteString("              }}\n")
			fieldsB.WriteString("            />\n")
			fieldsB.WriteString("            <div className=\"validator-hint\">Select a valid date</div>\n")
			fieldsB.WriteString("          </div>\n")
		case "bool":
			fieldsB.WriteString("          <label className=\"label my-2 cursor-pointer\" htmlFor=\"" + c.Name + "\">\n")
			fieldsB.WriteString("            <span className=\"label-text\">" + label + "</span>\n")
			fieldsB.WriteString("            <input\n")
			fieldsB.WriteString("              id=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              name=\"" + c.Name + "\"\n")
			fieldsB.WriteString("              type=\"checkbox\"\n")
			fieldsB.WriteString("              className=\"toggle\"\n")
			fieldsB.WriteString("              checked={form." + field + "}\n")
			fieldsB.WriteString("              onChange={(e) => {\n")
			fieldsB.WriteString("                const checked = e.currentTarget.checked\n")
			fieldsB.WriteString("                setForm((prev) => ({\n")
			fieldsB.WriteString("                  ...prev,\n")
			fieldsB.WriteString("                  " + field + ": checked,\n")
			fieldsB.WriteString("                }))\n")
			fieldsB.WriteString("              }}\n")
			fieldsB.WriteString("            />\n")
			fieldsB.WriteString("          </label>\n")
		}
	}

	s, err = replaceRegexOnce(s, `(?s)const \[form, setForm\] = useState\(\{.*?\n\s*\}\)`, stateB.String())
	if err != nil {
		return fmt.Errorf("updating tanstack form state: %w", err)
	}
	s, err = replaceRegexOnce(s, `(?s)if \([a-z0-9_]+ === 'new'\) \{\n\s*setForm\(\{.*?\n\s*\}\)\n\s*return\n\s*\}`, newStateB.String())
	if err != nil {
		return fmt.Errorf("updating tanstack new-state block: %w", err)
	}
	s, err = replaceRegexOnce(s, `(?s)setForm\(\{\n\s*id: sk\.id,.*?\n\s*\}\)`, fetchedStateB.String())
	if err != nil {
		return fmt.Errorf("updating tanstack fetched-state block: %w", err)
	}

	s, err = replaceRegion(s, "// GF_DETAIL_CREATE_FIELDS_START", "// GF_DETAIL_CREATE_FIELDS_END", payloadB.String())
	if err != nil {
		return fmt.Errorf("replacing tanstack create fields: %w", err)
	}
	s, err = replaceRegion(s, "// GF_DETAIL_EDIT_FIELDS_START", "// GF_DETAIL_EDIT_FIELDS_END", payloadB.String())
	if err != nil {
		return fmt.Errorf("replacing tanstack edit fields: %w", err)
	}
	s, err = replaceRegion(s, "{/* GF_DETAIL_FIELDS_START */}", "{/* GF_DETAIL_FIELDS_END */}", strings.TrimRight(fieldsB.String(), "\n"))
	if err != nil {
		return fmt.Errorf("replacing tanstack detail fields: %w", err)
	}

	hasDateColumn := false
	for _, c := range columns {
		if c.Type == "date" {
			hasDateColumn = true
			break
		}
	}

	markers := []string{
		"// GF_DETAIL_CREATE_FIELDS_START",
		"// GF_DETAIL_CREATE_FIELDS_END",
		"// GF_DETAIL_EDIT_FIELDS_START",
		"// GF_DETAIL_EDIT_FIELDS_END",
		"{/* GF_DETAIL_FIELDS_START */}",
		"{/* GF_DETAIL_FIELDS_END */}",
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
		if !hasDateColumn {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "function formatDate(") {
				inFormatDateFunc = true
				braceDepth = 0
				skip = true
			}
			if inFormatDateFunc {
				skip = true
				for _, ch := range line {
					if ch == '{' {
						braceDepth++
					} else if ch == '}' {
						braceDepth--
						if braceDepth == 0 {
							inFormatDateFunc = false
							break
						}
					}
				}
			}
		}
		if !skip {
			outLines = append(outLines, line)
		}
	}
	s = strings.Join(outLines, "\n")

	if err := os.WriteFile(destPath, []byte(s), 0o644); err != nil {
		return fmt.Errorf("writing tanstack detail page %s: %w", destPath, err)
	}
	return nil
}
