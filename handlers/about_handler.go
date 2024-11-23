package handlers

import (
    "html/template"
    "net/http"
    "ec2-restart-manager/auth"
    "ec2-restart-manager/models"
    "ec2-restart-manager/config"
	"log"
)

// AboutHandler renders the about page.
func AboutHandler(w http.ResponseWriter, r *http.Request) {
    // Use the helper function to determine if the user is logged in
    isLoggedIn := auth.IsUserLoggedIn(r)

    // Prepare the template data
    data := models.TemplateData{
        Title:             "About",
        Version:           config.Version,
        IsLoggedIn:        isLoggedIn,
        AzureAuthenticated: false,
    }

    // Load and parse the templates
    tmpl, err := template.ParseFiles("templates/about.html", "templates/layout.html")
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        return
    }

    // Execute the template with the data
    if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
        log.Printf("Error rendering about page: %v\n", err) // Log detailed error message
        http.Error(w, "Error rendering about page", http.StatusInternalServerError)
    }
}
