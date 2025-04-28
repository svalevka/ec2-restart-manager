package handlers

import (
    "html/template"
    "net/http"
    "log"

    "ec2-restart-manager/models"
    "ec2-restart-manager/auth"
    "ec2-restart-manager/config"
)

// StatusHandler renders the status page, showing the status of each instance restart
func StatusHandler(w http.ResponseWriter, r *http.Request) {
    isLoggedIn := auth.IsUserLoggedIn(r)

    // Safely retrieve a copy of the statusMap
    currentStatusMap := GetStatusMap()

    // Map instance statuses to EC2Instance objects for rendering
    var instancesWithStatus []models.EC2Instance
    for id, instanceStatus := range currentStatusMap {
        instance, err := models.GetInstanceDetails(id)
        if err != nil {
            log.Printf("Error retrieving instance details for ID %s: %v", id, err)
            continue
        }

        // Add status and timestamp to the instance details
        instance.State = instanceStatus.Status               // Assign the status
        instance.RestartTimestamp = instanceStatus.Timestamp // Assign the timestamp
        instancesWithStatus = append(instancesWithStatus, *instance)
    }

    // Prepare template data
    data := models.TemplateData{
        Title:     "Instance Restart Status",
        IsLoggedIn: isLoggedIn,
        Instances: instancesWithStatus, // Use the Instances field to pass instance data
        Version:  config.Version,
    }

    // Load and parse the templates
    tmpl, err := template.ParseFiles("templates/status.html", "templates/layout.html")
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        log.Printf("Error loading templates: %v\n", err)
        return
    }

    // Render the template
    if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
        log.Printf("Error rendering status page: %v\n", err)
        http.Error(w, "Error rendering status page", http.StatusInternalServerError)
    }
}
