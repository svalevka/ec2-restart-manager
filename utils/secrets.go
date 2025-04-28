// utils/secrets.go
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"ec2-restart-manager/aws" 
)

// LoadSecretToEnv fetches a specific key from AWS Secrets Manager and sets it as an environment variable.
func LoadSecretToEnv(environment, secretName, secretKey string) (string, error) {

	// Ensure AWS configuration is initialized
	if aws.AWSConfig.Credentials == nil {
		aws.InitAWSConfig()
	}

	// Initialize Secrets Manager client using centralized AWS configuration
	svc := secretsmanager.NewFromConfig(aws.AWSConfig)

	// Retrieve the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}

	// Parse the secret JSON string to get the specific key's value
	var secretData map[string]string
	if result.SecretString != nil {
		if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
			return "", fmt.Errorf("failed to parse secret JSON: %w", err)
		}
	} else {
		return "" ,fmt.Errorf("secret string is nil for secret: %s", secretName)
	}

	// Retrieve the specific key value and set it as the environment variable
	secretValue, ok := secretData[secretKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret", secretKey)
	}

	return secretValue, nil

}