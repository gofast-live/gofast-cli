package repo

import (
	"archive/zip"
	"context"
	"errors"
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

// DownloadRepo fetches the template and places it at projectDir, whose parent must already exist.
// The archive is extracted in a temporary directory next to projectDir, so the final rename
// never crosses filesystems and a failed download leaves nothing behind.
func DownloadRepo(ctx context.Context, email string, apiKey string, projectDir string) error {
	if os.Getenv("TEST") == "true" {
		cmd := exec.CommandContext(ctx, "cp", "-r", "/home/mat/projects/gofast-app", projectDir)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error copying test app: %w", err)
		}
		return nil
	}

	workDir, err := os.MkdirTemp(filepath.Dir(projectDir), ".gof-download-*")
	if err != nil {
		return fmt.Errorf("creating download directory: %w", err)
	}
	defer func() {
		err := os.RemoveAll(workDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error removing download directory: %v\n", err)
		}
	}()

	zipPath := filepath.Join(workDir, zipFileName)
	err = getFile(ctx, email, apiKey, zipPath)
	if err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}
	err = unzipFile(zipPath, workDir)
	if err != nil {
		return fmt.Errorf("error unzipping file: %w", err)
	}
	extracted, err := findExtractedDir(workDir)
	if err != nil {
		return err
	}
	err = os.Rename(extracted, projectDir)
	if err != nil {
		return fmt.Errorf("renaming %s to %s: %w", extracted, projectDir, err)
	}
	return nil
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

func getFile(ctx context.Context, email string, apiKey string, zipPath string) error {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.ServerURL+"/v2?email="+email, nil)
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
		err := resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing response body: %v\n", err)
		}
	}()

	file, err := os.OpenFile(zipPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing file: %v\n", err)
		}
	}()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying response body to file: %w", err)
	}
	return nil
}

// maxZipEntrySize caps each extracted file so a malformed archive cannot fill the disk.
const maxZipEntrySize = 512 << 20

func unzipFile(zipPath string, destDir string) error {
	if os.Getenv("TEST") == "true" {
		return nil
	}
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer func() {
		err := archive.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing archive: %v\n", err)
		}
	}()
	for _, file := range archive.File {
		err := extractZipEntry(file, destDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// extractZipEntry writes one archive entry under destDir.
func extractZipEntry(file *zip.File, destDir string) error {
	if !filepath.IsLocal(file.Name) {
		return fmt.Errorf("zip entry %q escapes the project directory", file.Name)
	}
	target := filepath.Join(destDir, file.Name)
	if file.FileInfo().IsDir() {
		err := os.MkdirAll(target, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		return nil
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("error opening file in zip: %w", err)
	}
	defer func() {
		err := src.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing source file: %v\n", err)
		}
	}()

	err = os.MkdirAll(filepath.Dir(target), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating parent directory: %w", err)
	}
	dst, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing destination file: %v\n", err)
		}
	}()

	written, err := io.CopyN(dst, src, maxZipEntrySize+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error copying file from zip: %w", err)
	}
	if written > maxZipEntrySize {
		return fmt.Errorf("zip entry %q is larger than %d bytes", file.Name, maxZipEntrySize)
	}
	return nil
}
