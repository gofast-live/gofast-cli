package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gofast-live/gofast-cli/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/cmd/gof/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [project_name]",
	Short: "Initialize the Go service",
	Long:  "Initialize the Go service with Docker and PostgreSQL setup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}
		projectName := args[0]
		if projectName == "" {
			cmd.Println("Project name cannot be empty.")
			return
		}
		err = downloadRepo(email, apiKey, projectName)
		if err != nil {
			cmd.Printf("Error downloading repository: %v\n", err)
			return
		}
		cmd.Printf("Project '%s' initialized successfully.\n", projectName)
	},
}

func downloadRepo(email string, apiKey string, projectName string) error {
	// If test env, copy ../gofast-app to the current directory
	if os.Getenv("TEST") == "true" {
		cmd := exec.Command("cp", "-r", "../gofast-app", "./"+projectName)
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
	// get the file
	err := getFile(email, apiKey)
	if err != nil {
		return err
	}
	// unzip the file
	err = unzipFile()
	if err != nil {
		return err
	}
	// remove the zip file
	err = os.Remove("gofast-app.zip")
	if err != nil {
		return err
	}
	// find and rename the folder `gofast-live-gofast-app-...` to the project name
	files, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			if strings.HasPrefix(f.Name(), "gofast-live-gofast-app-") {
				err := os.Rename(f.Name(), projectName)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

func getFile(email string, apiKey string) error {
	client := http.Client{}
	req, err := http.NewRequest("GET", config.SERVER_URL+"/v2?email="+email, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading file: %s", resp.Status)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	// save the file to the disk
	_, err = os.Create("gofast-app.zip")
	if err != nil {
		return err
	}
	file, err := os.OpenFile("gofast-app.zip", os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func unzipFile() error {
	if os.Getenv("TEST") == "true" {
		return nil
	}
	archive, err := zip.OpenReader("gofast-app.zip")
	if err != nil {
		return err
	}
	defer func() {
		err := archive.Close()
		if err != nil {
			fmt.Printf("error closing archive: %v\n", err)
		}
	}()
	for _, file := range archive.File {
		src, err := file.Open()
		if err != nil {
			return err
		}

		defer func() {
			err := src.Close()
			if err != nil {
				fmt.Printf("error closing source file: %v\n", err)
			}
		}()

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(file.Name, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		dst, err := os.Create(file.Name)
		if err != nil {
			return err
		}
		defer func() {
			err := dst.Close()
			if err != nil {
				fmt.Printf("error closing destination file: %v\n", err)
			}
		}()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
	return nil
}
