package repo

import (
	"os"
	"strings"
)

func cleaning(projectName string, protocol string, database string, paymentsProvider string, emailProvider string, filesProvider string) error {
	var err error
	docker_compose_file, err := os.ReadFile(projectName + "/docker-compose.yml")
	if err != nil {
		return err
	}
	docker_compose_file_str := string(docker_compose_file)

	var server string
	// var client string
	if protocol == "HTTP" {
		server = "go-http"
		err = os.RemoveAll(projectName + "/svelte-grpc")
		if err != nil {
			return err
		}
		err = os.RemoveAll(projectName + "/go-grpc")
		if err != nil {
			return err
		}
	} else if protocol == "gRPC" {
		server = "go-grpc"
		err = os.RemoveAll(projectName + "/svelte-http")
		if err != nil {
			return err
		}
		err = os.RemoveAll(projectName + "/go-http")
		if err != nil {
			return err
		}
	}

	if database == "SQLite" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: memory", "DB_PROVIDER: sqlite")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# SQLITE_FILE: local.db", "SQLITE_FILE: local.db")
	} else if database == "Turso" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: memory", "DB_PROVIDER: turso")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_URL: ${TURSO_URL}", "TURSO_URL: ${TURSO_URL}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# TURSO_TOKEN: ${TURSO_TOKEN}", "TURSO_TOKEN: ${TURSO_TOKEN}")
	} else if database == "PostgreSQL" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: memory", "DB_PROVIDER: postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: gofast", "POSTGRES_PASS: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: gofast", "POSTGRES_USER: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_NAME: gofast", "POSTGRES_NAME: gofast")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: db-postgres", "POSTGRES_HOST: db-postgres")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: 5432", "POSTGRES_PORT: 5432")
	}

	if paymentsProvider == "None" {
		err = os.RemoveAll(projectName + "/" + server + "/service/payments")
		if err != nil {
			return err
		}
	} else if paymentsProvider == "Stripe" {
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "PAYMENT_ENABLED: false", "PAYMENT_ENABLED: true")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_API_KEY: ${STRIPE_API_KEY}", "STRIPE_API_KEY: ${STRIPE_API_KEY}")
		docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}", "STRIPE_PRICE_ID: ${STRIPE_PRICE_ID}")
	} else if paymentsProvider == "Lemon Squeezy (not implemented)" {
		// TODO: Implement Lemon Squeezy
		return nil
	}

	err = os.WriteFile(projectName+"/docker-compose.yml", []byte(docker_compose_file_str), 0644)
	if err != nil {
		return err
	}
	return nil
}
