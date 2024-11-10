package handlers

import (
	"net/http"
	"html/template"
	"log"

	"ec2-restart-manager/aws"
	"ec2-restart-manager/config"
	"ec2-restart-manager/models"
	"ec2-restart-manager/utils"
    "ec2-restart-manager/auth"
)

// Parse layout.html and index.html together
var indexTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))

// Global variable to hold the configuration
var cfg *config.EnvConfig

// Initialize configuration when the package is loaded
func init() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
}

// Updated code for IndexHandler in index_handler.go
func IndexHandler(w http.ResponseWriter, r *http.Request) {
    // Fetch CSV from S3
    bucket := cfg.S3.Bucket
    key := cfg.S3.Key

    csvContent, err := aws.GetCSVFromS3(bucket, key)
    if err != nil {
        http.Error(w, "Failed to retrieve CSV from S3", http.StatusInternalServerError)
        log.Printf("Error retrieving CSV: %v", err)
        return
    }

    // Parse CSV into instances
    instances, err := utils.ParseCSVToStruct(csvContent)
    if err != nil {
        http.Error(w, "Failed to parse CSV data", http.StatusInternalServerError)
        log.Printf("Error parsing CSV: %v", err)
        return
    }

    // Load parsed instances into global cache
    models.LoadInstances(instances)

    // Extract unique values for filters
    uniqueOwners := utils.GetUniqueOwners(instances)
    uniqueServices := utils.GetUniqueServices(instances)
    uniqueAWSAccountNames := utils.GetUniqueAWSAccountNames(instances)
    uniqueRegions := utils.GetUniqueRegions(instances)

    // Initialize variables for filtering
    filteredInstances := instances
    selectedOwner := ""
    selectedService := ""
    selectedAWSAccountName := ""
    selectedRegion := ""

    // Handle filtering based on user input
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Failed to parse form data", http.StatusBadRequest)
            log.Printf("Error parsing form data: %v", err)
            return
        }

        // Retrieve selected filter values from the form
        selectedOwner = r.FormValue("owner")
        selectedService = r.FormValue("service")
        selectedAWSAccountName = r.FormValue("awsAccountName")
        selectedRegion = r.FormValue("region")

        // Apply filters to the instances
        filteredInstances = utils.FilterInstances(instances, selectedOwner, selectedService, selectedAWSAccountName, selectedRegion)
    }

    // Check if the user is logged in by looking for the session ID cookie
    sessionID, err := r.Cookie("session_id")
    isLoggedIn := err == nil && auth.SessionStore[sessionID.Value] != ""

    // Retrieve the user's name from the session store if logged in
    userName := ""
    if isLoggedIn {
        userName = auth.SessionStore[sessionID.Value]
    }

    // Prepare data to pass to the template
    data := models.TemplateData{
        Title:                 "EC2 Instance Manager",
        Instances:             filteredInstances,
        UniqueOwners:          uniqueOwners,
        SelectedOwner:         selectedOwner,
        UniqueServices:        uniqueServices,
        SelectedService:       selectedService,
        UniqueAWSAccountNames: uniqueAWSAccountNames,
        SelectedAWSAccountName: selectedAWSAccountName,
        UniqueRegions:         uniqueRegions,
        SelectedRegion:        selectedRegion,
        IsLoggedIn:            isLoggedIn, // Pass login status to the template
        UserName:              userName,   // Pass the user’s name to the template
    }

    // Render layout.html with index.html as the content
    if err := indexTemplate.ExecuteTemplate(w, "layout", data); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
        log.Printf("Error rendering index template: %v", err)
    }
}


