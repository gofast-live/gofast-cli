package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/lipgloss"
)

const (
	ServerURL      = "https://admin.gofast.live"
	Version        = "v2.17.0"
	ConfigFileName = "gofast.json"
)

func NoStyle() lipgloss.Style      { return lipgloss.NewStyle() }
func FocusedStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("032")) }
func BlurredStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("240")) }
func ActiveStyle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("244")) }
func ErrStyle() lipgloss.Style     { return lipgloss.NewStyle().Foreground(lipgloss.Color("9")) }
func SuccessStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) }
func HelpStyle() lipgloss.Style    { return BlurredStyle() }

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Model struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type Config struct {
	ProjectName         string    `json:"project_name"`
	Services            []Service `json:"services"`
	Models              []Model   `json:"models"`
	Integrations        []string  `json:"integrations"`
	InfraPopulated      bool      `json:"infra_populated"`
	MonitoringPopulated bool      `json:"monitoring_populated"`
}

type Service struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func writeConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	err = os.WriteFile(ConfigFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", ConfigFileName, err)
	}
	return nil
}

func ParseConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("gofast.json config file not found. Please run 'gof init <project_name> && cd <project_name>' to create a new project")
		}
		return nil, fmt.Errorf("reading %s: %w", ConfigFileName, err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ConfigFileName, err)
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

	return writeConfig(config)
}

func Initialize(projectDir, projectName string) error {
	cfg := Config{
		ProjectName:         projectName,
		InfraPopulated:      false,
		MonitoringPopulated: false,
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
		Integrations: []string{},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	configPath := filepath.Join(projectDir, ConfigFileName)
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("writing %s: %w", configPath, err)
	}
	return nil
}

func IsSvelte() bool {
	return HasService("svelte")
}

func IsTanstack() bool {
	return HasService("tanstack")
}

func HasService(name string) bool {
	config, err := ParseConfig()
	if err != nil {
		return false
	}
	for _, service := range config.Services {
		if service.Name == name {
			return true
		}
	}
	return false
}

func AddService(name, port string) error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}

	for _, service := range cfg.Services {
		if service.Name == name {
			return nil
		}
	}

	cfg.Services = append(cfg.Services, Service{Name: name, Port: port})
	return writeConfig(cfg)
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
	return writeConfig(cfg)
}

func MarkMonitoringPopulated() error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}

	if cfg.MonitoringPopulated {
		return nil
	}

	cfg.MonitoringPopulated = true
	return writeConfig(cfg)
}

func HasIntegration(name string) bool {
	cfg, err := ParseConfig()
	if err != nil {
		return false
	}
	return slices.Contains(cfg.Integrations, name)
}

func AddIntegration(name string) error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}

	// Check if already added
	if slices.Contains(cfg.Integrations, name) {
		return nil
	}

	cfg.Integrations = append(cfg.Integrations, name)
	return writeConfig(cfg)
}
