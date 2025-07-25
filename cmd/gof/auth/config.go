package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gofast-live/gofast-cli/cmd/gof/config"
)

type Config struct {
	Email  string `json:"email"`
	ApiKey string `json:"api_key"`
}

func checkConfig(email string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if email == "" {
			return errMsg{nil, "Email address is required"}
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return errMsg{err, "Invalid email address format"}
		}
		if apiKey == "" {
			return errMsg{nil, "API key is required"}
		}
		err := saveToConfig(email, apiKey)
		if err != nil {
			return errMsg{err, "Error saving configuration"}
		}
		err = validateConfig(email, apiKey)
		if err != nil {
			return errMsg{err, "Authentication failed, please check your email and API key"}
		}
		return authMsg{email, apiKey}
	}
}

func readConfig() (email string, apiKey string, err error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return "", "", err
	}
	defer func() {
		closeErr := jsonFile.Close()
		if closeErr != nil {
			fmt.Printf("error closing response body: %v\n", closeErr)
		}
	}()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return "", "", err
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		c.Email = ""
		c.ApiKey = ""
		err := saveToConfig(c.Email, c.ApiKey)
		if err != nil {
			return "", "", err
		}
	}
	return c.Email, c.ApiKey, nil
}

func saveToConfig(email string, apiKey string) error {
	path, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error getting user config directory: %w", err)
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		closeErr := jsonFile.Close()
		if closeErr != nil {
			fmt.Printf("error closing response body: %v\n", closeErr)
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
		c = Config{}
	}
	c.Email = email
	c.ApiKey = apiKey
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.SERVER_URL+"/repo", nil)
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
			fmt.Printf("error closing response body: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != 200 {
		return errors.New("error validating configuration: " + resp.Status)
	}
	return nil
}
