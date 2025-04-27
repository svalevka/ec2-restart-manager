// handlers/config_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"ec2-restart-manager/auth"
	"ec2-restart-manager/aws"
	"ec2-restart-manager/config"
	"ec2-restart-manager/models"

	ssm_sdk "github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	configSSMClient *ssm_sdk.Client
	environmentName string
)

// InjectSSMClient allows main.go to pass the global config SSM client
func InjectSSMClient(client *ssm_sdk.Client) {
	configSSMClient = client
}

// InjectEnvironment allows main.go to pass the environment name (dev/test/prod)
func InjectEnvironment(env string) {
	environmentName = env
}

// ConfigHandler displays and processes the configuration page
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := auth.IsUserLoggedIn(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	paramName := fmt.Sprintf("/ec2-restart-manager/%s/schedule", environmentName)

	var scheduleConfig models.ScheduleConfig

	// Load the current schedule from Parameter Store
	paramValue, err := aws.GetParameter(configSSMClient, paramName)
	if err != nil {
		log.Printf("Error loading schedule config from Parameter Store: %v", err)
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal([]byte(paramValue), &scheduleConfig); err != nil {
		log.Printf("Error parsing schedule config JSON: %v", err)
		http.Error(w, "Failed to parse configuration", http.StatusInternalServerError)
		return
	}

	// Handle form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		scheduleConfig.StgDevDay = r.FormValue("stg_dev_day")
		scheduleConfig.StgDevTime = r.FormValue("stg_dev_time")
		scheduleConfig.ProdDay = r.FormValue("prod_day")
		scheduleConfig.ProdTime = r.FormValue("prod_time")

		jsonData, err := json.MarshalIndent(scheduleConfig, "", "  ")
		if err != nil {
			log.Printf("Error serializing schedule config: %v", err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}

		if err := aws.PutParameter(configSSMClient, paramName, string(jsonData)); err != nil {
			log.Printf("Error saving schedule config to Parameter Store: %v", err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/config?updated=true", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := models.TemplateData{
		Title:      "Schedule Configuration",
		IsLoggedIn: isLoggedIn,
		Version:    config.Version,
		Data: map[string]interface{}{
			"ScheduleConfig": scheduleConfig,
			"Updated":        r.URL.Query().Get("updated") == "true",
			"Days":           []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
		},
	}

	tmpl, err := template.ParseFiles("templates/config.html", "templates/layout.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Error loading templates: %v\n", err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Error rendering config page: %v\n", err)
		http.Error(w, "Error rendering configuration page", http.StatusInternalServerError)
	}
}