// aws/ssm.go
package aws

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

// NewSSMClient creates an SSM client using the provided AWS Config and region
func NewSSMClient(cfg aws.Config, region string) (*ssm.Client, error) {
    // Override the region in the provided AWS config
    cfg.Region = region

    // Create and return the SSM client using the provided config
    ssmClient := ssm.NewFromConfig(cfg)
    log.Printf("SSM client created for region %s", region)
    return ssmClient, nil
}

// ExecuteSSMCommand runs a command on an EC2 instance using SSM Run Command
func ExecuteSSMCommand(ssmClient *ssm.Client, instanceID string, command string, commandName string) (string, error) {
    input := &ssm.SendCommandInput{
        InstanceIds: []string{instanceID},
        DocumentName: aws.String("AWS-RunShellScript"),
        Parameters: map[string][]string{
            "commands": {command},
        },
        Comment: aws.String(commandName),
    }

    output, err := ssmClient.SendCommand(context.Background(), input)
    if err != nil {
        return "", fmt.Errorf("failed to execute command on instance %s: %w", instanceID, err)
    }

    log.Printf("Command execution initiated on instance %s, command ID: %s", 
        instanceID, *output.Command.CommandId)
    
    return *output.Command.CommandId, nil
}

// GetCommandStatus retrieves the status of a command execution
func GetCommandStatus(ssmClient *ssm.Client, commandID string, instanceID string) (string, string, error) {
    input := &ssm.GetCommandInvocationInput{
        CommandId: aws.String(commandID),
        InstanceId: aws.String(instanceID),
    }

    output, err := ssmClient.GetCommandInvocation(context.Background(), input)
    if err != nil {
        return "", "", fmt.Errorf("failed to retrieve command status: %w", err)
    }

    return string(output.Status), *output.StandardOutputContent, nil
}

// GetParameter retrieves a parameter value from AWS SSM Parameter Store
func GetParameter(ssmClient *ssm.Client, name string) (string, error) {
    input := &ssm.GetParameterInput{
        Name:           aws.String(name),
        WithDecryption: aws.Bool(true), // even if not SecureString, fine to request decryption
    }

    output, err := ssmClient.GetParameter(context.Background(), input)
    if err != nil {
        return "", fmt.Errorf("failed to get parameter %s: %w", name, err)
    }

    return *output.Parameter.Value, nil
}


// PutParameter saves or updates a parameter in AWS SSM Parameter Store
func PutParameter(ssmClient *ssm.Client, name string, value string) error {
    input := &ssm.PutParameterInput{
        Name:      aws.String(name),
        Value:     aws.String(value),
        Type:      "String", // You are using regular strings
        Overwrite: aws.Bool(true),     // Always overwrite existing values
    }

    _, err := ssmClient.PutParameter(context.Background(), input)
    if err != nil {
        return fmt.Errorf("failed to put parameter %s: %w", name, err)
    }

    return nil
}
