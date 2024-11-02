package handlers

import (
	"net/http"

	"ec2-restart-manager/aws"
	"ec2-restart-manager/models"
	"ec2-restart-manager/utils"
	"html/template"
	"log"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch CSV from S3
	bucket := "inventory-copper-test"
	key := "ec2_inventory-current.csv"

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
	

    // Prepare data to pass to the template
    data := models.TemplateData{
        Instances:              filteredInstances,
        UniqueOwners:           uniqueOwners,
        SelectedOwner:          selectedOwner,
        UniqueServices:         uniqueServices,
        SelectedService:        selectedService,
        UniqueAWSAccountNames:  uniqueAWSAccountNames,
        SelectedAWSAccountName: selectedAWSAccountName,
        UniqueRegions:          uniqueRegions,
        SelectedRegion:         selectedRegion,
    }

	// Render the template
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}
