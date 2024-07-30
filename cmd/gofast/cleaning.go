package main

import (
	"os"
	"strings"
)

// TODO
// Clean ports on docker-compose.yml
// Clean github actions linting per setting
func cleaning(projectName string, protocol string, client string, database string, paymentsProvider string, emailProvider string, filesProvider string) error {
	// remove .git folder
	_ = os.RemoveAll(projectName + "/.git")

	var err error
	docker_compose_file, err := os.ReadFile(projectName + "/docker-compose.yml")
	if err != nil {
		return err
	}
	docker_compose_file_str := string(docker_compose_file)
	docker_compose_lines := strings.Split(docker_compose_file_str, "\n")

	// Protocol
	var route_file_path string
	if protocol == "HTTP" {
		route_file_path = projectName + "/go/http/route.go"
		_ = os.RemoveAll(projectName + "/proto.sh")
		_ = os.RemoveAll(projectName + "/proto")
		_ = os.RemoveAll(projectName + "/go/proto")
		_ = os.RemoveAll(projectName + "/go/grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/proto")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/server/grpc.ts")
		for _, file := range []string{"email_service_grpc.ts", "note_service_grpc.ts", "payment_service_grpc.ts", "user_service_grpc.ts"} {
			_ = os.RemoveAll(projectName + "/svelte/src/lib/services/" + file)
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
		route_file_path = projectName + "/go/grpc/route.go"
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
			if strings.Contains(line, "\"io\"") || strings.Contains(line, "\"strconv\"") || strings.Contains(line, "\"server/service/email\"") || strings.Contains(line, "\"server/service/file\"") || strings.Contains(line, "\"server/service/note\"") || strings.Contains(line, "\"server/service/payment\"") || strings.Contains(line, "\"server/service/user\"") {
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

	var run_cmd []string
	run_cmd = append(run_cmd, "JWT_SECRET=gofast_is_the_best \\\n")
	run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=Iv23litoS0DJltaklISr \\\n")
	run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=c6ed4d8bc5bcb687162da0ea0d9bc614e31004a8 \\\n")

	docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
	// Database
	if database != "SQLite" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "SQLITE_FILE: ./local.db", "# SQLITE_FILE: ./local.db")
	}
	if database == "Memory" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: memory")
	} else if database == "Turso" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: turso")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_URL: ${TURSO_URL}", "TURSO_URL: ${TURSO_URL}")
		run_cmd = append(run_cmd, "TURSO_URL=TURSO_URL \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_TOKEN: ${TURSO_TOKEN}", "TURSO_TOKEN: ${TURSO_TOKEN}")
		run_cmd = append(run_cmd, "TURSO_TOKEN=TURSO_TOKEN \\\n")
	} else if database == "PostgreSQL" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: gofast", "POSTGRES_PASS: gofast")
		run_cmd = append(run_cmd, "POSTGRES_PASS=POSTGRES_PASS \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: gofast", "POSTGRES_USER: gofast")
		run_cmd = append(run_cmd, "POSTGRES_USER=POSTGRES_USER \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_NAME: gofast", "POSTGRES_NAME: gofast")
		run_cmd = append(run_cmd, "POSTGRES_NAME=POSTGRES_NAME \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: postgres", "POSTGRES_HOST: postgres")
		run_cmd = append(run_cmd, "POSTGRES_HOST=POSTGRES_HOST \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: 5432", "POSTGRES_PORT: 5432")
		run_cmd = append(run_cmd, "POSTGRES_PORT=POSTGRES_PORT \\\n")
	}
	if database != "PostgreSQL" {
		lines := strings.Split(docker_compose_file_str, "\n")
		new_lines := lines[:len(lines)-10]
		docker_compose_file_str = strings.Join(new_lines, "\n")
	}

	route_file, _ := os.ReadFile(route_file_path)
	route_file_str := string(route_file)
	route_lines := strings.Split(route_file_str, "\n")

	// Payments
	if paymentsProvider == "None" {
		_ = os.RemoveAll(projectName + "/go/service/payment")
		route_lines = remove_line(route_lines, "\"io\"")
		route_lines = remove_line(route_lines, "\"server/service/payment\"")
		route_lines = remove_lines_from_to(route_lines, "// Payment Routes", "// Note Routes")
	} else if paymentsProvider == "Stripe" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_ENABLED: false", "PAYMENT_ENABLED: true")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_API_KEY: ${STRIPE_API_KEY}", "STRIPE_API_KEY: ${STRIPE_API_KEY}")
		run_cmd = append(run_cmd, "STRIPE_API_KEY=STRIPE_API_KEY \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}", "STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}")
		run_cmd = append(run_cmd, "STRIPE_PRICE_ID=STRIPE_PRICE_ID \\\n")
	} else if paymentsProvider == "Lemon Squeezy (not implemented)" {
		// TODO: Implement Lemon Squeezy
		return nil
	}
	// Emails
	if emailProvider == "None" {
		_ = os.RemoveAll(projectName + "/go/service/email")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_ENABLED: true", "EMAIL_ENABLED: false")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "# EMAIL_PROVIDER: local")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_FROM: ${EMAIL_FROM}", "# EMAIL_FROM: ${EMAIL_FROM}")
		route_lines = remove_line(route_lines, "\"server/service/email\"")
		route_lines = remove_lines_from_to(route_lines, "// Email Routes", "// File Routes")
	} else if emailProvider == "Postmark" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: postmark")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTMARK_API_KEY: ${POSTMARK_API_KEY}", "POSTMARK_API_KEY: ${POSTMARK_API_KEY}")
		run_cmd = append(run_cmd, "POSTMARK_API_KEY=POSTMARK_API_KEY \\\n")
	} else if emailProvider == "Sendgrid" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: sendgrid")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SENDGRID_API_KEY: ${SENDGRID_API_KEY}", "SENDGRID_API_KEY: ${SENDGRID_API_KEY}")
		run_cmd = append(run_cmd, "SENDGRID_API_KEY=SENDGRID_API_KEY \\\n")
	} else if emailProvider == "Resend" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: resend")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# RESEND_API_KEY: ${RESEND_API_KEY}", "RESEND_API_KEY: ${RESEND_API_KEY}")
		run_cmd = append(run_cmd, "RESEND_API_KEY=RESEND_API_KEY \\\n")
	}
	// Files
	if filesProvider == "None" {
		// TODO: if gRPC, remove whole http section
		_ = os.RemoveAll(projectName + "/go/service/file")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_ENABLED: true", "FILE_ENABLED: false")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "# FILE_PROVIDER: local")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_DIR: ./files", "# FILE_DIR: ./files")
		if protocol == "HTTP" {
			_ = os.RemoveAll(projectName + "/go/http/file_route.go")
			server_file_lines := strings.Split(projectName+"/go/http/server.go", "\n")
			server_file_lines = remove_line(server_file_lines, "setupFileRoute")
			_ = os.WriteFile(projectName+"/go/http/server.go", []byte(strings.Join(server_file_lines, "\n")), 0644)
		} else if protocol == "gRPC" {
			_ = os.RemoveAll(projectName + "/go/http")
			mainFileContent, _ := os.ReadFile(projectName + "/go/main.go")
			main_file_lines := strings.Split(string(mainFileContent), "\n")
			main_file_lines = remove_line(main_file_lines, "\"server/http\"")
			main_file_lines = remove_line(main_file_lines, "Run the HTTP server")
			main_file_lines = remove_line(main_file_lines, "http.RunHTTP")
			_ = os.WriteFile(projectName+"/go/main.go", []byte(strings.Join(main_file_lines, "\n")), 0644)
		}

	} else if filesProvider == "AWS S3" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: s3")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=BUCKET_NAME \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_REGION: ${S3_REGION}", "S3_REGION: ${S3_REGION}")
		run_cmd = append(run_cmd, "S3_REGION=S3_REGION \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_ACCESS_KEY: ${S3_ACCESS_KEY}", "S3_ACCESS_KEY: ${S3_ACCESS_KEY}")
		run_cmd = append(run_cmd, "S3_ACCESS_KEY=S3_ACCESS_KEY \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_SECRET_KEY: ${S3_SECRET_KEY}", "S3_SECRET_KEY: ${S3_SECRET_KEY}")
		run_cmd = append(run_cmd, "S3_SECRET_KEY=S3_SECRET_KEY \\\n")
	} else if filesProvider == "Cloudflare R2" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: r2")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=BUCKET_NAME \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_ENDPOINT: ${R2_ENDPOINT}", "R2_ENDPOINT: ${R2_ENDPOINT}")
		run_cmd = append(run_cmd, "R2_ENDPOINT=R2_ENDPOINT \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_ACCESS_KEY: ${R2_ACCESS_KEY}", "R2_ACCESS_KEY: ${R2_ACCESS_KEY}")
		run_cmd = append(run_cmd, "R2_ACCESS_KEY=R2_ACCESS_KEY \\\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# R2_SECRET_KEY: ${R2_SECRET_KEY}", "R2_SECRET_KEY: ${R2_SECRET_KEY}")
		run_cmd = append(run_cmd, "R2_SECRET_KEY=R2_SECRET_KEY \\\n")
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return err
	}
	route_file_str = strings.Join(route_lines, "\n")
	err = os.WriteFile(route_file_path, []byte(route_file_str), 0644)
	if err != nil {
		return err
	}

	// Append the cmd to Readme
	run_cmd = append(run_cmd, "docker compose up --build")
	readme_file, _ := os.ReadFile(projectName + "/README.md")
	readme_file_lines := strings.Split(string(readme_file), "\n")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, run_cmd...)
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_str := strings.Join(readme_file_lines, "\n")
	err = os.WriteFile(projectName+"/README.md", []byte(readme_file_str), 0644)
	if err != nil {
		return err
	}
	return nil
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
