// handlers/command_handler.go
package handlers

import (
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"
    "html/template"

    "ec2-restart-manager/aws"
    "ec2-restart-manager/models"
    "ec2-restart-manager/auth"
    "ec2-restart-manager/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Region to timezone mapping
var regionTimezoneMap = map[string]string{
    // UK
    "eu-west-2": "Europe/London",
    // Ireland
    "eu-west-1": "Europe/Dublin",
    // Frankfurt
    "eu-central-1": "Europe/Berlin",
    // Paris
    "eu-west-3": "Europe/Paris",
    // Stockholm
    "eu-north-1": "Europe/Stockholm",
    // Singapore
    "ap-southeast-1": "Asia/Singapore",
    // Hong Kong
    "ap-east-1": "Asia/Hong_Kong",
    // Dubai (Middle East)
    "me-central-1": "Asia/Dubai",
    // Tokyo
    "ap-northeast-1": "Asia/Tokyo",
    // US East (N. Virginia)
    "us-east-1": "America/New_York",
    // US East (Ohio)
    "us-east-2": "America/New_York",
    // US West (N. California)
    "us-west-1": "America/Los_Angeles",
    // US West (Oregon)
    "us-west-2": "America/Los_Angeles",
    // Default
    "default": "UTC",
}

// Day mapping for converting weekday names to systemd day names
var dayMap = map[string]string{
    "Monday": "Mon",
    "Tuesday": "Tue",
    "Wednesday": "Wed",
    "Thursday": "Thu",
    "Friday": "Fri",
    "Saturday": "Sat",
    "Sunday": "Sun",
}

// getUTCTimeFromRegional converts a time in a regional timezone to UTC for systemd timer
func getUTCTimeFromRegional(day string, timeStr string, regionTimezone string) (string, string, error) {
    // Parse the time
    timeParts := strings.Split(timeStr, ":")
    if len(timeParts) != 2 {
        return "", "", fmt.Errorf("invalid time format: %s", timeStr)
    }
    
    hour, err := strconv.Atoi(timeParts[0])
    if err != nil {
        return "", "", fmt.Errorf("invalid hour: %s", timeParts[0])
    }
    
    minute, err := strconv.Atoi(timeParts[1])
    if err != nil {
        return "", "", fmt.Errorf("invalid minute: %s", timeParts[1])
    }
    
    // Find day number
    dayNum := 0
    switch day {
    case "Monday":
        dayNum = 1
    case "Tuesday":
        dayNum = 2
    case "Wednesday":
        dayNum = 3
    case "Thursday":
        dayNum = 4
    case "Friday":
        dayNum = 5
    case "Saturday":
        dayNum = 6
    case "Sunday":
        dayNum = 0
    default:
        return "", "", fmt.Errorf("invalid day: %s", day)
    }
    
    // Create a time.Time for this week's instance of that day and time
    now := time.Now()
    // Find the most recent occurrence of the target day
    daysToAdd := (7 + dayNum - int(now.Weekday())) % 7
    targetDay := now.AddDate(0, 0, daysToAdd)
    
    // Create the time in the region's timezone
    regionLoc, err := time.LoadLocation(regionTimezone)
    if err != nil {
        return "", "", fmt.Errorf("invalid region timezone: %s", regionTimezone)
    }
    
    // Create time in the regional timezone
    regionTime := time.Date(
        targetDay.Year(), targetDay.Month(), targetDay.Day(),
        hour, minute, 0, 0, regionLoc,
    )
    
    // Convert to UTC for the systemd timer
    utcTime := regionTime.UTC()
    
    // Extract the day of week and time in UTC
    utcDay := dayMap[utcTime.Weekday().String()]
    if utcDay == "" {
        utcDay = utcTime.Weekday().String()[:3]
    }
    
    utcTimeStr := fmt.Sprintf("%02d:%02d", utcTime.Hour(), utcTime.Minute())
    
    return utcDay, utcTimeStr, nil
}

var command_role_name = "ec2-restart-manager-restarter"

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
    // First refresh the schedule configuration
    if err := models.LoadScheduleConfig(); err != nil {
        log.Printf("Error refreshing schedule configuration: %v", err)
        // Continue anyway with the cached config
    } else {
        log.Printf("Successfully refreshed schedule configuration before command execution")
    }

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

    // Get the schedule configuration (now guaranteed to be fresh)
    scheduleConfig := models.GetScheduleConfig()
    log.Printf("Using schedule config: Dev/Stg day=%s time=%s, Prod day=%s time=%s", 
               scheduleConfig.StgDevDay, scheduleConfig.StgDevTime, 
               scheduleConfig.ProdDay, scheduleConfig.ProdTime)

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
                // For staging/dev: Use systemd-run with configured day/time with reboot
                // Get the instance's region and determine timezone
                timezone := regionTimezoneMap[instance.Region]
                if timezone == "" {
                    timezone = regionTimezoneMap["default"] // Use UTC if region not found
                }
                
                // Convert regional time to UTC for systemd timer
                utcDay, utcTime, err := getUTCTimeFromRegional(
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime, 
                    timezone, // The regional timezone for this instance
                )
                if err != nil {
                    log.Printf("Error converting timezone for instance %s in region %s: %v", instanceID, instance.Region, err)
                    updateCommandStatus(instanceID, "Failed to convert timezone", "", "", "")
                    continue
                }
                
                // Format the systemd calendar specification with UTC time
                systemdCalendar := fmt.Sprintf("%s %s", utcDay, utcTime)
                
                // We still use TZ for the script execution to ensure all log timestamps use regional time
                command = fmt.Sprintf(`sudo systemd-run --on-calendar="%s" --unit=security-update-stgdev /bin/bash -c 'export TZ=%s; exec >> /var/log/patching.log 2>&1; echo ""; echo "=== NEW SECURITY UPDATE RUN: $(date) ==="; echo "SCHEDULED-UPDATE starting at $(date)"; SLEEP_TIME=$((RANDOM %% 1800)); echo "Will sleep for $SLEEP_TIME seconds and update at $(date -d "+$SLEEP_TIME seconds")"; sleep $SLEEP_TIME; echo "Starting security update at $(date)"; %s && echo "SCHEDULED-UPDATE completed at $(date), rebooting now" && sudo reboot'`,
                    systemdCalendar,
                    timezone,
                    baseCommand)
                commandName = fmt.Sprintf("Scheduled Security Patching (%s %s GMT with reboot)", 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
            } else if envClass == "prod" {
                // For prod: Use systemd-run with configured day/time without reboot
                // Get the instance's region and determine timezone
                timezone := regionTimezoneMap[instance.Region]
                if timezone == "" {
                    timezone = regionTimezoneMap["default"] // Use UTC if region not found
                }
                
                // Convert regional time to UTC for systemd timer
                utcDay, utcTime, err := getUTCTimeFromRegional(
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime, 
                    timezone, // The regional timezone for this instance
                )
                if err != nil {
                    log.Printf("Error converting timezone for instance %s in region %s: %v", instanceID, instance.Region, err)
                    updateCommandStatus(instanceID, "Failed to convert timezone", "", "", "")
                    continue
                }
                
                // Format the systemd calendar specification with UTC time
                systemdCalendar := fmt.Sprintf("%s %s", utcDay, utcTime)
                
                // We still use TZ for the script execution to ensure all log timestamps use regional time
                command = fmt.Sprintf(`sudo systemd-run --on-calendar="%s" --unit=security-update-prod /bin/bash -c 'export TZ=%s; exec >> /var/log/patching.log 2>&1; echo ""; echo "=== NEW SECURITY UPDATE RUN: $(date) ==="; echo "SCHEDULED-UPDATE starting at $(date)"; SLEEP_TIME=$((RANDOM %% 1800)); echo "Will sleep for $SLEEP_TIME seconds and update at $(date -d "+$SLEEP_TIME seconds")"; sleep $SLEEP_TIME; echo "Starting security update at $(date)"; %s && echo "SCHEDULED-UPDATE completed at $(date)"'`,
                    systemdCalendar,
                    timezone,
                    baseCommand)
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
                // For staging/dev: Use systemd-run with configured day/time with reboot
                // Get the instance's region and determine timezone
                timezone := regionTimezoneMap[instance.Region]
                if timezone == "" {
                    timezone = regionTimezoneMap["default"] // Use UTC if region not found
                }
                
                // Convert regional time to UTC for systemd timer
                utcDay, utcTime, err := getUTCTimeFromRegional(
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime, 
                    timezone, // The regional timezone for this instance
                )
                if err != nil {
                    log.Printf("Error converting timezone for instance %s in region %s: %v", instanceID, instance.Region, err)
                    updateCommandStatus(instanceID, "Failed to convert timezone", "", "", "")
                    continue
                }
                
                // Format the systemd calendar specification with UTC time
                systemdCalendar := fmt.Sprintf("%s %s", utcDay, utcTime)
                
                // We still use TZ for the script execution to ensure all log timestamps use regional time
                command = fmt.Sprintf(`sudo systemd-run --on-calendar="%s" --unit=upgrade-stgdev /bin/bash -c 'export TZ=%s; exec >> /var/log/patching.log 2>&1; echo ""; echo "=== NEW SYSTEM UPGRADE RUN: $(date) ==="; echo "SCHEDULED-UPDATE starting at $(date)"; SLEEP_TIME=$((RANDOM %% 1800)); echo "Will sleep for $SLEEP_TIME seconds and upgrade at $(date -d "+$SLEEP_TIME seconds")"; sleep $SLEEP_TIME; echo "Starting system upgrade at $(date)"; %s && echo "SCHEDULED-UPDATE completed at $(date), rebooting now" && sudo reboot'`,
                    systemdCalendar,
                    timezone,
                    baseCommand)
                commandName = fmt.Sprintf("Scheduled System Upgrade (%s %s GMT with reboot)", 
                    scheduleConfig.StgDevDay, 
                    scheduleConfig.StgDevTime)
            } else if envClass == "prod" {
                // For prod: Use systemd-run with configured day/time without reboot
                // Get the instance's region and determine timezone
                timezone := regionTimezoneMap[instance.Region]
                if timezone == "" {
                    timezone = regionTimezoneMap["default"] // Use UTC if region not found
                }
                
                // Convert regional time to UTC for systemd timer
                utcDay, utcTime, err := getUTCTimeFromRegional(
                    scheduleConfig.ProdDay, 
                    scheduleConfig.ProdTime, 
                    timezone, // The regional timezone for this instance
                )
                if err != nil {
                    log.Printf("Error converting timezone for instance %s in region %s: %v", instanceID, instance.Region, err)
                    updateCommandStatus(instanceID, "Failed to convert timezone", "", "", "")
                    continue
                }
                
                // Format the systemd calendar specification with UTC time
                systemdCalendar := fmt.Sprintf("%s %s", utcDay, utcTime)
                
                // We still use TZ for the script execution to ensure all log timestamps use regional time
                command = fmt.Sprintf(`sudo systemd-run --on-calendar="%s" --unit=upgrade-prod /bin/bash -c 'export TZ=%s; exec >> /var/log/patching.log 2>&1; echo ""; echo "=== NEW SYSTEM UPGRADE RUN: $(date) ==="; echo "SCHEDULED-UPDATE starting at $(date)"; SLEEP_TIME=$((RANDOM %% 1800)); echo "Will sleep for $SLEEP_TIME seconds and upgrade at $(date -d "+$SLEEP_TIME seconds")"; sleep $SLEEP_TIME; echo "Starting system upgrade at $(date)"; %s && echo "SCHEDULED-UPDATE completed at $(date)"'`,
                    systemdCalendar,
                    timezone,
                    baseCommand)
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