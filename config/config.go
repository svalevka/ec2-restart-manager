// Modified code for config/config.go
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
	S3       S3Config     `yaml:"s3"`
	AzureAD  AzureADConfig `yaml:"azure_ad"`
	Region   string        `yaml:"region"`
	// Adding Environment field to store the environment name
	Environment string        // This is not from yaml, will be set programmatically
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
