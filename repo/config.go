package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	Email  string `json:"email"`
	ApiKey string `json:"api_key"`
}

func saveToConfig(email string, apiKey string) error {
	path, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("Could not get user config dir")
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Could not open config file")
	}
	defer jsonFile.Close()
	// marshal existing json
	// add new key value
	// write to file
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("Could not read config file")
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		// clean up file
		_ = jsonFile.Truncate(0)
		c = Config{}
	}
	c.Email = email
	c.ApiKey = apiKey
	data, err = json.Marshal(c)
	if err != nil {
		return fmt.Errorf("Could not marshal config data")
	}
	_, err = jsonFile.WriteAt(data, 0)
	if err != nil {
		return fmt.Errorf("Could not write to config file")
	}
	return nil
}

func checkConfig(email string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if email == "" {
			return errMsg(fmt.Errorf("Email is required"))
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return errMsg(fmt.Errorf("Invalid email address"))
		}
		if apiKey == "" {
			return errMsg(fmt.Errorf("API key is required"))
		}
		err := saveToConfig(email, apiKey)
		if err != nil {
			return errMsg(err)
		}
		return configValid("")
	}
}

type Response struct {
	GithubToken string `json:"github_token"`
}

func validateConfig() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Could not get user config dir")
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not open config file")
	}
	defer jsonFile.Close()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return "", fmt.Errorf("Could not read config file")
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return "", fmt.Errorf("Could not unmarshal config data")
	}

	// make http call with query email and header authorization with api key
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:4003/repo", nil)
	if err != nil {
		return "", fmt.Errorf("Could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiKey)
	q := req.URL.Query()
	q.Add("email", c.Email)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Could not make request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Invalid credentials")
	}
	var r Response
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", fmt.Errorf("Could not decode response")
	}
	return r.GithubToken, nil
}
