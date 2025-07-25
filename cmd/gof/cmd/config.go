package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg    error
	authMsg   struct{ email, apiKey string }
	copyMsg   struct{ err error }
	finishMsg struct{ docker []string }
)

var (
	SERVER_URL   = "https://admin.gofast.live"
	VERSION      = "v1.2.1"
	noStyle      = lipgloss.NewStyle()
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	helpStyle    = blurredStyle
)

type Config struct {
	Email  string `json:"email"`
	ApiKey string `json:"api_key"`
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
		err = validateConfig(email, apiKey)
		if err != nil {
			return errMsg(err)
		}
		return authMsg{email, apiKey}
	}
}

func readConfig() (email string, apiKey string, err error) {
	path, err := os.UserConfigDir()
	if err != nil {
        panic(err)
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
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
		return fmt.Errorf("Could not get user config dir")
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Could not open config file")
	}
	defer jsonFile.Close()
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
	_ = jsonFile.Truncate(0)
	_, err = jsonFile.WriteAt(data, 0)
	if err != nil {
		return fmt.Errorf("Could not write to config file")
	}
	return nil
}

func validateConfig(email string, apiKey string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", SERVER_URL+"/repo", nil)
	if err != nil {
		return fmt.Errorf("Could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	q := req.URL.Query()
	q.Add("email", email)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not make request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid credentials")
	}
	return nil
}
