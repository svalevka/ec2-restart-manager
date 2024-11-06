package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

)

type AzureADConfig struct {
	TenantID    string `yaml:"tenant_id"`
	ClientID    string `yaml:"client_id"`
	RedirectURL string `yaml:"redirect_url"`
	GroupID     string `yaml:"group_id"`
}

type S3Config struct {
	Bucket string `yaml:"bucket"`
	Key    string `yaml:"key"`
}

type EnvConfig struct {
	Environment string        // Field to store environment name
	S3          S3Config      `yaml:"s3"`
	AzureAD     AzureADConfig `yaml:"azure_ad"`
	Region      string        `yaml:"region"` // New field for region
}

type Config struct {
	Env map[string]EnvConfig `yaml:"env"`
}

func LoadConfig() (*EnvConfig, error) {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// Determine the environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "test" // Default to 'test' if ENVIRONMENT is not set
	}

	// Retrieve the configuration for the specified environment
	envConfig, exists := cfg.Env[env]
	if !exists {
		return nil, fmt.Errorf("environment %q configuration not found", env)
	}

	envConfig.Environment = env // Set the environment name in the config

	return &envConfig, nil
}
