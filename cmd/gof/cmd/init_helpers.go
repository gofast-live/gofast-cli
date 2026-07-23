package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPostgresHostPort = 5432
	portProbeTimeout        = 200 * time.Millisecond
)

var projectNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// parseProjectArg splits a user arg into destination directory and project name.
func parseProjectArg(arg string) (projectDir string, projectName string, err error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", "", fmt.Errorf("project path cannot be empty")
	}

	if strings.HasPrefix(arg, "~/") {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", "", fmt.Errorf("resolving home directory: %w", homeErr)
		}
		arg = filepath.Join(home, arg[2:])
	} else if arg == "~" {
		return "", "", fmt.Errorf(`project path cannot be "~"; provide a project name or path like ~/myapp`)
	}

	projectDir = filepath.Clean(arg)
	projectName = filepath.Base(projectDir)
	if projectName == "." || projectName == string(filepath.Separator) || projectName == "" {
		return "", "", fmt.Errorf("could not determine project name from %q", arg)
	}
	if !projectNamePattern.MatchString(projectName) {
		return "", "", fmt.Errorf(
			"invalid project name %q: must start with a letter and contain only letters, numbers, underscores, or hyphens",
			projectName,
		)
	}
	return projectDir, projectName, nil
}

func hostPortInUse(port int) bool {
	addrs := []string{
		net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		net.JoinHostPort("::1", strconv.Itoa(port)),
	}
	for _, addr := range addrs {
		conn, err := net.DialTimeout("tcp", addr, portProbeTimeout)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// suggestFreePostgresPort finds a free host port to recommend in error messages.
func suggestFreePostgresPort() int {
	for port := defaultPostgresHostPort + 1; port <= defaultPostgresHostPort+100; port++ {
		if !hostPortInUse(port) {
			return port
		}
	}
	return defaultPostgresHostPort + 1
}

// applyProjectName rewrites only compose container_name prefixes.
func applyProjectName(composeContent, projectName string) string {
	return strings.ReplaceAll(composeContent, "container_name: gofast-", "container_name: "+projectName+"-")
}

// applyHostPostgresPort rewrites the host-published Postgres port and Makefile DSN port.
// Internal container POSTGRES_PORT stays 5432.
// Supports both current templates (port=5432) and legacy templates (no port=).
func applyHostPostgresPort(composeContent, makefileContent string, hostPort int) (string, string, error) {
	if hostPort == defaultPostgresHostPort {
		return composeContent, makefileContent, nil
	}

	oldPublish := fmt.Sprintf("- %d:%d", defaultPostgresHostPort, defaultPostgresHostPort)
	newPublish := fmt.Sprintf("- %d:%d", hostPort, defaultPostgresHostPort)
	if !strings.Contains(composeContent, oldPublish) {
		return "", "", fmt.Errorf("docker-compose.yml missing expected port mapping %q", oldPublish)
	}
	composeContent = strings.Replace(composeContent, oldPublish, newPublish, 1)

	oldDSNPort := fmt.Sprintf("port=%d", defaultPostgresHostPort)
	newDSNPort := fmt.Sprintf("port=%d", hostPort)
	switch {
	case strings.Contains(makefileContent, oldDSNPort):
		makefileContent = strings.Replace(makefileContent, oldDSNPort, newDSNPort, 1)
	case strings.Contains(makefileContent, "host=localhost user=postgres"):
		makefileContent = strings.Replace(
			makefileContent,
			"host=localhost user=postgres",
			fmt.Sprintf("host=localhost port=%d user=postgres", hostPort),
			1,
		)
	default:
		return "", "", fmt.Errorf("Makefile missing expected migrate DSN (port=%d or host=localhost user=postgres)", defaultPostgresHostPort)
	}

	return composeContent, makefileContent, nil
}

func rollbackProject(projectDir string) {
	down := exec.Command("docker", "compose", "down", "-v")
	down.Dir = projectDir
	_ = down.Run()
	_ = os.RemoveAll(projectDir)
}
