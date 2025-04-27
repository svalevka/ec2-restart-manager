package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type S3Config struct {
	Bucket string `yaml:"bucket"`
	Key    string `yaml:"key"`
}

type AzureADConfig struct {
	TenantID    string `yaml:"tenant_id"`
	ClientID    string `yaml:"client_id"`
	RedirectURL string `yaml:"redirect_url"`
	GroupID     string `yaml:"group_id"`
}

type EnvConfig struct {
	S3      S3Config     `yaml:"s3"`
	AzureAD AzureADConfig `yaml:"azure_ad"`
	Region  string        `yaml:"region"`
}

type Config struct {
	Env map[string]EnvConfig `yaml:"env"`
}

// LoadConfig loads the YAML config file and returns the config for the active environment
func LoadConfig() (*EnvConfig, error) {
	configFile, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return nil, err
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "dev" // default
	}

	envConfig, ok := config.Env[environment]
	if !ok {
		return nil, err
	}

	return &envConfig, nil
}
