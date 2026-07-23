package repo

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
)

const (
	zipFileName     = "gofast-app.zip"
	extractedPrefix = "gofast-live-gofast-app-"
)

func DownloadRepo(email string, apiKey string, projectDir string) error {
	parent := filepath.Dir(projectDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("error creating parent directory %q: %w", parent, err)
	}

	if os.Getenv("TEST") == "true" {
		cmd := exec.Command("cp", "-r", "/home/mat/projects/gofast-app", projectDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error copying test app: %w", err)
		}
		return nil
	}

	succeeded := false
	defer func() {
		if !succeeded {
			cleanupDownloadArtifacts(".")
		}
	}()

	if err := getFile(email, apiKey); err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}
	if err := unzipFile(); err != nil {
		return fmt.Errorf("error unzipping file: %w", err)
	}
	if err := os.Remove(zipFileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing zip file: %w", err)
	}

	extracted, err := findExtractedDir(".")
	if err != nil {
		return err
	}
	if err := os.Rename(extracted, projectDir); err != nil {
		return fmt.Errorf("error renaming template to %q: %w", projectDir, err)
	}

	succeeded = true
	return nil
}

func cleanupDownloadArtifacts(dir string) {
	_ = os.Remove(filepath.Join(dir, zipFileName))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), extractedPrefix) {
			_ = os.RemoveAll(filepath.Join(dir, entry.Name()))
		}
	}
}

func findExtractedDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading directory %q: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), extractedPrefix) {
			return filepath.Join(dir, entry.Name()), nil
		}
	}
	return "", fmt.Errorf("extracted template directory with prefix %q not found", extractedPrefix)
}

func getFile(email string, apiKey string) error {
	client := http.Client{}
	req, err := http.NewRequest("GET", config.SERVER_URL+"/v2?email="+email, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading file: %s", resp.Status)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	file, err := os.OpenFile(zipFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()
	if _, err = io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("error copying response body to file: %w", err)
	}
	return nil
}

func unzipFile() error {
	if os.Getenv("TEST") == "true" {
		return nil
	}
	archive, err := zip.OpenReader(zipFileName)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			fmt.Printf("error closing archive: %v\n", err)
		}
	}()
	for _, file := range archive.File {
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file in zip: %w", err)
		}

		if file.FileInfo().IsDir() {
			if err := src.Close(); err != nil {
				fmt.Printf("error closing source file: %v\n", err)
			}
			if err := os.MkdirAll(file.Name, os.ModePerm); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(file.Name), os.ModePerm); err != nil {
			_ = src.Close()
			return fmt.Errorf("error creating parent directory: %w", err)
		}

		dst, err := os.Create(file.Name)
		if err != nil {
			_ = src.Close()
			return fmt.Errorf("error creating destination file: %w", err)
		}

		_, copyErr := io.Copy(dst, src)
		closeSrcErr := src.Close()
		closeDstErr := dst.Close()
		if copyErr != nil {
			return fmt.Errorf("error copying file from zip: %w", copyErr)
		}
		if closeSrcErr != nil {
			fmt.Printf("error closing source file: %v\n", closeSrcErr)
		}
		if closeDstErr != nil {
			fmt.Printf("error closing destination file: %v\n", closeDstErr)
		}
	}
	return nil
}
