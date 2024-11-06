package utils

import (
	//"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// LoadSecretToEnv fetches a specific key from AWS Secrets Manager and sets it as an environment variable.
// It only loads the secret if the environment is "test".
func LoadSecretToEnv(environment, secretName, secretKey, envVarName string) error {
	if environment != "test" {
		return nil // Only load the secret if the environment is "test"
	}

	// Initialize a session using default credentials and config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize AWS session: %w", err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.New(sess)

	// Retrieve the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		return fmt.Errorf("failed to retrieve secret: %w", err)
	}

	// Parse the secret JSON string to get the specific key's value
	var secretData map[string]string
	if result.SecretString != nil {
		if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
			return fmt.Errorf("failed to parse secret JSON: %w", err)
		}
	} else {
		return fmt.Errorf("secret string is nil for secret: %s", secretName)
	}

	// Retrieve the specific key value and set it as the environment variable
	secretValue, exists := secretData[secretKey]
	if !exists {
		return fmt.Errorf("key %s not found in secret %s", secretKey, secretName)
	}
	if err := os.Setenv(envVarName, secretValue); err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}
	log.Printf("Successfully set environment variable %s", envVarName)

	return nil
}
