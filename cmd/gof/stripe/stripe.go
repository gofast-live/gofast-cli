package stripe

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/repo"
)

// Strip removes all stripe-related files and code from a freshly initialized project.
// Called by init command after downloading the template.
func Strip(projectPath string) error {
	// 1. Remove payment domain folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "domain", "payment")); err != nil {
		return fmt.Errorf("removing payment domain: %w", err)
	}

	// 2. Remove payment transport folder
	if err := os.RemoveAll(filepath.Join(projectPath, "app", "service-core", "transport", "payment")); err != nil {
		return fmt.Errorf("removing payment transport: %w", err)
	}

	// 3. Remove subscriptions migration
	if err := os.Remove(filepath.Join(projectPath, "app", "service-core", "storage", "migrations", "00003_create_subscriptions.sql")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing subscriptions migration: %w", err)
	}

	// 4. Strip payment from main.go
	if err := stripMainGo(projectPath); err != nil {
		return fmt.Errorf("stripping main.go: %w", err)
	}

	// 5. Strip payment from main.proto
	if err := stripMainProto(projectPath); err != nil {
		return fmt.Errorf("stripping main.proto: %w", err)
	}

	// 6. Simplify CheckUserAccess in login service
	if err := simplifyLoginService(projectPath); err != nil {
		return fmt.Errorf("simplifying login service: %w", err)
	}

	// 7. Strip subscription queries from query.sql
	if err := stripSubscriptionQueries(projectPath); err != nil {
		return fmt.Errorf("stripping subscription queries: %w", err)
	}

	// 8. Strip GF_STRIPE markers from all Go files
	if err := stripStripeMarkers(projectPath); err != nil {
		return fmt.Errorf("stripping stripe markers: %w", err)
	}

	return nil
}

