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

func AssumeRoleInAccount(accountID string) (aws.Config, error) {
    // Define the role ARN based on the account ID and the role name
    roleArn := fmt.Sprintf("arn:aws:iam::%s:role/ec2-restart-manager", accountID)

    // Create an STS client using the global AWS configuration
    stsClient := sts.NewFromConfig(AWSConfig)

    // Create an AssumeRole provider with the specified role ARN
    assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)

    // Create a new AWS config with the assumed role's credentials provider
    assumedConfig, err := config.LoadDefaultConfig(context.Background(),
        config.WithCredentialsProvider(aws.NewCredentialsCache(assumeRoleProvider)),
    )
    if err != nil {
        return aws.Config{}, fmt.Errorf("failed to create AWS config with assumed role credentials: %w", err)
    }

    log.Printf("Assumed role %s in account %s", roleArn, accountID)
    return assumedConfig, nil
}


func GetCallerIdentity(cfg aws.Config) error {
    stsClient := sts.NewFromConfig(cfg)

    // Call GetCallerIdentity
    result, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
    if err != nil {
        return fmt.Errorf("failed to get caller identity: %w", err)
    }

    // Output the caller identity to the console and logs
    callerInfo := fmt.Sprintf("Assumed Role ARN: %s, Account: %s, User ID: %s",
        *result.Arn, *result.Account, *result.UserId)
    fmt.Println(callerInfo)   // Print to console
    log.Println(callerInfo)   // Log for records

    return nil
}