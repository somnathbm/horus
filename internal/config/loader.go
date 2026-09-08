package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Load app config
func Load(configPath string) (AppConfig, error) {
	// read the file first
	byteContent, err := os.ReadFile(configPath)
	if err != nil {
		return AppConfig{}, fmt.Errorf("error reading file: %w", err)
	}

	// parse yaml
	var appConfig AppConfig
	err = yaml.Load(byteContent, &appConfig, yaml.WithKnownFields(true))
	if err != nil {
		return AppConfig{}, fmt.Errorf("error parsing file: %w", err)
	}

	// validate
	normalizedConfig, err := validateConfig(appConfig)
	if err != nil {
		return AppConfig{}, fmt.Errorf("validation error: %w", err)
	}

	return normalizedConfig, nil
}