// stripMainGo removes payment imports, deps, and route mounting from main.go
func stripMainGo(projectPath string) error {
	mainPath := filepath.Join(projectPath, "app", "service-core", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	s := string(content)

	// Remove payment imports
	s = removeLines(s, `"gofast/service-core/domain/payment"`)
	s = removeLines(s, `paymentRoute "gofast/service-core/transport/payment"`)

	// Remove payment deps initialization
	s = removeBlock(s, "// Initialize payment deps", "}")

	// Remove payment route mounting
	s = removeBlock(s, "// Mount payment routes", `server.MountFunc("/payments-webhook", paymentServer.Webhook)`)

	return os.WriteFile(mainPath, []byte(s), 0644)
}

// stripMainProto removes payment service definitions from main.proto
func stripMainProto(projectPath string) error {
	protoPath := filepath.Join(projectPath, "proto", "v1", "main.proto")
	content, err := os.ReadFile(protoPath)
	if err != nil {
		return err
	}

	s := string(content)

	// Remove entire payment section from "// --- Payment Service ---" to end of file
	// (PaymentService is typically the last service in the proto)
	startMarker := "// --- Payment Service ---"
	startIdx := strings.Index(s, startMarker)
	if startIdx != -1 {
		// Find start of line
		lineStart := strings.LastIndex(s[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		}
		s = strings.TrimRight(s[:lineStart], "\n") + "\n"
	}

	return os.WriteFile(protoPath, []byte(s), 0644)
}

// stripSubscriptionQueries removes subscription-related queries from query.sql
func stripSubscriptionQueries(projectPath string) error {
	queryPath := filepath.Join(projectPath, "app", "service-core", "storage", "query.sql")
	content, err := os.ReadFile(queryPath)
	if err != nil {
		return err
	}

	s := string(content)

	// Remove the entire subscriptions section (from "-- Subscriptions --" to end of UpsertSubscription)
	startMarker := "-- Subscriptions --"
	startIdx := strings.Index(s, startMarker)
	if startIdx != -1 {
		// Find start of line
		lineStart := strings.LastIndex(s[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		}

		// Find end - look for the "returning *;" that ends UpsertSubscription
		endMarker := "returning *;"
		endIdx := strings.Index(s[startIdx:], endMarker)
		if endIdx != -1 {
			endIdx = startIdx + endIdx + len(endMarker)
			// Skip trailing newlines
			for endIdx < len(s) && (s[endIdx] == '\n' || s[endIdx] == '\r') {
				endIdx++
			}
			s = s[:lineStart] + s[endIdx:]
		}
	}

	return os.WriteFile(queryPath, []byte(s), 0644)
}

// stripStripeMarkers removes all code between GF_STRIPE_START and GF_STRIPE_END markers from Go files
func stripStripeMarkers(projectPath string) error {
	serviceCoreDir := filepath.Join(projectPath, "app", "service-core")

	return filepath.Walk(serviceCoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		s := string(content)
		if !strings.Contains(s, "// GF_STRIPE_START") {
			return nil
		}

		s = removeMarkerBlocks(s, "// GF_STRIPE_START", "// GF_STRIPE_END")

		return os.WriteFile(path, []byte(s), 0644)
	})
}

// removeMarkerBlocks removes all blocks between startMarker and endMarker (inclusive of markers)
func removeMarkerBlocks(content, startMarker, endMarker string) string {
	for {
		startIdx := strings.Index(content, startMarker)
		if startIdx == -1 {
			break
		}

		// Find start of line containing start marker
		lineStart := strings.LastIndex(content[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++ // Move past the newline
		}

		// Find end marker
		endIdx := strings.Index(content[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx = startIdx + endIdx + len(endMarker)

		// Skip to end of line
		if endIdx < len(content) && content[endIdx] == '\n' {
			endIdx++
		}

		content = content[:lineStart] + content[endIdx:]
	}
	return content
}

// restoreStripeMarkers walks through the template and restores GF_STRIPE blocks to matching files
func restoreStripeMarkers(tmpProject string) error {
	templateServiceCore := filepath.Join(tmpProject, "app", "service-core")
	localServiceCore := filepath.Join("app", "service-core")

	return filepath.Walk(templateServiceCore, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(srcPath, ".go") {
			return nil
		}

		srcContent, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		// Skip files without stripe markers
		if !strings.Contains(string(srcContent), "// GF_STRIPE_START") {
			return nil
		}

		// Get relative path and build destination path
		relPath, err := filepath.Rel(templateServiceCore, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(localServiceCore, relPath)

		// Check if destination file exists
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil // Skip if dest doesn't exist
			}
			return err
		}

		// Restore each stripe block from source to destination
		newContent := restoreMarkerBlocksToFile(string(srcContent), string(dstContent))

		return os.WriteFile(dstPath, []byte(newContent), 0644)
	})
}

// restoreMarkerBlocksToFile finds GF_STRIPE blocks in src and inserts them into dst at matching locations
func restoreMarkerBlocksToFile(src, dst string) string {
	startMarker := "// GF_STRIPE_START"
	endMarker := "// GF_STRIPE_END"

	// Find all blocks in source
	srcRemainder := src
	for {
		startIdx := strings.Index(srcRemainder, startMarker)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(srcRemainder[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}
		endIdx = startIdx + endIdx + len(endMarker)

		// Include to end of line
		if endIdx < len(srcRemainder) && srcRemainder[endIdx] == '\n' {
			endIdx++
		}

		// Find the line start for the block
		lineStart := strings.LastIndex(srcRemainder[:startIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++
		}

		block := srcRemainder[lineStart:endIdx]

		// Find context: what comes before this block in the source?
		// Look for the previous non-empty line to use as an anchor
		anchor := findAnchorBefore(srcRemainder, lineStart)

		// Insert block into dst after the anchor
		dst = insertBlockAfterAnchor(dst, block, anchor)

		srcRemainder = srcRemainder[endIdx:]
	}

	return dst
}

// findAnchorBefore finds a unique line before the given position to use as an anchor
func findAnchorBefore(content string, pos int) string {
	// Look backwards for a non-empty, non-comment line
	lines := strings.Split(content[:pos], "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") {
			return line
		}
	}
	return ""
}

// insertBlockAfterAnchor inserts block into content after the line containing anchor
func insertBlockAfterAnchor(content, block, anchor string) string {
	if anchor == "" {
		return content
	}

	// Find anchor in content
	anchorIdx := strings.Index(content, anchor)
	if anchorIdx == -1 {
		return content
	}

	// Find end of the anchor line
	lineEnd := strings.Index(content[anchorIdx:], "\n")
	if lineEnd == -1 {
		lineEnd = len(content) - anchorIdx
	}
	insertPos := anchorIdx + lineEnd + 1

	// Extract a unique identifier from the block to check for duplicates
	// Use the first non-marker line as identifier
	blockLines := strings.Split(block, "\n")
	var identifier string
	for _, line := range blockLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "// GF_STRIPE") {
			identifier = trimmed
			break
		}
	}

	// Check if this specific block already exists
	if identifier != "" && strings.Contains(content, identifier) {
		return content
	}

	return content[:insertPos] + block + content[insertPos:]
}

// simplifyLoginService replaces CheckUserAccess with a simple version that doesn't check subscriptions
func simplifyLoginService(projectPath string) error {
	loginPath := filepath.Join(projectPath, "app", "service-core", "domain", "login", "service.go")
	content, err := os.ReadFile(loginPath)
	if err != nil {
		return err
	}

	s := string(content)

	// Find and replace the full CheckUserAccess function with simplified version
	oldFunc := `func CheckUserAccess(ctx context.Context, deps Deps, user query.User) (subscriptionActive bool, access int64, err error) {
	ctx, span, done := ot.StartSpan(ctx, "login.service.CheckUserAccess")
	defer func() { done(err) }()

	// Start with base access (no plan bits)
	access = user.Access
	access &^= auth.ProPlan
	access &^= auth.BasicPlan

	// Check for active subscription
	sub, err := deps.Store.SelectActiveSubscription(ctx, user.ID)
	if err != nil {
		// No active subscription found
		span.AddEvent("No active subscription")
		return false, access, nil
	}

	// Derive plan bits from subscription
	span.AddEvent("Active subscription found")
	switch sub.StripePriceID {
	case deps.Cfg.StripePriceIDBasic:
		access |= auth.BasicPlan
	case deps.Cfg.StripePriceIDPro:
		access |= auth.ProPlan
	}

	return true, access, nil
}`

	newFunc := `func CheckUserAccess(ctx context.Context, deps Deps, user query.User) (subscriptionActive bool, access int64, err error) {
	// Simplified version without Stripe subscription checking
	// Use 'gof add stripe' to enable subscription-based access control
	return false, user.Access, nil
}`

	s = strings.Replace(s, oldFunc, newFunc, 1)

	return os.WriteFile(loginPath, []byte(s), 0644)
}

// removeLines removes all lines containing the given substring
func removeLines(content, substring string) string {
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		if !strings.Contains(line, substring) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// removeBlock removes a block of text starting with startMarker and ending with endMarker (inclusive)
func removeBlock(content, startMarker, endMarker string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}

	// Find end marker after start
	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		return content
	}
	endIdx += startIdx + len(endMarker)

	// Include trailing newline if present
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}

	// Find start of the line containing startMarker
	lineStart := strings.LastIndex(content[:startIdx], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++ // skip the newline itself
	}

	return content[:lineStart] + content[endIdx:]
}

