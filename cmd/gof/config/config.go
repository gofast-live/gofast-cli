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
	VERSION        = "v2.4.0"
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
	ProjectName string    `json:"project_name"`
	Models      []string  `json:"models"`
	Services    []Service `json:"services"`
}

type Service struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func ParseConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("gofast.json config file not found. Please run 'gof init <project_name> && cd <project_name>' to create a new project")
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

// Initialize creates an initial gofast.json in the provided project directory
// using the Config struct as the canonical schema.
func Initialize(projectName string) error {
	cfg := Config{
		ProjectName: projectName,
		Models:      []string{"skeleton"},
		Services: []Service{
			{Name: "core", Port: "4000"},
			{Name: "client", Port: "3000"},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(projectName+"/"+ConfigFileName, data, 0644)
}
