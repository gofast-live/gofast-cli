package main

import (
	"os"
	"strings"
)

// TODO
// Clean ports on docker-compose.yml
// Clean github actions linting per setting
// Clean libs on svelte and next
// Clean proto.sh
func cleaning(projectName string, protocol string, client string, start string, database string, paymentsProvider string, emailProvider string, filesProvider string, selectedMonitoring string) ([]string, error) {
	// remove .git folder
	_ = os.RemoveAll(projectName + "/.git")

	var err error
	docker_compose_file, err := os.ReadFile(projectName + "/docker-compose.yml")
	if err != nil {
		return nil, err
	}
	docker_compose_file_str := string(docker_compose_file)
	docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "  container_name: gofast-", "  container_name: "+projectName+"-")
	docker_compose_lines := strings.Split(docker_compose_file_str, "\n")

	// Protocol
	if protocol == "HTTP" {
		_ = os.RemoveAll(projectName + "/proto.sh")
		_ = os.RemoveAll(projectName + "/proto")
		_ = os.RemoveAll(projectName + "/go/proto")
		_ = os.RemoveAll(projectName + "/go/grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/proto")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/server/grpc.ts")
		for _, file := range []string{"email_service_grpc.ts", "note_service_grpc.ts", "payment_service_grpc.ts", "user_service_grpc.ts"} {
			_ = os.RemoveAll(projectName + "/svelte/src/lib/server/services/" + file)
		}
		_ = os.RemoveAll(projectName + "/next/app/lib/proto")
		_ = os.RemoveAll(projectName + "/next/app/lib/server/grpc.ts")
		for _, file := range []string{"email_service_grpc.ts", "note_service_grpc.ts", "payment_service_grpc.ts", "user_service_grpc.ts"} {
			_ = os.RemoveAll(projectName + "/next/app/lib/services/" + file)
		}

		docker_compose_lines = remove_line(docker_compose_lines, "GRPC_PORT")

		// Remove gRPC from main.go
		mainFileContent, _ := os.ReadFile(projectName + "/go/main.go")
		main_file_lines := strings.Split(string(mainFileContent), "\n")
		var new_main_file_lines []string
		for i, line := range main_file_lines {
			if strings.Contains(line, "\"server/grpc\"") || strings.Contains(line, "grpc.RunGRPC") || strings.Contains(line, "Run the gRPC server") {
				continue
			}
			new_main_file_lines = append(new_main_file_lines, main_file_lines[i])
		}
		_ = os.WriteFile(projectName+"/go/main.go", []byte(strings.Join(new_main_file_lines, "\n")), 0644)
	} else if protocol == "gRPC" {
		// Remove HTTP files from Svelte and Next.js
		for _, file := range []string{"email_service_http.ts", "note_service_http.ts", "payment_service_http.ts", "user_service_http.ts"} {
			_ = os.Remove(projectName + "/svelte/src/lib/server/services/" + file)
		}
		for _, file := range []string{"email_service_http.ts", "note_service_http.ts", "payment_service_http.ts", "user_service_http.ts"} {
			_ = os.Remove(projectName + "/next/app/lib/server/services/" + file)
		}

		// Clean HTTP routes
		http_route_file, _ := os.ReadFile(projectName + "/go/http/route.go")
		http_route_file_lines := strings.Split(string(http_route_file), "\n")
		http_route_file_lines = remove_lines_from_to(http_route_file_lines, "// Auth Routes", "// End Routes")
		var new_http_route_file_lines []string
		for i, line := range http_route_file_lines {
			if strings.Contains(line, "\"io\"") || strings.Contains(line, "\"strconv\"") || strings.Contains(line, "\"server/services/email\"") || strings.Contains(line, "\"server/services/file\"") || strings.Contains(line, "\"server/services/note\"") || strings.Contains(line, "\"server/services/payment\"") || strings.Contains(line, "\"server/services/user\"") {
				continue
			}
			new_http_route_file_lines = append(new_http_route_file_lines, http_route_file_lines[i])
		}
		_ = os.WriteFile(projectName+"/go/http/route.go", []byte(strings.Join(new_http_route_file_lines, "\n")), 0644)

		// Replace _http with _grpc in Svelte and Next.js
		replace_http(projectName+"/svelte/src/", []string{
			"hooks.server.ts",
			"routes/auth/+page.server.ts",
			"routes/auth/[provider]/+page.server.ts",
			"routes/(app)/notes/+page.server.ts",
			"routes/(app)/notes/[note_id]/+page.server.ts",
			"routes/(app)/emails/+page.server.ts",
			"routes/(app)/payments/+page.server.ts",
		})
		replace_http(projectName+"/next/app/", []string{
			"auth/auth_form.tsx",
			"auth/[provider]/route.ts",
			"(app)/layout.tsx",
			"(app)/page.tsx",
			"(app)/notes/page.tsx",
			"(app)/notes/insert_note_form.tsx",
			"(app)/notes/[note_id]/page.tsx",
			"(app)/notes/[note_id]/update_note_form.tsx",
			"(app)/notes/[note_id]/delete_note_form.tsx",
			"(app)/emails/page.tsx",
			"(app)/emails/send_email_form.tsx",
			"(app)/payments/page.tsx",
			"(app)/payments/billing_form.tsx",
		})
	}

	// Client
	if client == "None" {
		_ = os.RemoveAll(projectName + "/svelte")
		_ = os.RemoveAll(projectName + "/next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  go:")
	} else if client == "SvelteKit" {
		_ = os.RemoveAll(projectName + "/next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  next:", "  go:")
	} else if client == "Next.js" {
		_ = os.RemoveAll(projectName + "/svelte")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  next:")
		docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "- 3001:3000", "- 3000:3000")
		docker_compose_lines = strings.Split(docker_compose_file_str, "\n")
	}

	// Base project
	var run_cmd []string
	if start == "Generate base project (SQLite, Grafana Monitoring, Mocked payments, Local files, Log Emails)" {
		run_cmd = append(run_cmd, "JWT_SECRET=gofast_is_the_best \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=Iv23litoS0DJltaklISr \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=c6ed4d8bc5bcb687162da0ea0d9bc614e31004a8 \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_ID=646089287190-m252eqv203c3fsv1gt1m29nkq2t6lrp6.apps.googleusercontent.com \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_SECRET=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 \\")
		run_cmd = append(run_cmd, "EMAIL_FROM=admin@gofast.live \\")
		run_cmd = append(run_cmd, "docker compose up --build")
		run_cmd = append(run_cmd, "\n")
		run_cmd = append(run_cmd, "For Grafana Monitoring, check the README.md in `/grafana` folder")
		readme_file, _ := os.ReadFile(projectName + "/README.md")
		readme_file_lines := strings.Split(string(readme_file), "\n")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, run_cmd...)
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_str := strings.Join(readme_file_lines, "\n")
		err = os.WriteFile(projectName+"/README.md", []byte(readme_file_str), 0644)
		if err != nil {
			return nil, err
		}
		docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
		err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
		if err != nil {
			return nil, err
		}
		return run_cmd, nil
	} else {
		run_cmd = append(run_cmd, "JWT_SECRET=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_ID=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_SECRET=__CHANGE_ME__ \\")
	}
	docker_compose_file_str = strings.Join(docker_compose_lines, "\n")

	// Database
	if database != "SQLite" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "SQLITE_FILE: ./storage/local.db", "# SQLITE_FILE: ./storage/local.db")
	}
	if database == "Turso with Embedded Replicas" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: turso")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_URL: ${TURSO_URL}", "TURSO_URL: ${TURSO_URL}")
		run_cmd = append(run_cmd, "TURSO_URL=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_TOKEN: ${TURSO_TOKEN}", "TURSO_TOKEN: ${TURSO_TOKEN}")
		run_cmd = append(run_cmd, "TURSO_TOKEN=__CHANGE_ME__ \\")
	} else if database == "PostgreSQL (local)" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: ${POSTGRES_HOST}", "POSTGRES_HOST: ${POSTGRES_HOST}")
		run_cmd = append(run_cmd, "POSTGRES_HOST=postgres \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: ${POSTGRES_PORT}", "POSTGRES_PORT: ${POSTGRES_PORT}")
		run_cmd = append(run_cmd, "POSTGRES_PORT=5432 \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_DB: ${POSTGRES_DB}", "POSTGRES_DB: ${POSTGRES_DB}")
		run_cmd = append(run_cmd, "POSTGRES_DB=gofast \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: ${POSTGRES_PASS}", "POSTGRES_PASS: ${POSTGRES_PASS}")
		run_cmd = append(run_cmd, "POSTGRES_PASS=gofast \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: ${POSTGRES_USER}", "POSTGRES_USER: ${POSTGRES_USER}")
		run_cmd = append(run_cmd, "POSTGRES_USER=gofast \\")
	} else if database == "PostgreSQL (remote)" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: ${POSTGRES_HOST}", "POSTGRES_HOST: ${POSTGRES_HOST}")
		run_cmd = append(run_cmd, "POSTGRES_HOST=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: ${POSTGRES_PORT}", "POSTGRES_PORT: ${POSTGRES_PORT}")
		run_cmd = append(run_cmd, "POSTGRES_PORT=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_DB: ${POSTGRES_DB}", "POSTGRES_DB: ${POSTGRES_DB}")
		run_cmd = append(run_cmd, "POSTGRES_DB=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: ${POSTGRES_PASS}", "POSTGRES_PASS: ${POSTGRES_PASS}")
		run_cmd = append(run_cmd, "POSTGRES_PASS=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: ${POSTGRES_USER}", "POSTGRES_USER: ${POSTGRES_USER}")
		run_cmd = append(run_cmd, "POSTGRES_USER=__CHANGE_ME__ \\")
	}
	if database != "PostgreSQL (local)" {
		lines := strings.Split(docker_compose_file_str, "\n")
		new_lines := lines[:len(lines)-10]
		docker_compose_file_str = strings.Join(new_lines, "\n")
	}

	// Payments
	if paymentsProvider == "Stripe" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_PROVIDER: local", "PAYMENT_PROVIDER: stripe")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_API_KEY: ${STRIPE_API_KEY}", "STRIPE_API_KEY: ${STRIPE_API_KEY}")
		run_cmd = append(run_cmd, "STRIPE_API_KEY=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}", "STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}")
		run_cmd = append(run_cmd, "STRIPE_PRICE_ID=__CHANGE_ME__ \\")
	} else if paymentsProvider == "Lemon Squeezy (not implemented)" {
		// TODO: Implement Lemon Squeezy
		return nil, nil
	}

	// Emails
	run_cmd = append(run_cmd, "EMAIL_FROM=__CHANGE_ME__ \\")
	if emailProvider == "Postmark" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: postmark")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTMARK_API_KEY: ${POSTMARK_API_KEY}", "POSTMARK_API_KEY: ${POSTMARK_API_KEY}")
		run_cmd = append(run_cmd, "POSTMARK_API_KEY=__CHANGE_ME__ \\")
	} else if emailProvider == "Sendgrid" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: sendgrid")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SENDGRID_API_KEY: ${SENDGRID_API_KEY}", "SENDGRID_API_KEY: ${SENDGRID_API_KEY}")
		run_cmd = append(run_cmd, "SENDGRID_API_KEY=__CHANGE_ME__ \\")
	} else if emailProvider == "Resend" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: resend")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# RESEND_API_KEY: ${RESEND_API_KEY}", "RESEND_API_KEY: ${RESEND_API_KEY}")
		run_cmd = append(run_cmd, "RESEND_API_KEY=__CHANGE_ME__ \\")
	}

	// Files
	if filesProvider == "AWS S3" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: s3")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_REGION: ${S3_REGION}", "S3_REGION: ${S3_REGION}")
		run_cmd = append(run_cmd, "S3_REGION=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_ACCESS_KEY: ${S3_ACCESS_KEY}", "S3_ACCESS_KEY: ${S3_ACCESS_KEY}")
		run_cmd = append(run_cmd, "S3_ACCESS_KEY=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_SECRET_KEY: ${S3_SECRET_KEY}", "S3_SECRET_KEY: ${S3_SECRET_KEY}")
		run_cmd = append(run_cmd, "S3_SECRET_KEY=__CHANGE_ME__ \\")
	} else if filesProvider == "Cloudflare R2" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: r2")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_ENDPOINT: ${R2_ENDPOINT}", "R2_ENDPOINT: ${R2_ENDPOINT}")
		run_cmd = append(run_cmd, "R2_ENDPOINT=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_ACCESS_KEY: ${R2_ACCESS_KEY}", "R2_ACCESS_KEY: ${R2_ACCESS_KEY}")
		run_cmd = append(run_cmd, "R2_ACCESS_KEY=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_SECRET_KEY: ${R2_SECRET_KEY}", "R2_SECRET_KEY: ${R2_SECRET_KEY}")
		run_cmd = append(run_cmd, "R2_SECRET_KEY=__CHANGE_ME__ \\")
	}
	run_cmd = append(run_cmd, "docker compose up --build")

	// Monitoring
	lines := strings.Split(docker_compose_file_str, "\n")
	if selectedMonitoring == "No" {
		_ = os.RemoveAll(projectName + "/grafana")
		// Remove last 34 lines from docker-compose.yml
		var new_lines []string
		if database != "PostgreSQL (local)" {
			new_lines = lines[:len(lines)-33]
		} else {
			ten_last_lines := lines[len(lines)-10:]
			new_lines = lines[:len(lines)-42]
			new_lines = append(new_lines, ten_last_lines...)
		}
		new_lines = remove_lines_from_to(new_lines, "logging:", "command:")
		docker_compose_file_str = strings.Join(new_lines, "\n")
	} else {
		run_cmd = append(run_cmd, "\n")
		run_cmd = append(run_cmd, "For Grafana Monitoring, check the README.md in `/grafana` folder")
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return nil, err
	}

	// Append the cmd to Readme
	readme_file, _ := os.ReadFile(projectName + "/README.md")
	readme_file_lines := strings.Split(string(readme_file), "\n")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, run_cmd...)
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_str := strings.Join(readme_file_lines, "\n")
	err = os.WriteFile(projectName+"/README.md", []byte(readme_file_str), 0644)
	if err != nil {
		return nil, err
	}
	return run_cmd, nil
}

func remove_line(lines []string, to_remove string) []string {
	var new_lines []string
	for _, line := range lines {
		if !strings.Contains(line, to_remove) {
			new_lines = append(new_lines, line)
		}
	}
	return new_lines
}

func remove_lines_from_to(lines []string, from string, to string) []string {
	var new_lines []string
	var found bool
	for i, line := range lines {
		if strings.Contains(line, from) {
			found = true
		}
		if strings.Contains(line, to) {
			found = false
		}
		if !found {
			new_lines = append(new_lines, lines[i])
		}
	}
	return new_lines
}

func replace_http(directory string, files []string) {
	for _, file := range files {
		content, _ := os.ReadFile(directory + file)
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, "_http") {
				lines[i] = strings.ReplaceAll(line, "_http", "_grpc")
			}
		}
		_ = os.WriteFile(directory+"/"+file, []byte(strings.Join(lines, "\n")), 0644)
	}
}