// Add adds Stripe payment integration to an existing project.
// Called by 'gof add stripe' command.
func Add(email, apiKey string) error {
	// 1. Download template to temp location
	tmpDir, err := os.MkdirTemp("", "gofast-stripe-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpProject := filepath.Join(tmpDir, "template")
	if err := repo.DownloadRepo(email, apiKey, tmpProject); err != nil {
		return fmt.Errorf("downloading template: %w", err)
	}

	// 2. Copy payment domain folder
	srcDomain := filepath.Join(tmpProject, "app", "service-core", "domain", "payment")
	dstDomain := filepath.Join("app", "service-core", "domain", "payment")
	if err := copyDir(srcDomain, dstDomain); err != nil {
		return fmt.Errorf("copying payment domain: %w", err)
	}

	// 3. Copy payment transport folder
	srcTransport := filepath.Join(tmpProject, "app", "service-core", "transport", "payment")
	dstTransport := filepath.Join("app", "service-core", "transport", "payment")
	if err := copyDir(srcTransport, dstTransport); err != nil {
		return fmt.Errorf("copying payment transport: %w", err)
	}

	// 4. Copy and renumber subscriptions migration
	if err := addSubscriptionsMigration(tmpProject); err != nil {
		return fmt.Errorf("adding subscriptions migration: %w", err)
	}

	// 5. Add payment proto definitions
	if err := addPaymentProto(tmpProject); err != nil {
		return fmt.Errorf("adding payment proto: %w", err)
	}

	// 6. Wire payment into main.go
	if err := wirePaymentMain(tmpProject); err != nil {
		return fmt.Errorf("wiring payment into main.go: %w", err)
	}

	// 7. Restore full CheckUserAccess in login service
	if err := restoreLoginService(tmpProject); err != nil {
		return fmt.Errorf("restoring login service: %w", err)
	}

	// 8. Add subscription queries to query.sql
	if err := addSubscriptionQueries(tmpProject); err != nil {
		return fmt.Errorf("adding subscription queries: %w", err)
	}

	// 9. Restore GF_STRIPE marker blocks from template
	if err := restoreStripeMarkers(tmpProject); err != nil {
		return fmt.Errorf("restoring stripe markers: %w", err)
	}

	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, info.Mode())
	})
}

