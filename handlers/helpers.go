package handlers

import (
	"log"

	"ec2-restart-manager/aws"
	"ec2-restart-manager/models"
	"ec2-restart-manager/utils"
)

// updateInstancesFromS3 fetches and loads the latest instances from S3
func updateInstancesFromS3(bucket, key string) error {
	// Fetch CSV from S3
	csvContent, err := aws.GetCSVFromS3(bucket, key)
	if err != nil {
		return err
	}

	// Parse CSV into instances
	instances, err := utils.ParseCSVToStruct(csvContent)
	if err != nil {
		return err
	}

	// Load parsed instances into global cache
	models.LoadInstances(instances)
    if utils.Debug {
		log.Printf("Instances successfully loaded from S3: %d", len(instances))
	}
	return nil
}
