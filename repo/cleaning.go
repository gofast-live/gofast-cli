package repo

import (
	"os"
	"strings"
)

func cleaning(projectName string, protocol string, database string, paymentsProvider string, emailProvider string, filesProvider string) error {
	var err error
	if protocol == "HTTP" {
		err = os.RemoveAll(projectName + "/svelte-grpc")
		if err != nil {
			return err
		}
		err = os.RemoveAll(projectName + "/go-grpc")
		if err != nil {
			return err
		}
	} else if protocol == "gRPC" {
		err = os.RemoveAll(projectName + "/svelte-http")
		if err != nil {
			return err
		}
		err = os.RemoveAll(projectName + "/go-http")
		if err != nil {
			return err
		}
	}

	if database == "PostgreSQL" {
		docker_compose_file, err := os.ReadFile(projectName + "/docker-compose.yml")
		if err != nil {
			return err
		}
        docker_compose_file_str := string(docker_compose_file)
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "DB_PROVIDER: sqlite", "DB_PROVIDER: postgres")
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PASS: gofast", "POSTGRES_PASS: gofast")
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_USER: gofast", "POSTGRES_USER: gofast")
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_NAME: gofast", "POSTGRES_NAME: gofast")
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_HOST: db-postgres", "POSTGRES_HOST: db-postgres")
        docker_compose_file_str = strings.ReplaceAll(docker_compose_file_str, "# POSTGRES_PORT: 5432", "POSTGRES_PORT: 5432")
        err = os.WriteFile(projectName + "/docker-compose.yml", []byte(docker_compose_file_str), 0644)
        if err != nil {
            return err
        }
    }
	return nil
}
