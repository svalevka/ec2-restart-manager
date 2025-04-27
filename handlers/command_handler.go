// handlers/command_handler.go
package handlers

import (
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    "html/template"

    "ec2-restart-manager/aws"
    "ec2-restart-manager/models"
    "ec2-restart-manager/auth"
    "ec2-restart-manager/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

var command_role_name = "ec2-restart-manager"

// Mutex to handle concurrent access to the command status map
var commandStatusLock sync.Mutex

// Struct to store the status and output of each command execution
type CommandStatus struct {
    Status    string // e.g., "Success", "Failed", "InProgress"
    Output    string // Command output
    Timestamp string // ISO 8601 format timestamp
    CommandID string // AWS SSM Command ID
    Command   string // The command that was executed
}

// Map to store the status of each command execution operation
var commandStatusMap = make(map[string]CommandStatus)

// CommandHandler handles the request to execute commands on EC2 instances
func CommandHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        log.Printf("Error parsing form data: %v", err)
        return
    }

    instanceIDs := r.Form["instance_ids"]
    commandType := r.FormValue("command_type")
    customCommand := r.FormValue("custom_command")
    
    if len(instanceIDs) == 0 {
        http.Error(w, "No instance IDs provided", http.StatusBadRequest)
        return
    }

    // Get the schedule configuration
    scheduleConfig := models.GetScheduleConfig()

    for _, instanceID := range instanceIDs {
        // Retrieve instance details such as account number and region
        instance, err := models.GetInstanceDetails(instanceID)
        if err != nil {
            log.Printf("Error fetching instance details for %s: %v", instanceID, err)
            updateCommandStatus(instanceID, "Failed to fetch instance details", "", "", "")
            continue
        }

        // Determine which command to execute based on command type and environment class
        var command, commandName string
        
        if commandType == "patching" {
            baseCommand := "sudo yum update-minimal --security -y || sudo dnf update --security --bugfix --enhancement=important --enhancement=moderate --enhancement=low -y"
            envClass := instance.EnvironmentClass
            
            if envClass == "stg" || envClass == "dev" {
                // For staging/dev: Use configured day/time with reboot
                command = fmt.Sprintf("echo \"sleep $((RANDOM %% 1800)); %s && sudo reboot\" | at -t $(date -d 'next %s %s GMT' +%%Y%%m%%d%%H%%M)", 
                    baseCommand, 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
                commandName = fmt.Sprintf("Scheduled Security Patching (%s %s GMT with reboot)", 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
            } else if envClass == "prod" {
                // For prod: Use configured day/time without reboot
                command = fmt.Sprintf("echo \"sleep $((RANDOM %% 1800)); %s\" | at -t $(date -d 'next %s %s GMT' +%%Y%%m%%d%%H%%M)", 
                    baseCommand, 
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime)
                commandName = fmt.Sprintf("Scheduled Security Patching (%s %s GMT)", 
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime)
            } else {
                // For other environments, run immediately
                command = baseCommand
                commandName = "Security Patching"
            }
        } else if commandType == "upgrade" {
            baseCommand := "sudo yum update -y || sudo dnf update -y"
            envClass := instance.EnvironmentClass
            
            if envClass == "stg" || envClass == "dev" {
                // For staging/dev: Use configured day/time with reboot
                command = fmt.Sprintf("echo \"sleep $((RANDOM %% 1800)); %s && sudo reboot\" | at -t $(date -d 'next %s %s GMT' +%%Y%%m%%d%%H%%M)", 
                    baseCommand, 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
                commandName = fmt.Sprintf("Scheduled System Upgrade (%s %s GMT with reboot)", 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
            } else if envClass == "prod" {
                // For prod: Use configured day/time without reboot
                command = fmt.Sprintf("echo \"sleep $((RANDOM %% 1800)); %s\" | at -t $(date -d 'next %s %s GMT' +%%Y%%m%%d%%H%%M)", 
                    baseCommand, 
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime)
                commandName = fmt.Sprintf("Scheduled System Upgrade (%s %s GMT)", 
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime)
            } else {
                // For other environments, run immediately
                command = baseCommand
                commandName = "System Upgrade"
            }
        } else if commandType == "custom" && customCommand != "" {
            command = customCommand
            commandName = "Custom Command"
        } else {
            updateCommandStatus(instanceID, "Invalid command type", "", "", "")
            continue
        }

        // Assume the role in the target AWS account and get the AWS Config
        assumedConfig, err := aws.AssumeRoleInAccount(command_role_name, instance.AWSAccountNumber)
        if err != nil {
            log.Printf("Error assuming role in account %s for instance %s: %v", instance.AWSAccountNumber, instanceID, err)
            updateCommandStatus(instanceID, "Failed to assume role in account", "", "", command)
            continue
        }

        // Create an SSM client using the assumed role config and target region
        ssmClient, err := aws.NewSSMClient(assumedConfig, instance.Region)
        if err != nil {
            log.Printf("Error creating SSM client in region %s for instance %s: %v", instance.Region, instanceID, err)
            updateCommandStatus(instanceID, "Failed to create SSM client", "", "", command)
            continue
        }

        // Execute the command on the instance
        commandID, err := aws.ExecuteSSMCommand(ssmClient, instanceID, command, commandName)
        if err != nil {
            log.Printf("Failed to execute command on instance %s: %v", instanceID, err)
            updateCommandStatus(instanceID, "Failed to execute command", "", "", command)
        } else {
            log.Printf("Command execution initiated on instance %s", instanceID)
            updateCommandStatus(instanceID, "InProgress", "", commandID, command)
            
            // Start a goroutine to check the command status periodically
            go checkCommandStatus(ssmClient, commandID, instanceID, instance.Region, command)
        }
    }

    // Redirect to the command status page
    http.Redirect(w, r, "/command-status", http.StatusSeeOther)
}

// checkCommandStatus periodically checks the status of a command execution
func checkCommandStatus(ssmClient *ssm.Client, commandID string, instanceID string, region string, command string) {
    // Wait a few seconds before starting to check status
    time.Sleep(5 * time.Second)
    
    // Check status every 10 seconds for up to 10 minutes
    for i := 0; i < 60; i++ {
        status, output, err := aws.GetCommandStatus(ssmClient, commandID, instanceID)
        if err != nil {
            log.Printf("Error checking command status for instance %s: %v", instanceID, err)
            updateCommandStatus(instanceID, "Error checking status", "", commandID, command)
            return
        }
        
        // Update the status in our map
        updateCommandStatus(instanceID, status, output, commandID, command)
        
        // If the command is no longer in progress, we're done
        if status != "InProgress" && status != "Pending" {
            return
        }
        
        // Wait before checking again
        time.Sleep(10 * time.Second)
    }
    
    // If we get here, the command has been running for too long
    updateCommandStatus(instanceID, "Timeout", "", commandID, command)
}

// updateCommandStatus safely updates the commandStatusMap for a specific instance ID
func updateCommandStatus(instanceID, status, output, commandID, command string) {
    commandStatusLock.Lock()
    defer commandStatusLock.Unlock()
    commandStatusMap[instanceID] = CommandStatus{
        Status:    status,
        Output:    output,
        Timestamp: time.Now().Format(time.RFC3339),
        CommandID: commandID,
        Command:   command,
    }
}

// GetCommandStatusMap provides a thread-safe way to access the commandStatusMap
func GetCommandStatusMap() map[string]CommandStatus {
    commandStatusLock.Lock()
    defer commandStatusLock.Unlock()
    // Create a copy to avoid concurrent modification issues
    copyMap := make(map[string]CommandStatus)
    for k, v := range commandStatusMap {
        copyMap[k] = v
    }
    return copyMap
}

// CommandStatusHandler renders the command status page
func CommandStatusHandler(w http.ResponseWriter, r *http.Request) {
    isLoggedIn := auth.IsUserLoggedIn(r)

    // Safely retrieve a copy of the commandStatusMap
    currentStatusMap := GetCommandStatusMap()

    // Map command statuses to EC2Instance objects for rendering
    var instancesWithStatus []models.EC2Instance
    for id, cmdStatus := range currentStatusMap {
        instance, err := models.GetInstanceDetails(id)
        if err != nil {
            log.Printf("Error retrieving instance details for ID %s: %v", id, err)
            continue
        }

        // Add status, output, and timestamp to the instance details
        instance.State = cmdStatus.Status
        instance.CommandOutput = cmdStatus.Output
        instance.CommandTimestamp = cmdStatus.Timestamp
        instance.Command = cmdStatus.Command
        instancesWithStatus = append(instancesWithStatus, *instance)
    }

    // Prepare template data
    data := models.TemplateData{
        Title:     "Command Execution Status",
        IsLoggedIn: isLoggedIn,
        Instances: instancesWithStatus,
        Version:   config.Version,
    }

    // Load and parse the templates
    tmpl, err := template.ParseFiles("templates/command_status.html", "templates/layout.html")
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        log.Printf("Error loading templates: %v\n", err)
        return
    }

    // Render the template
    if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
        log.Printf("Error rendering command status page: %v\n", err)
        http.Error(w, "Error rendering command status page", http.StatusInternalServerError)
    }
}