// addSubscriptionsMigration copies the subscriptions migration with the next available number
func addSubscriptionsMigration(tmpProject string) error {
	// Find next migration number
	migrationsDir := filepath.Join("app", "service-core", "storage", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var maxNum int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			// Extract number from filename like "00004_create_foo.sql"
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) >= 1 {
				num, _ := strconv.Atoi(parts[0])
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}
	nextNum := maxNum + 1

	// Read source migration
	srcPath := filepath.Join(tmpProject, "app", "service-core", "storage", "migrations", "00003_create_subscriptions.sql")
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Write with new number
	dstName := fmt.Sprintf("%05d_create_subscriptions.sql", nextNum)
	dstPath := filepath.Join(migrationsDir, dstName)
	return os.WriteFile(dstPath, content, 0644)
}

// addPaymentProto appends payment service definitions to main.proto
func addPaymentProto(tmpProject string) error {
	// Read source proto to extract payment section
	srcPath := filepath.Join(tmpProject, "proto", "v1", "main.proto")
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Extract payment section
	paymentSection := extractBlock(string(srcContent), "// --- Payment Service ---", "}")
	if paymentSection == "" {
		return fmt.Errorf("payment section not found in template proto")
	}

	// Read destination proto
	dstPath := filepath.Join("proto", "v1", "main.proto")
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return err
	}

	// Check if already present
	if strings.Contains(string(dstContent), "PaymentService") {
		return nil // Already added
	}

	// Append payment section
	newContent := string(dstContent) + "\n" + paymentSection + "\n"
	return os.WriteFile(dstPath, []byte(newContent), 0644)
}

// wirePaymentMain adds payment imports, deps, and route mounting to main.go
func wirePaymentMain(tmpProject string) error {
	mainPath := filepath.Join("app", "service-core", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	s := string(content)

	// Check if already wired
	if strings.Contains(s, `"gofast/service-core/domain/payment"`) {
		return nil // Already added
	}

	// Add payment import after login import
	s = strings.Replace(s,
		`loginSvc "gofast/service-core/domain/login"`,
		`loginSvc "gofast/service-core/domain/login"
	"gofast/service-core/domain/payment"`,
		1)

	// Add payment route import after login route import
	s = strings.Replace(s,
		`loginRoute "gofast/service-core/transport/login"`,
		`loginRoute "gofast/service-core/transport/login"
	paymentRoute "gofast/service-core/transport/payment"`,
		1)

	// Add payment deps after login deps
	s = strings.Replace(s,
		`Twilio: loginSvc.NewTwilioClient(),
	}`,
		`Twilio: loginSvc.NewTwilioClient(),
	}
	// Initialize payment deps
	paymentDeps := payment.Deps{
		Cfg:   cfg,
		Store: store,
	}`,
		1)

	// Add payment route mounting after login route mounting
	s = strings.Replace(s,
		`server.MountFunc("/login-callback", loginServer.LoginCallback)`,
		`server.MountFunc("/login-callback", loginServer.LoginCallback)
	// Mount payment routes
	paymentServer := paymentRoute.NewPaymentServer(paymentDeps)
	path, handler = v1connect.NewPaymentServiceHandler(paymentServer, server.Interceptors())
	server.Mount(path, handler)
	server.MountFunc("/payments-webhook", paymentServer.Webhook)`,
		1)

	return os.WriteFile(mainPath, []byte(s), 0644)
}

// restoreLoginService replaces simplified CheckUserAccess with full version
func restoreLoginService(tmpProject string) error {
	// Read full version from template
	srcPath := filepath.Join(tmpProject, "app", "service-core", "domain", "login", "service.go")
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	fullFunc := extractFunction(string(srcContent), "func CheckUserAccess")
	if fullFunc == "" {
		return fmt.Errorf("CheckUserAccess not found in template")
	}

	// Read current login service
	dstPath := filepath.Join("app", "service-core", "domain", "login", "service.go")
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return err
	}

	// Replace simplified version with full version
	simplifiedFunc := `func CheckUserAccess(ctx context.Context, deps Deps, user query.User) (subscriptionActive bool, access int64, err error) {
	// Simplified version without Stripe subscription checking
	// Use 'gof add stripe' to enable subscription-based access control
	return false, user.Access, nil
}`

	s := strings.Replace(string(dstContent), simplifiedFunc, fullFunc, 1)
	return os.WriteFile(dstPath, []byte(s), 0644)
}

