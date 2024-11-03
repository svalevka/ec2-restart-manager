// config/config.go
package config

import (
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    S3 struct {
        Bucket string `yaml:"bucket"`
        Key    string `yaml:"key"`
    } `yaml:"s3"`
    AzureAD struct {
        TenantID    string `yaml:"tenant_id"`
        ClientID    string `yaml:"client_id"`
        RedirectURL string `yaml:"redirect_url"`
        GroupID     string `yaml:"group_id"`
    } `yaml:"azure_ad"`
}

func LoadConfig() (*Config, error) {
    data, err := os.ReadFile("config/config.yaml")
    if err != nil {
        return nil, err
    }
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    if err != nil {
        return nil, err
    }
    return &cfg, nil
}
