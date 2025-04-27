package models

import (
	"encoding/json"
	"fmt"
	"sync"

	"ec2-restart-manager/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ScheduleConfig struct {
	StgDevDay  string `json:"stg_dev_day"`
	StgDevTime string `json:"stg_dev_time"`
	ProdDay    string `json:"prod_day"`
	ProdTime   string `json:"prod_time"`
}

var (
	scheduleConfig     ScheduleConfig
	scheduleConfigLock sync.RWMutex
	ssmClient          *ssm.Client
	envName            string
)

// InjectSSMClient injects the SSM client
func InjectSSMClient(client *ssm.Client) {
	ssmClient = client
}

// InjectEnvName injects the environment name (e.g., "prod", "dev", "test")
func InjectEnvName(env string) {
	envName = env
}

// LoadScheduleConfig loads the schedule configuration from SSM Parameter Store
func LoadScheduleConfig() error {
	scheduleConfigLock.Lock()
	defer scheduleConfigLock.Unlock()

	paramName := fmt.Sprintf("/ec2-restart-manager/%s/schedule", envName)

	paramValue, err := aws.GetParameter(ssmClient, paramName)
	if err != nil {
		return fmt.Errorf("failed to load schedule config from SSM: %w", err)
	}

	var loadedConfig ScheduleConfig
	err = json.Unmarshal([]byte(paramValue), &loadedConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal schedule config: %w", err)
	}

	scheduleConfig = loadedConfig
	return nil
}

// GetScheduleConfig returns the current in-memory schedule config
func GetScheduleConfig() ScheduleConfig {
	scheduleConfigLock.RLock()
	defer scheduleConfigLock.RUnlock()
	return scheduleConfig
}

// SaveScheduleConfig saves the schedule configuration to SSM Parameter Store
func SaveScheduleConfig(newConfig ScheduleConfig) error {
	scheduleConfigLock.Lock()
	defer scheduleConfigLock.Unlock()

	paramName := fmt.Sprintf("/ec2-restart-manager/%s/schedule", envName)

	jsonData, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schedule config: %w", err)
	}

	if err := aws.PutParameter(ssmClient, paramName, string(jsonData)); err != nil {
		return fmt.Errorf("failed to save schedule config to SSM: %w", err)
	}

	// Update in-memory config after successful save
	scheduleConfig = newConfig

	return nil
}
