// utils/unique_values.go
package utils

import (
    "ec2-restart-manager/models"
    "sort"
    "strings"
)

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func GetUniqueOwners(instances []models.EC2Instance) []string {
    var owners []string
    for _, instance := range instances {
        owner := strings.TrimSpace(instance.Owner)
        if owner != "" {
            if !contains(owners, owner) {
                owners = append(owners, owner)
            }
        }
    }
    sort.Strings(owners)
    return owners
}

func GetUniqueServices(instances []models.EC2Instance) []string {
    var services []string
    for _, instance := range instances {
        service := strings.TrimSpace(instance.Service)
        if service != "" {
            if !contains(services, service) {
                services = append(services, service)
            }
        }
    }
    sort.Strings(services)
    return services
}

func GetUniqueAWSAccountNames(instances []models.EC2Instance) []string {
    var accounts []string
    for _, instance := range instances {
        account := strings.TrimSpace(instance.AWSAccountName)
        if account != "" {
            if !contains(accounts, account) {
                accounts = append(accounts, account)
            }
        }
    }
    sort.Strings(accounts)
    return accounts
}

func GetUniqueRegions(instances []models.EC2Instance) []string {
    var regions []string
    for _, instance := range instances {
        region := strings.TrimSpace(instance.Region)
        if region != "" {
            if !contains(regions, region) {
                regions = append(regions, region)
            }
        }
    }
    sort.Strings(regions)
    return regions
}

func FilterInstances(instances []models.EC2Instance, owner, service, awsAccountName, region string) []models.EC2Instance {
    var filtered []models.EC2Instance
    for _, instance := range instances {
        if (owner == "" || instance.Owner == owner) &&
            (service == "" || instance.Service == service) &&
            (awsAccountName == "" || instance.AWSAccountName == awsAccountName) &&
            (region == "" || instance.Region == region) {
            filtered = append(filtered, instance)
        }
    }
    return filtered
}