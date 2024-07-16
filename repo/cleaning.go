package repo

import (
	"os"
	"strings"
)

// TODO
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

	// Client
	if client == "None" {
		_ = os.RemoveAll(projectName + "/svelte")
		_ = os.RemoveAll(projectName + "/next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "svelte:", "go:")
	} else if client == "SvelteKit" {
		_ = os.RemoveAll(projectName + "/next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "next:", "go:")
	} else if client == "Next.js" {
		_ = os.RemoveAll(projectName + "/svelte")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "svelte:", "next:")
	}

	// Protocol
	var route_file_path string
	if protocol == "HTTP" {
		route_file_path = projectName + "/go/http/route.go"
		_ = os.RemoveAll(projectName + "/proto.sh")
		_ = os.RemoveAll(projectName + "/proto")
		_ = os.RemoveAll(projectName + "/go/proto")
		_ = os.RemoveAll(projectName + "/go/grpc")
		mainFileContent, _ := os.ReadFile(projectName + "/go/main.go")
		lines := strings.Split(string(mainFileContent), "\n")
		var new_lines []string
		for i, line := range lines {
			if strings.Contains(line, "\"server/grpc\"") || strings.Contains(line, "grpc.RunGRPC") || strings.Contains(line, "Run the gRPC server") {
				continue
			}
			new_lines = append(new_lines, lines[i])
		}
		_ = os.WriteFile(projectName+"/go/main.go", []byte(strings.Join(new_lines, "\n")), 0644)
	} else if protocol == "gRPC" {
		route_file_path = projectName + "/go/grpc/route.go"
		// TODO: Implement gRPC
	}

	if protocol == "HTTP" && client == "SvelteKit" {
		_ = os.RemoveAll(projectName + "/svelte/src/routes/(app)/notes_grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/routes/(app)/emails_grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/routes/(app)/files_grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/routes/(app)/billing_grpc")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/services/user_service_grpc.ts")
		_ = os.RemoveAll(projectName + "/svelte/src/lib/services/note_service_grpc.ts")
        _ = os.RemoveAll(projectName + "/svelte/src/lib/services/email_service_grpc.ts")
        _ = os.RemoveAll(projectName + "/svelte/src/lib/services/file_service_grpc.ts")
        _ = os.RemoveAll(projectName + "/svelte/src/lib/services/payment_service_grpc.ts")

	}

	docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
	// Database
	if database != "SQLite" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "SQLITE_FILE: local.db", "# SQLITE_FILE: local.db")
	}
	if database == "Memory" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: memory")
	} else if database == "Turso" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: turso")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_URL: ${TURSO_URL}", "TURSO_URL: ${TURSO_URL}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_TOKEN: ${TURSO_TOKEN}", "TURSO_TOKEN: ${TURSO_TOKEN}")
	} else if database == "PostgreSQL" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: gofast", "POSTGRES_PASS: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: gofast", "POSTGRES_USER: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_NAME: gofast", "POSTGRES_NAME: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: db-postgres", "POSTGRES_HOST: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: 5432", "POSTGRES_PORT: 5432")
	}
	if database != "PostgreSQL" {
		lines := strings.Split(docker_compose_file_str, "\n")
		new_lines := lines[:len(lines)-10]
		docker_compose_file_str = strings.Join(new_lines, "\n")
	}

	route_file, _ := os.ReadFile(route_file_path)
	route_file_str := string(route_file)
	lines := strings.Split(route_file_str, "\n")
	var new_routes []string

	// Payments
	if paymentsProvider == "None" {
		_ = os.RemoveAll(projectName + "/go/service/payment")
		new_routes = remove_line(lines, "\"io\"")
		new_routes = remove_line(new_routes, "\"server/service/payment\"")
		new_routes = remove_lines_from_to(new_routes, "// Payment routes", "// Note routes")
	} else if paymentsProvider == "Stripe" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_ENABLED: false", "PAYMENT_ENABLED: true")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_API_KEY: ${STRIPE_API_KEY}", "STRIPE_API_KEY: ${STRIPE_API_KEY}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}", "STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}")
	} else if paymentsProvider == "Lemon Squeezy (not implemented)" {
		// TODO: Implement Lemon Squeezy
		return nil
	}
	// Emails
	if emailProvider == "None" {
		_ = os.RemoveAll(projectName + "/go/service/email")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_ENABLED: true", "EMAIL_ENABLED: false")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "# EMAIL_PROVIDER: local")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_FROM: admin@gofast.live", "# EMAIL_FROM: admin@gofast.live")
		new_routes = remove_line(lines, "\"server/service/email\"")
		new_routes = remove_lines_from_to(new_routes, "// Email routes", "// File routes")
	} else if emailProvider == "Postmark" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: postmark")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTMARK_API_KEY: ${POSTMARK_API_KEY}", "POSTMARK_API_KEY: ${POSTMARK_API_KEY}")
	} else if emailProvider == "Sendgrid" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: sendgrid")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SENDGRID_API_KEY: ${SENDGRID_API_KEY}", "SENDGRID_API_KEY: ${SENDGRID_API_KEY}")
	} else if emailProvider == "Resend" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: resend")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# RESEND_API_KEY: ${RESEND_API_KEY}", "RESEND_API_KEY: ${RESEND_API_KEY}")
	}
	// Files
	if filesProvider == "None" {
		_ = os.RemoveAll(projectName + "/go/service/file")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_ENABLED: true", "FILE_ENABLED: false")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "# FILE_PROVIDER: local")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_DIR: ./files", "# FILE_DIR: ./files")
		new_routes = remove_line(lines, "\"server/service/file\"")
		new_routes = remove_lines_from_to(new_routes, "// File routes", "// End of routes")
	} else if filesProvider == "S3/D2" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: s3")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_BUCKET: ${S3_BUCKET}", "S3_BUCKET: ${S3_BUCKET}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_REGION: ${S3_REGION}", "S3_REGION: ${S3_REGION}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_ACCESS_KEY: ${S3_ACCESS_KEY}", "S3_ACCESS_KEY: ${S3_ACCESS_KEY}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_SECRET_KEY: ${S3_SECRET_KEY}", "S3_SECRET_KEY: ${S3_SECRET_KEY}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# S3_ENDPOINT: ${S3_ENDPOINT}", "S3_ENDPOINT: ${S3_ENDPOINT}")
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return err
	}
	route_file_str = strings.Join(new_routes, "\n")
	err = os.WriteFile(route_file_path, []byte(route_file_str), 0644)
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
