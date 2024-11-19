package aws

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
)

// NewEC2Client creates an EC2 client using the provided AWS Config and region
func NewEC2Client(cfg aws.Config, region string) (*ec2.Client, error) {
    // Override the region in the provided AWS Config
    cfg.Region = region

    // Create and return the EC2 client using the provided config
    ec2Client := ec2.NewFromConfig(cfg)
    log.Printf("EC2 client created for region %s", region)
    return ec2Client, nil
}

// RestartEC2Instance restarts an EC2 instance with the provided EC2 client
func RestartEC2Instance(ec2Client *ec2.Client, instanceID string) error {
    // Define the input for the RebootInstances API call
    input := &ec2.RebootInstancesInput{
        InstanceIds: []string{instanceID},
    }

    // Attempt to reboot the instance
    _, err := ec2Client.RebootInstances(context.Background(), input)
    if err != nil {
        return fmt.Errorf("failed to restart instance %s: %w", instanceID, err)
    }

    log.Printf("Instance %s successfully restarted", instanceID)
    return nil
}

// ListInstances retrieves and outputs the list of instance IDs available to the assumed role in the specified region
func ListInstances(ec2Client *ec2.Client) error {
    input := &ec2.DescribeInstancesInput{}
    instances := []string{}

    // Paginate through all instances
    paginator := ec2.NewDescribeInstancesPaginator(ec2Client, input)
    for paginator.HasMorePages() {
        page, err := paginator.NextPage(context.Background())
        if err != nil {
            return fmt.Errorf("failed to describe instances: %w", err)
        }

        for _, reservation := range page.Reservations {
            for _, instance := range reservation.Instances {
                instances = append(instances, *instance.InstanceId)
            }
        }
    }

    // Output the list of instances to both the console and logs
    instanceList := fmt.Sprintf("Instances available to the assumed role: %v", instances)
    fmt.Println(instanceList)   // Print to console
    log.Println(instanceList)   // Log for records

    return nil
}
