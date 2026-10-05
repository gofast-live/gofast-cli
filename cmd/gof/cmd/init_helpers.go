package cmd

import (
	"context"
	"errors"
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

func hostPortInUse(ctx context.Context, port int) bool {
	dialer := net.Dialer{Timeout: portProbeTimeout}
	addrs := []string{
		net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		net.JoinHostPort("::1", strconv.Itoa(port)),
	}
	for _, addr := range addrs {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// suggestFreePostgresPort finds a free host port to recommend in error messages.
func suggestFreePostgresPort(ctx context.Context) int {
	for port := defaultPostgresHostPort + 1; port <= defaultPostgresHostPort+100; port++ {
		if !hostPortInUse(ctx, port) {
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

// composeProjectExists reports whether Docker already has containers for a Compose project
// with this name. Compose derives the project name from the lowercased directory basename,
// so a new project would share containers and volumes with the existing one.
func composeProjectExists(ctx context.Context, projectName string) (bool, error) {
	filter := "label=com.docker.compose.project=" + strings.ToLower(projectName)
	list := exec.CommandContext(ctx, "docker", "ps", "--all", "--quiet", "--filter", filter) //nolint:gosec // G204: projectName is validated by projectNamePattern
	output, err := list.Output()
	if err != nil {
		return false, fmt.Errorf("listing docker containers: %w", err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// topmostMissingDir returns the highest ancestor of path (or path itself) that does not exist yet.
func topmostMissingDir(path string) string {
	for {
		parent := filepath.Dir(path)
		_, err := os.Stat(parent)
		if err == nil || parent == path {
			return path
		}
		path = parent
	}
}

// rollbackProject removes what a failed init created: the compose resources when a compose
// step was reached, then rollbackRoot (the project directory, or the topmost parent init created).
func rollbackProject(ctx context.Context, projectDir, rollbackRoot string, composeStarted bool) error {
	var downErr error
	if composeStarted {
		down := exec.CommandContext(ctx, "docker", "compose", "down", "-v")
		down.Dir = projectDir
		output, err := down.CombinedOutput()
		if err != nil {
			downErr = fmt.Errorf("removing compose resources: %w\nOutput: %s", err, output)
		}
	}
	removeErr := os.RemoveAll(rollbackRoot)
	if removeErr != nil {
		removeErr = fmt.Errorf("removing %s: %w", rollbackRoot, removeErr)
	}
	return errors.Join(downErr, removeErr)
}