// extractBlock extracts text from startMarker to the matching closing brace
func extractBlock(content, startMarker, endMarker string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return ""
	}

	// Find the start of the line
	lineStart := strings.LastIndex(content[:startIdx], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	// For payment section, find the service closing brace
	// Count braces to find matching close
	remainder := content[startIdx:]
	braceCount := 0
	endIdx := 0
	inService := false

	for i, ch := range remainder {
		if strings.HasPrefix(remainder[i:], "service PaymentService") {
			inService = true
		}
		if ch == '{' {
			braceCount++
		} else if ch == '}' {
			braceCount--
			if inService && braceCount == 0 {
				endIdx = startIdx + i + 1
				break
			}
		}
	}

	if endIdx == 0 {
		return ""
	}

	return content[lineStart:endIdx]
}

// extractFunction extracts a function definition from content
func extractFunction(content, funcSignature string) string {
	startIdx := strings.Index(content, funcSignature)
	if startIdx == -1 {
		return ""
	}

	// Find opening brace
	braceStart := strings.Index(content[startIdx:], "{")
	if braceStart == -1 {
		return ""
	}

	// Count braces to find matching close
	braceCount := 0
	for i := startIdx + braceStart; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return content[startIdx : i+1]
			}
		}
	}

	return ""
}

// addSubscriptionQueries appends subscription queries to query.sql
func addSubscriptionQueries(tmpProject string) error {
	// Read source query.sql to extract subscription section
	srcPath := filepath.Join(tmpProject, "app", "service-core", "storage", "query.sql")
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Extract subscription section
	startMarker := "-- Subscriptions --"
	startIdx := strings.Index(string(srcContent), startMarker)
	if startIdx == -1 {
		return fmt.Errorf("subscriptions section not found in template query.sql")
	}

	// Find start of line
	lineStart := strings.LastIndex(string(srcContent[:startIdx]), "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	// Find end - look for "returning *;"
	endMarker := "returning *;"
	endIdx := strings.Index(string(srcContent[startIdx:]), endMarker)
	if endIdx == -1 {
		return fmt.Errorf("end of subscriptions section not found")
	}
	endIdx = startIdx + endIdx + len(endMarker)

	subscriptionSection := string(srcContent[lineStart:endIdx])

	// Read destination query.sql
	dstPath := filepath.Join("app", "service-core", "storage", "query.sql")
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return err
	}

	// Check if already present
	if strings.Contains(string(dstContent), "-- Subscriptions --") {
		return nil // Already added
	}

	// Append subscription section
	newContent := string(dstContent) + "\n" + subscriptionSection + "\n"
	return os.WriteFile(dstPath, []byte(newContent), 0644)
}

// GetNextMigrationNumber returns the next available migration number
func GetNextMigrationNumber() (int, error) {
	migrationsDir := filepath.Join("app", "service-core", "storage", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, err
	}

	var numbers []int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) >= 1 {
				num, _ := strconv.Atoi(parts[0])
				numbers = append(numbers, num)
			}
		}
	}

	if len(numbers) == 0 {
		return 1, nil
	}

	sort.Ints(numbers)
	return numbers[len(numbers)-1] + 1, nil
}
