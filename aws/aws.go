package aws

import (
    "context"
    "log"
	"fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
)

// Global AWS configuration
var AWSConfig aws.Config

// Initialize AWS configuration
func InitAWSConfig() error {
    var err error
    AWSConfig, err = config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-west-2"))
    if err != nil {
        return fmt.Errorf("unable to load AWS configuration: %w", err)
    }
    log.Printf("AWS configuration loaded with region: %s", AWSConfig.Region)
    return nil
}
