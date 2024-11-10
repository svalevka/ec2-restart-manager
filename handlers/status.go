package handlers

import (
    "html/template"
    "net/http"
	
    "ec2-restart-manager/models"
    "ec2-restart-manager/auth"
    "log"
)


// StatusHandler renders the status page, showing the status of each instance restart
func StatusHandler(w http.ResponseWriter, r *http.Request) {
    isLoggedIn := auth.IsUserLoggedIn(r)

    // Acquire the lock before reading from statusMap
    statusLock.Lock()
    defer statusLock.Unlock()

    // Prepare template data
    data := models.TemplateData{
        Title:      "Instance Restart Status",
        StatusMap:  statusMap, // Pass the statusMap to the template
        IsLoggedIn: isLoggedIn,
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
