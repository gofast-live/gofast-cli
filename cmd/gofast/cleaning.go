package main

import (
	"os"
	"strings"
)

func cleaning(projectName string, client string, start string, databaseProvider string, paymentsProvider string, emailProvider string, filesProvider string, selectedMonitoring string) ([]string, error) {
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

	// Client
	if client == "None" {
		_ = os.RemoveAll(projectName + "/service-svelte")
		_ = os.RemoveAll(projectName + "/service-next")
		_ = os.RemoveAll(projectName + "/service-vue")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  postgres:", false)
	} else if client == "SvelteKit" {
		_ = os.RemoveAll(projectName + "/service-next")
		_ = os.RemoveAll(projectName + "/service-vue")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  next:", "  postgres:", false)
	} else if client == "Next.js" {
		_ = os.RemoveAll(projectName + "/service-svelte")
		_ = os.RemoveAll(projectName + "/service-vue")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  next:", false)
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  vue:", "  postgres:", false)
		docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "3003", "3000")
		docker_compose_lines = strings.Split(docker_compose_file_str, "\n")
	} else if client == "Vue.js" {
		_ = os.RemoveAll(projectName + "/service-svelte")
		_ = os.RemoveAll(projectName + "/service-next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  vue:", false)
		docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "3004", "3000")
		docker_compose_lines = strings.Split(docker_compose_file_str, "\n")
	}

	docker_compose_file_str = strings.Join(docker_compose_lines, "\n")

	// Base project
	var run_cmd []string
	if start == "Generate base project (Local PostgreSQL, Mocked payments, Log Emails, Local files)" {
		run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=Iv23litoS0DJltaklISr \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=c6ed4d8bc5bcb687162da0ea0d9bc614e31004a8 \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_ID=646089287190-m252eqv203c3fsv1gt1m29nkq2t6lrp6.apps.googleusercontent.com \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_SECRET=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 \\")
		run_cmd = append(run_cmd, "DATABASE_PROVIDER=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_HOST=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_PORT=5432 \\")
		run_cmd = append(run_cmd, "POSTGRES_DB=db \\")
		run_cmd = append(run_cmd, "POSTGRES_USER=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_PASSWORD=postgres \\")
		run_cmd = append(run_cmd, "PAYMENT_PROVIDER=local \\")
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=local \\")
		run_cmd = append(run_cmd, "EMAIL_FROM=admin@gofast.live \\")
		run_cmd = append(run_cmd, "FILE_PROVIDER=local \\")
		run_cmd = append(run_cmd, "LOCAL_FILE_DIR=/file \\")
		run_cmd = append(run_cmd, "docker compose up --build")
		readme_file, _ := os.ReadFile(projectName + "/README.md")
		readme_file_lines := strings.Split(string(readme_file), "\n")
		readme_file_lines = append(readme_file_lines, "Generate new JWT keys for the project:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, "sh scripts/keys.sh")
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "Compile the SQL queries using sqlc:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, "sh scripts/sqlc.sh")
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "Spin up the project:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, run_cmd...)
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "Run the Atlas migrations:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, "sh scripts/atlas.sh")
		readme_file_lines = append(readme_file_lines, "```")
		if databaseProvider == "Turso" {
			readme_file_lines = append(readme_file_lines, "Turso needs TURSO_URL and TURSO_TOKEN environment variables to be set.")
		}
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "Access the project at:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, "# Client")
		readme_file_lines = append(readme_file_lines, "http://localhost:3000")
		readme_file_lines = append(readme_file_lines, "# Admin")
		readme_file_lines = append(readme_file_lines, "http://localhost:3001")
		readme_file_lines = append(readme_file_lines, "# PgAdmin")
		readme_file_lines = append(readme_file_lines, "http://localhost:5050")
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "For Grafana Monitoring, check the README.md in `/grafana` folder.  ")
		readme_file_lines = append(readme_file_lines, "For Kubernetes Deployment + Monitoring, check the README.md in `/kube` folder.")
		readme_file_str := strings.Join(readme_file_lines, "\n")
		err = os.WriteFile(projectName+"/README.md", []byte(readme_file_str), 0644)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
		if err != nil {
			return nil, err
		}
		return run_cmd, nil
	} else {
		run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_ID=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_SECRET=__CHANGE_ME__ \\")
	}

	// Database
	if databaseProvider == "PostgreSQL" {
		run_cmd = append(run_cmd, "DATABASE_PROVIDER=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_HOST=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_PORT=5432 \\")
		run_cmd = append(run_cmd, "POSTGRES_DB=db \\")
		run_cmd = append(run_cmd, "POSTGRES_USER=postgres \\")
		run_cmd = append(run_cmd, "POSTGRES_PASSWORD=postgres \\")
	} else if databaseProvider == "SQLite" {
		run_cmd = append(run_cmd, "DATABASE_PROVIDER=sqlite \\")
	} else if databaseProvider == "Turso" {
		run_cmd = append(run_cmd, "DATABASE_PROVIDER=turso \\")
		run_cmd = append(run_cmd, "TURSO_URL=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "TURSO_TOKEN=__CHANGE_ME__ \\")
	}

	if databaseProvider == "SQLite" || databaseProvider == "Turso" {
		// Modify scripts/sqlc.sh
		sqlcScriptPath := projectName + "/scripts/sqlc.sh"
		sqlcScriptBytes, err := os.ReadFile(sqlcScriptPath)
		if err != nil {
			return nil, err
		}
		sqlcScriptContent := string(sqlcScriptBytes)
		sqlcScriptContent = strings.ReplaceAll(sqlcScriptContent, "DB_TYPE=${1:-postgres}", "DB_TYPE=${1:-sqlite}")
		err = os.WriteFile(sqlcScriptPath, []byte(sqlcScriptContent), 0644)
		if err != nil {
			return nil, err
		}

		// Modify scripts/atlas.sh
		atlasScriptPath := projectName + "/scripts/atlas.sh"
		atlasScriptBytes, err := os.ReadFile(atlasScriptPath)
		if err != nil {
			return nil, err
		}
		atlasScriptContent := string(atlasScriptBytes)
		if databaseProvider == "Turso" {
			atlasScriptContent = strings.ReplaceAll(atlasScriptContent, "DB_TYPE=${1:-postgres}", "DB_TYPE=${1:-turso}")
		} else {
			atlasScriptContent = strings.ReplaceAll(atlasScriptContent, "DB_TYPE=${1:-postgres}", "DB_TYPE=${1:-sqlite}")
		}
		err = os.WriteFile(atlasScriptPath, []byte(atlasScriptContent), 0644)
		if err != nil {
			return nil, err
		}

		// Modify service-go-user/dmain/note/service.go
		serviceGoUserPath := projectName + "/service-go-user/domain/note/service.go"
		serviceGoUserBytes, err := os.ReadFile(serviceGoUserPath)
		if err != nil {
			return nil, err
		}
		serviceGoUserContent := string(serviceGoUserBytes)
		serviceGoUserContent = strings.ReplaceAll(serviceGoUserContent, "Limit:  limit,", "Limit:  int64(limit),")
		serviceGoUserContent = strings.ReplaceAll(serviceGoUserContent, "Offset: (page - 1) * limit,", "Offset: int64((page - 1) * limit),")
		err = os.WriteFile(serviceGoUserPath, []byte(serviceGoUserContent), 0644)
		if err != nil {
			return nil, err
		}


		// Modify docker-compose.yml for SQLite or Turso, remove postgres service, starts at "  postgres:", end at "      POSTGRES_DB: db", ai!
	}

	// Payments
	if paymentsProvider == "Local (mock)" {
		run_cmd = append(run_cmd, "PAYMENT_PROVIDER=local \\")
	} else if paymentsProvider == "Stripe" {
		run_cmd = append(run_cmd, "PAYMENT_PROVIDER=stripe \\")
		run_cmd = append(run_cmd, "STRIPE_API_KEY=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "STRIPE_PRICE_ID_BASIC=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "STRIPE_PRICE_ID_PREMIUM=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "STRIPE_WEBHOOK_SECRET=__CHANGE_ME__ \\")
	}

	// Emails
	if emailProvider == "Local (log)" {
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=local \\")
	} else if emailProvider == "Postmark" {
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=postmark \\")
		run_cmd = append(run_cmd, "POSTMARK_API_KEY=__CHANGE_ME__ \\")
	} else if emailProvider == "Sendgrid" {
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=sendgrid \\")
		run_cmd = append(run_cmd, "SENDGRID_API_KEY=__CHANGE_ME__ \\")
	} else if emailProvider == "Resend" {
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=resend \\")
		run_cmd = append(run_cmd, "RESEND_API_KEY=__CHANGE_ME__ \\")
	} else if emailProvider == "AWS SES" {
		run_cmd = append(run_cmd, "EMAIL_PROVIDER=ses \\")
		run_cmd = append(run_cmd, "SES_REGION=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "SES_ACCESS_KEY=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "SES_SECRET_KEY=__CHANGE_ME__ \\")
	}
	run_cmd = append(run_cmd, "EMAIL_FROM=__CHANGE_ME__ \\")

	// Files
	if filesProvider == "Local (folder)" {
		run_cmd = append(run_cmd, "FILE_PROVIDER=local \\")
		run_cmd = append(run_cmd, "LOCAL_FILE_DIR=/file \\")
	} else if filesProvider == "AWS S3" {
		run_cmd = append(run_cmd, "FILE_PROVIDER=s3 \\")
		run_cmd = append(run_cmd, "S3_REGION=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "S3_ACCESS_KEY=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "S3_SECRET_KEY=__CHANGE_ME__ \\")
	} else if filesProvider == "Cloudflare R2" {
		run_cmd = append(run_cmd, "FILE_PROVIDER=r2 \\")
		run_cmd = append(run_cmd, "R2_ENDPOINT=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "R2_ACCESS_KEY=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "R2_SECRET_KEY=__CHANGE_ME__ \\")
	} else if filesProvider == "Google Cloud Storage" {
		run_cmd = append(run_cmd, "FILE_PROVIDER=gcs \\")
		run_cmd = append(run_cmd, "GOOGLE_APPLICATION_CREDENTIALS=__CHANGE_ME__ \\")
	} else if filesProvider == "Azure Blob Storage" {
		run_cmd = append(run_cmd, "FILE_PROVIDER=azblob \\")
		run_cmd = append(run_cmd, "AZBLOB_ACCOUNT_NAME=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "AZBLOB_ACCOUNT_KEY=__CHANGE_ME__ \\")
	}
	if filesProvider != "Local" {
		run_cmd = append(run_cmd, "BUCKET_NAME=__CHANGE_ME__ \\")
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return nil, err
	}

	// Monitoring
	if selectedMonitoring == "Kubernetes + VictoriaMetrics Monitoring" {
		_ = os.RemoveAll(projectName + "/grafana")
	} else if selectedMonitoring == "Grafana + Loki + Prometheus Monitoring using Docker" {
		_ = os.RemoveAll(projectName + "/kube")
	} else {
		_ = os.RemoveAll(projectName + "/kube")
		_ = os.RemoveAll(projectName + "/grafana")
	}

	// Append the cmd to Readme
	run_cmd = append(run_cmd, "docker compose up --build")
	readme_file, _ := os.ReadFile(projectName + "/README.md")
	readme_file_lines := strings.Split(string(readme_file), "\n")
	readme_file_lines = append(readme_file_lines, "Generate new JWT keys for the project:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, "sh scripts/keys.sh")
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "Compile the SQL queries using sqlc:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, "sh scripts/sqlc.sh")
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "Spin up the project:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, run_cmd...)
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "Run the Atlas migrations:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, "sh scripts/atlas.sh")
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "Access the project at:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, "# Client")
	readme_file_lines = append(readme_file_lines, "http://localhost:3000")
	readme_file_lines = append(readme_file_lines, "# Admin")
	readme_file_lines = append(readme_file_lines, "http://localhost:3001")
	readme_file_lines = append(readme_file_lines, "# PgAdmin")
	readme_file_lines = append(readme_file_lines, "http://localhost:5050")
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "For Grafana Monitoring, check the README.md in `/grafana` folder.  ")
	readme_file_lines = append(readme_file_lines, "For Kubernetes Deployment + Monitoring, check the README.md in `/kube` folder.")
	readme_file_str := strings.Join(readme_file_lines, "\n")
	err = os.WriteFile(projectName+"/README.md", []byte(readme_file_str), 0644)
	if err != nil {
		return nil, err
	}
	return run_cmd, nil
}

func remove_lines_from_to(lines []string, from string, to string, removeTo bool) []string {
	var new_lines []string
	var found bool
	for i, line := range lines {
		if strings.Contains(line, from) {
			found = true
		}
		if strings.Contains(line, to) {
			found = false
			if removeTo {
				continue
			}
		}
		if !found {
			new_lines = append(new_lines, lines[i])
		}
	}
	return new_lines
}
