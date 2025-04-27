package aws

import (
    "context"
    "log"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    "github.com/aws/aws-sdk-go-v2/credentials/stscreds" 
)

// Global AWS configuration
var AWSConfig aws.Config  // ✅ Your global config object

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


// AssumeRoleInAccount assumes a role in another AWS account and returns a new AWS config
func AssumeRoleInAccount(roleName string, accountID string) (aws.Config, error) {
    roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)

    stsClient := sts.NewFromConfig(AWSConfig)

    assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)

    assumedConfig, err := config.LoadDefaultConfig(context.Background(),
        config.WithCredentialsProvider(aws.NewCredentialsCache(assumeRoleProvider)),
    )
    if err != nil {
        return aws.Config{}, fmt.Errorf("failed to assume role %s in account %s: %w", roleArn, accountID, err)
    }

    log.Printf("Assumed role %s in account %s", roleArn, accountID)
    return assumedConfig, nil
}

// GetCallerIdentity prints the current caller identity
func GetCallerIdentity(cfg aws.Config) error {
    stsClient := sts.NewFromConfig(cfg)

    result, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
    if err != nil {
        return fmt.Errorf("failed to get caller identity: %w", err)
    }

    callerInfo := fmt.Sprintf("Assumed Role ARN: %s, Account: %s, User ID: %s",
        *result.Arn, *result.Account, *result.UserId)
    fmt.Println(callerInfo)
    log.Println(callerInfo)

    return nil
}
