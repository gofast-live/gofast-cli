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

func DownloadRepo(ctx context.Context, email string, apiKey string, projectName string) error {
	if os.Getenv("TEST") == "true" {
		cmd := exec.CommandContext(ctx, "cp", "-r", "/home/mat/projects/gofast-app", projectName)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error copying test app: %w", err)
		}
		return nil
	}
	// get the file
	err := getFile(ctx, email, apiKey)
	if err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}
	// unzip the file
	err = unzipFile()
	if err != nil {
		return fmt.Errorf("error unzipping file: %w", err)
	}
	// remove the zip file
	err = os.Remove("gofast-app.zip")
	if err != nil {
		return fmt.Errorf("error removing zip file: %w", err)
	}
	// find and rename the folder `gofast-live-gofast-app-...` to the project name
	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("error reading current directory: %w", err)
	}
	for _, f := range files {
		if f.IsDir() {
			if strings.HasPrefix(f.Name(), "gofast-live-gofast-app-") {
				err = os.Rename(f.Name(), projectName)
				if err != nil {
					return fmt.Errorf("renaming %s to %s: %w", f.Name(), projectName, err)
				}
				break
			}
		}
	}

	return nil
}

func getFile(ctx context.Context, email string, apiKey string) error {
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

	// save the file to the disk
	_, err = os.Create("gofast-app.zip")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	file, err := os.OpenFile("gofast-app.zip", os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
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

func unzipFile() error {
	if os.Getenv("TEST") == "true" {
		return nil
	}
	archive, err := zip.OpenReader("gofast-app.zip")
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
		if !filepath.IsLocal(file.Name) {
			return fmt.Errorf("zip entry %q escapes the project directory", file.Name)
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

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(file.Name, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
			continue
		}

		dst, err := os.Create(file.Name)
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
	}
	return nil
}
