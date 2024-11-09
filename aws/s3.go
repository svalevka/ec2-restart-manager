// aws/s3.go
package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client to be used for S3 operations
var S3Client *s3.Client

// SetupS3Client initializes the S3 client using the global AWS configuration
func SetupS3Client() {
	if AWSConfig.Credentials == nil {
		InitAWSConfig() // Ensure AWSConfig is initialized
	}
	S3Client = s3.NewFromConfig(AWSConfig)

}

// GetCSVFromS3 retrieves a CSV file from an S3 bucket
func GetCSVFromS3(bucket, key string) ([]byte, error) {
	output, err := S3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3 bucket '%s' with key '%s': %w", bucket, key, err)
	}
	defer output.Body.Close()

	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body from S3 bucket '%s' with key '%s': %w", bucket, key, err)
	}
	return content, nil
}
