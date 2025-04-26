// models/ec2_instance.go
package models

import (
	"fmt"
)


type EC2Instance struct {
	AWSAccountName   string `csv:"AWS Account Name"`
	AWSAccountNumber string `csv:"AWS Account ID"`
	State            string `csv:"State"`
	EC2Name          string `csv:"EC2 Name"`
	Service          string `csv:"Service"`
	Owner            string `csv:"Owner"`
	ID               string `csv:"ID"`
	Region           string `csv:"Region"`
	EnvironmentClass string `csv:"EnvironmentClass"`
	RestartTimestamp string 
    CommandOutput    string  // Output of the most recent command execution
    CommandTimestamp string  // When the command was executed
    Command          string  // The command that was executed
	// Add other fields as needed
}

type TemplateData struct {
	Title				   string
	Version 			   string
	Instances              []EC2Instance
	UniqueOwners           []string
	SelectedOwner          string
	UniqueServices         []string
	SelectedService        string
	UniqueAWSAccountNames  []string
	SelectedAWSAccountName string
	UniqueRegions          []string
	SelectedRegion         string
	IsLoggedIn			   bool
	UserName 			   string	
	AzureAuthenticated	   bool
	StatusMap              map[string]string 
}

// Global cache to store EC2 instances by their ID
var instanceCache = make(map[string]EC2Instance)

// LoadInstances populates the instance cache from a slice of EC2Instance structs
func LoadInstances(instances []EC2Instance) {
	for _, instance := range instances {
		instanceCache[instance.ID] = instance
	}
}

// GetInstanceDetails retrieves the details of an EC2 instance by its ID
func GetInstanceDetails(instanceID string) (*EC2Instance, error) {
	instance, exists := instanceCache[instanceID]
	if !exists {
		return nil, fmt.Errorf("instance ID %s not found", instanceID)
	}
	return &instance, nil
}

// GetInstances retrieves all EC2 instances from the global instance cache
func GetInstances() []EC2Instance {
	instances := make([]EC2Instance, 0, len(instanceCache))
	for _, instance := range instanceCache {
		instances = append(instances, instance)
	}
	return instances
}


