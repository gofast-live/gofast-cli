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
	var route_file_path string
	if protocol == "HTTP" {
		server = "go-http"
		route_file_path = projectName + "/go-http/http/routes.go"
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
		route_file_path = projectName + "/go-grpc/grpc/routes.go"
		err = os.RemoveAll(projectName + "/svelte-http")
		if err != nil {
			return err
		}
		err = os.RemoveAll(projectName + "/go-http")
		if err != nil {
			return err
		}
	}

	route_file, err := os.ReadFile(route_file_path)
	if err != nil {
		return err
	}
	route_file_str := string(route_file)

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
	if database != "PostgreSQL" {
		lines := strings.Split(docker_compose_file_str, "\n")
		new_lines := lines[:len(lines)-10]
		docker_compose_file_str = strings.Join(new_lines, "\n")
	}

	if paymentsProvider == "None" {
		err = os.RemoveAll(projectName + "/" + server + "/service/payment")
		if err != nil {
			return err
		}
		lines := strings.Split(route_file_str, "\n")
		var new_lines []string
		for i := range lines {
			if i >= 33 && i <= 47 || i == 7 {
				continue
			}
			new_lines = append(new_lines, lines[i])
		}
		route_file_str = strings.Join(new_lines, "\n")
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
	err = os.WriteFile(route_file_path, []byte(route_file_str), 0644)
	if err != nil {
		return err
	}
	return nil
}
