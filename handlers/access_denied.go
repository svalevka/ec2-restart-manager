package handlers

import (
	"html/template"
	"net/http"
	"ec2-restart-manager/models"
)

// AccessDeniedHandler handles scenarios where the user is authenticated with Azure
// but does not have the required group membership.
func AccessDeniedHandler(w http.ResponseWriter, r *http.Request) {
	// Prepare the template data with AzureAuthenticated set to true to display the logout option
	data := models.TemplateData{
		Title:             "Access Denied",
		IsLoggedIn:        false,                 // User is not logged into the app
		AzureAuthenticated: true,                 // User is authenticated with Azure
		UserName:          "",                    // Optionally, set a display name if available
	}

	// Load and parse the templates
	tmpl, err := template.ParseFiles("templates/access_denied.html", "templates/layout.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Execute the template with the data
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering access denied page", http.StatusInternalServerError)
	}
}
