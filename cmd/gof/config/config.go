package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/charmbracelet/lipgloss"
)

const (
	SERVER_URL     = "https://admin.gofast.live"
	VERSION        = "v0.0.1"
	ConfigFileName = "gofast.json"
)

var (
	NoStyle      = lipgloss.NewStyle()
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("032"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ActiveStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	ErrStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	HelpStyle    = BlurredStyle
)

type Config struct {
	ProjectName string   `json:"project_name"`
	Models      []string `json:"models"`
}

func ParseConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("gofast.json config file not found. Please run 'gof init <project_name>' to create it")
		}
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func AddModel(modelName string) error {
	config, err := ParseConfig()
	if err != nil {
		return err
	}

	if slices.Contains(config.Models, modelName) {
		return fmt.Errorf("model '%s' already exists in the config", modelName)
	}

	config.Models = append(config.Models, modelName)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFileName, data, 0644)
}

