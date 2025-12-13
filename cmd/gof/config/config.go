package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

const (
	SERVER_URL     = "https://admin.gofast.live"
	VERSION        = "v2.10.1"
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

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Model struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type Config struct {
	ProjectName     string    `json:"project_name"`
	Services        []Service `json:"services"`
	Models          []Model   `json:"models"`
	InfraPopulated  bool      `json:"infra_populated"`
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

func AddModel(modelName string, columns []Column) error {
	config, err := ParseConfig()
	if err != nil {
		return err
	}

	for _, m := range config.Models {
		if m.Name == modelName {
			return fmt.Errorf("model '%s' already exists in the config", modelName)
		}
	}

	newModel := Model{
		Name:    modelName,
		Columns: columns,
	}
	config.Models = append(config.Models, newModel)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFileName, data, 0644)
}

func Initialize(projectName string) error {
	cfg := Config{
		ProjectName:    projectName,
		InfraPopulated: false,
		Services: []Service{
			{Name: "core", Port: "4000"},
		},
		Models: []Model{
			{
				Name: "skeleton",
				Columns: []Column{
					{Name: "name", Type: "string"},
					{Name: "age", Type: "number"},
					{Name: "death", Type: "time"},
					{Name: "zombie", Type: "bool"},
				},
			},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(projectName+"/"+ConfigFileName, data, 0644)
}

func IsSvelte() bool {
	config, err := ParseConfig()
	if err != nil {
		return false
	}
	for _, service := range config.Services {
		if service.Name == "svelte" {
			return true
		}
	}
	return false
}

func MarkInfraPopulated() error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}

	if cfg.InfraPopulated {
		return nil
	}

	cfg.InfraPopulated = true

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFileName, data, 0644)
}
