package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
)

type Config struct {
	Email  string `json:"email"`
	APIKey string `json:"api_key"`
}

func checkConfig(email string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if email == "" {
			return errMsg{nil, "Email address is required"}
		}
		_, err := mail.ParseAddress(email)
		if err != nil {
			return errMsg{err, "Invalid email address format"}
		}
		if apiKey == "" {
			return errMsg{nil, "API key is required"}
		}
		// Validate first, so a typo doesn't replace credentials that still work
		err = validateConfig(email, apiKey)
		if err != nil {
			return errMsg{err, "Authentication failed, please check your email and API key"}
		}
		err = saveToConfig(email, apiKey)
		if err != nil {
			return errMsg{err, "Error saving configuration"}
		}
		return authMsg{email, apiKey}
	}
}

func CheckAuthentication() (string, string, error) {
	if os.Getenv("TEST") == "true" {
		return "test@gofast.local", "test-api-key", nil
	}

	path, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("getting user config dir: %w", err)
	}
	configPath := path + "/gofast.json"
	jsonFile, err := os.OpenFile(configPath, os.O_RDWR, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", errors.New("config file not found. Please run 'gof auth'")
		}
		return "", "", fmt.Errorf("opening %s: %w", configPath, err)
	}
	defer func() {
		closeErr := jsonFile.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing config file: %v\n", closeErr)
		}
	}()

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return "", "", fmt.Errorf("reading auth config: %w", err)
	}

	if len(data) == 0 {
		return "", "", errors.New("config file is empty. Please run 'gof auth'")
	}

	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return "", "", errors.New("failed to parse config file. It might be corrupted. Please run 'gof auth'")
	}

	if c.Email == "" || c.APIKey == "" {
		return "", "", errors.New("email or API key not found in config. Please run 'gof auth'")
	}

	err = validateConfig(c.Email, c.APIKey)
	if err != nil {
		return "", "", fmt.Errorf("authentication failed: %w. Please run 'gof auth'", err)
	}

	return c.Email, c.APIKey, nil
}

func saveToConfig(email string, apiKey string) error {
	path, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error getting user config directory: %w", err)
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		closeErr := jsonFile.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing config file: %v\n", closeErr)
		}
	}()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		_ = jsonFile.Truncate(0)
		c = Config{Email: "", APIKey: ""}
	}
	c.Email = email
	c.APIKey = apiKey
	//nolint:gosec // G117: the CLI stores the user's own API key in their local config on purpose
	data, err = json.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	_ = jsonFile.Truncate(0)
	_, err = jsonFile.WriteAt(data, 0)
	if err != nil {
		return fmt.Errorf("error writing to config file: %w", err)
	}
	return nil
}

func validateConfig(email string, apiKey string) error {
	// the auth TUI has no caller context, so the request carries its own deadline
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.ServerURL+"/repo", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	q := req.URL.Query()
	q.Add("email", email)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing response body: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return errors.New("error validating configuration: " + resp.Status)
	}
	return nil
}
