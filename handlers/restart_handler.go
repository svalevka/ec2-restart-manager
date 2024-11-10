package handlers

import (
    "fmt"
    "log"
    "net/http"
    "sync"
    "ec2-restart-manager/aws"
    "ec2-restart-manager/models"
)

// Mutex to handle concurrent access to the status map
var statusLock sync.Mutex

// Map to store the status of each instance restart operation
var statusMap = make(map[string]string)

// RestartHandler handles the request to restart EC2 instances in multiple accounts/regions
func RestartHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        log.Printf("Error parsing form data: %v", err)
        return
    }

    instanceIDs := r.Form["instance_ids"]
    if len(instanceIDs) == 0 {
        http.Error(w, "No instance IDs provided", http.StatusBadRequest)
        return
    }

    for _, instanceID := range instanceIDs {
        // Retrieve instance details such as account number and region
        instance, err := models.GetInstanceDetails(instanceID)
        if err != nil {
            log.Printf("Error fetching instance details for %s: %v", instanceID, err)
            continue
        }

        // Assume the role in the target AWS account and get the AWS Config
        assumedConfig, err := aws.AssumeRoleInAccount(instance.AWSAccountNumber)
        if err != nil {
            log.Printf("Error assuming role in account %s for instance %s: %v", instance.AWSAccountNumber, instanceID, err)
            continue
        }

        // Confirm the assumed role identity
        fmt.Println("Confirming assumed role identity:")
        if err := aws.GetCallerIdentity(assumedConfig); err != nil {
            log.Printf("Failed to confirm assumed role identity: %v", err)
            continue
        }

        // Create an EC2 client using the assumed role config and target region
        ec2Client, err := aws.NewEC2Client(assumedConfig, instance.Region)
        if err != nil {
            log.Printf("Error creating EC2 client in region %s for instance %s: %v", instance.Region, instanceID, err)
            continue
        }

        // Optional: List instances available to the assumed role to confirm visibility
        err = aws.ListInstances(ec2Client)
        if err != nil {
            log.Printf("Failed to list instances in region %s: %v", instance.Region, err)
        }

        // Attempt to restart the specific instance
        err = aws.RestartEC2Instance(ec2Client, instanceID)
        if err != nil {
            log.Printf("Failed to restart instance %s: %v", instanceID, err)
        } else {
            log.Printf("Successfully restarted instance %s in region %s", instanceID, instance.Region)
        }
    }

    // Redirect to /status page after the restart process
    http.Redirect(w, r, "/status", http.StatusSeeOther)
}
