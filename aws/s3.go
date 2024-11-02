// aws/s3.go
package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client *s3.Client
)

// SetupAWSClients initializes the AWS SDK configuration and sets up clients.
func SetupAWSClients() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Initialize the S3 client
	S3Client = s3.NewFromConfig(cfg)
}
