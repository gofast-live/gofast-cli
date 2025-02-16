package main

import (
	"os"
	"strings"
)

func cleaning(projectName string, client string, start string, paymentsProvider string, emailProvider string, filesProvider string, selectedMonitoring string) ([]string, error) {
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
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "3001", "3000")
		docker_compose_lines = strings.Split(docker_compose_file_str, "\n")
	} else if client == "Vue.js" {
		_ = os.RemoveAll(projectName + "/service-svelte")
		_ = os.RemoveAll(projectName + "/service-next")
		docker_compose_lines = remove_lines_from_to(docker_compose_lines, "  svelte:", "  vue:", false)
		docker_compose_file_str = strings.Join(docker_compose_lines, "\n")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "3002", "3000")
	}

	// Base project
	var run_cmd []string
	if start == "Generate base project (Local PostgreSQL, Grafana Monitoring, Mocked payments, Local files, Log Emails)" {
		run_cmd = append(run_cmd, "GITHUB_CLIENT_ID=Iv23litoS0DJltaklISr \\")
		run_cmd = append(run_cmd, "GITHUB_CLIENT_SECRET=c6ed4d8bc5bcb687162da0ea0d9bc614e31004a8 \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_ID=646089287190-m252eqv203c3fsv1gt1m29nkq2t6lrp6.apps.googleusercontent.com \\")
		run_cmd = append(run_cmd, "GOOGLE_CLIENT_SECRET=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 \\")
        run_cmd = append(run_cmd, "POSTGRES_HOST=postgres \\")
        run_cmd = append(run_cmd, "POSTGRES_PORT=5432 \\")
        run_cmd = append(run_cmd, "POSTGRES_DB=postgres \\")
        run_cmd = append(run_cmd, "POSTGRES_PASS=postgres \\")
        run_cmd = append(run_cmd, "POSTGRES_USER=postgres \\")
		run_cmd = append(run_cmd, "EMAIL_FROM=admin@gofast.live \\")
		run_cmd = append(run_cmd, "docker compose up --build")
		readme_file, _ := os.ReadFile(projectName + "/README.md")
		readme_file_lines := strings.Split(string(readme_file), "\n")
		readme_file_lines = append(readme_file_lines, "Generate new JWT keys for the project:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, "cd scripts && sh key.sh")
		readme_file_lines = append(readme_file_lines, "```")
		readme_file_lines = append(readme_file_lines, "")
		readme_file_lines = append(readme_file_lines, "Spin up the project:")
		readme_file_lines = append(readme_file_lines, "```bash")
		readme_file_lines = append(readme_file_lines, run_cmd...)
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
	docker_compose_file_str = strings.Join(docker_compose_lines, "\n")

	// Payments
	if paymentsProvider == "Stripe" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_PROVIDER: local", "PAYMENT_PROVIDER: stripe")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_API_KEY: ${STRIPE_API_KEY}", "STRIPE_API_KEY: ${STRIPE_API_KEY}")
		run_cmd = append(run_cmd, "STRIPE_API_KEY=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}", "STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}")
		run_cmd = append(run_cmd, "STRIPE_PRICE_ID=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_WEBHOOK_SECRET: ${STRIPE_WEBHOOK_SECRET}", "STRIPE_WEBHOOK_SECRET: ${STRIPE_WEBHOOK_SECRET}")
		run_cmd = append(run_cmd, "STRIPE_WEBHOOK_SECRET=__CHANGE_ME__ \\")
	} else if paymentsProvider == "Lemon Squeezy" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_PROVIDER: local", "PAYMENT_PROVIDER: lemon")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# LEMON_API_KEY: ${LEMON_API_KEY}", "LEMON_API_KEY: ${LEMON_API_KEY}")
		run_cmd = append(run_cmd, "LEMON_API_KEY=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# LEMON_VARIANT_ID: ${LEMON_VARIANT_ID}", "LEMON_VARIANT_ID: ${LEMON_VARIANT_ID}")
		run_cmd = append(run_cmd, "LEMON_VARIANT_ID=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# LEMON_STORE_ID: ${LEMON_STORE_ID}", "LEMON_STORE_ID: ${LEMON_STORE_ID}")
		run_cmd = append(run_cmd, "LEMON_STORE_ID=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# LEMON_WEBHOOK_SECRET: ${LEMON_WEBHOOK_SECRET}", "LEMON_WEBHOOK_SECRET: ${LEMON_WEBHOOK_SECRET}")
		run_cmd = append(run_cmd, "LEMON_WEBHOOK_SECRET=__CHANGE_ME__ \\")
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
	} else if emailProvider == "AWS SES" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "EMAIL_PROVIDER: local", "EMAIL_PROVIDER: ses")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SES_REGION: ${SES_REGION}", "SES_REGION: ${SES_REGION}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SES_ACCESS_KEY: ${SES_ACCESS_KEY}", "SES_ACCESS_KEY: ${SES_ACCESS_KEY}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SES_SECRET_KEY: ${SES_SECRET_KEY}", "SES_SECRET_KEY: ${SES_SECRET_KEY}")
		run_cmd = append(run_cmd, "SES_REGION=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "SES_ACCESS_KEY=__CHANGE_ME__ \\")
		run_cmd = append(run_cmd, "SES_SECRET_KEY=__CHANGE_ME__ \\")
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
	} else if filesProvider == "Google Cloud Storage" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: gcs")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}", "GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}")
		run_cmd = append(run_cmd, "GOOGLE_APPLICATION_CREDENTIALS=__CHANGE_ME__ \\")
	} else if filesProvider == "Azure Blob Storage" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "FILE_PROVIDER: local", "FILE_PROVIDER: azblob")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# BUCKET_NAME: ${BUCKET_NAME}", "BUCKET_NAME: ${BUCKET_NAME}")
		run_cmd = append(run_cmd, "BUCKET_NAME=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# AZBLOB_ACCOUNT_NAME: ${AZBLOB_ACCOUNT_NAME}", "AZBLOB_ACCOUNT_NAME: ${AZBLOB_ACCOUNT_NAME}")
		run_cmd = append(run_cmd, "AZBLOB_ACCOUNT_NAME=__CHANGE_ME__ \\")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# AZBLOB_ACCOUNT_KEY: ${AZBLOB_ACCOUNT_KEY}", "AZBLOB_ACCOUNT_KEY: ${AZBLOB_ACCOUNT_KEY}")
		run_cmd = append(run_cmd, "AZBLOB_ACCOUNT_KEY=__CHANGE_ME__ \\")
	}

	// Monitoring
	lines := strings.Split(docker_compose_file_str, "\n")
	if selectedMonitoring == "No" {
		_ = os.RemoveAll(projectName + "/grafana")
		_ = os.RemoveAll(projectName + "/kube")
		new_lines := remove_lines_from_to(lines, "logging:", "loki-retries:", true)
		docker_compose_file_str = strings.Join(new_lines, "\n")
	} else if selectedMonitoring == "Grafana + Loki + Prometheus Monitoring using Docker" {
		_ = os.RemoveAll(projectName + "/kube")
	} else {
		_ = os.RemoveAll(projectName + "/grafana")
		new_lines := remove_lines_from_to(lines, "logging:", "loki-retries:", true)
		docker_compose_file_str = strings.Join(new_lines, "\n")
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return nil, err
	}

	// Append the cmd to Readme
    run_cmd = append(run_cmd, "POSTGRES_HOST=postgres \\")
    run_cmd = append(run_cmd, "POSTGRES_PORT=5432 \\")
    run_cmd = append(run_cmd, "POSTGRES_DB=postgres \\")
    run_cmd = append(run_cmd, "POSTGRES_PASS=postgres \\")
    run_cmd = append(run_cmd, "POSTGRES_USER=postgres \\")
	run_cmd = append(run_cmd, "docker compose up --build")
	readme_file, _ := os.ReadFile(projectName + "/README.md")
	readme_file_lines := strings.Split(string(readme_file), "\n")
	readme_file_lines = append(readme_file_lines, "Generate new JWT keys for the project:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, "cd scripts && sh key.sh")
	readme_file_lines = append(readme_file_lines, "```")
	readme_file_lines = append(readme_file_lines, "")
	readme_file_lines = append(readme_file_lines, "Spin up the project:")
	readme_file_lines = append(readme_file_lines, "```bash")
	readme_file_lines = append(readme_file_lines, run_cmd...)
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
