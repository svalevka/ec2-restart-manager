// aws/s3.go
package aws

import (
	"context"
	"io"
	"log"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client *s3.Client
)


func SetupAWSClients() {
    cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-west-2"))
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

	log.Printf("Region being used: %s", cfg.Region)

    // Initialize the S3 client
    S3Client = s3.NewFromConfig(cfg)
}

func GetCSVFromS3(bucket, key string) ([]byte, error) {

	output, err := S3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Wrap the error with bucket and key information
		return nil, fmt.Errorf("failed to get object from S3 bucket '%s' with key '%s': %w", bucket, key, err)
	}
	defer output.Body.Close()

	content, err := io.ReadAll(output.Body)
	if err != nil {
		// Wrap the error with bucket and key information
		return nil, fmt.Errorf("failed to read object body from S3 bucket '%s' with key '%s': %w", bucket, key, err)
	}
	return content, nil
}
