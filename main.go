<<<<<<< HEAD
// Modified code for main.go (relevant sections only)
=======
// main.go
>>>>>>> patch-manager/main
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"ec2-restart-manager/auth"
	"ec2-restart-manager/aws"
	"ec2-restart-manager/config"
	"ec2-restart-manager/handlers"
<<<<<<< HEAD
	"ec2-restart-manager/models"
	"ec2-restart-manager/utils"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var configSSMClient *ssm.Client

func main() {
	// Load the app configuration
=======
	"ec2-restart-manager/utils"
)

func main() {

	// Load the configuration
>>>>>>> patch-manager/main
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

<<<<<<< HEAD
	// Initialize AWS session
	if err := aws.InitAWSConfig(); err != nil {
		log.Fatalf("Failed to initialize AWS configuration: %v", err)
	}
	aws.SetupS3Client()

	// Create SSM client for config (fixed region eu-west-2)
	configSSMClient, err = aws.NewSSMClient(aws.AWSConfig, "eu-west-2")
	if err != nil {
		log.Fatalf("Failed to create config SSM client: %v", err)
	}

	// Inject SSM client and environment into models and handlers
	handlers.InjectSSMClient(configSSMClient)
	models.InjectSSMClient(configSSMClient)
	handlers.InjectEnvironment(cfg.Environment)
	models.InjectEnvName(cfg.Environment)

	// Load the schedule config from Parameter Store
	if err := models.LoadScheduleConfig(); err != nil {
		log.Printf("Error loading schedule configuration: %v", err)
	}

	// Debug configuration print
=======
	// Initialize AWS configuration and clients
	if err := aws.InitAWSConfig(); err != nil {
		log.Fatalf("Failed to initialize AWS configuration: %v", err)
	}
	aws.SetupS3Client() // No need to handle error here since SetupS3Client does not return a value

	// Debug print configuration if enabled
>>>>>>> patch-manager/main
	if utils.Debug {
		configJSON, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println("Config:", string(configJSON))
	}

<<<<<<< HEAD
	// If environment is "test", load Azure secret
	if cfg.Environment == "test" {
		secretValue, err := utils.LoadSecretToEnv(
			cfg.Environment,
			"platform/ec2-restart-manager",
			"AZURE_AD_CLIENT_SECRET_TEST",
=======

	// Load the Azure app secret into the environment if the config environment is 'test'
	if cfg.Environment == "test" {
		secretValue, err := utils.LoadSecretToEnv(
			cfg.Environment,                   // Environment (e.g., "test")
			"platform/ec2-restart-manager",    // Secret name in AWS Secrets Manager
			"AZURE_AD_CLIENT_SECRET_TEST",     // Key within the secret
>>>>>>> patch-manager/main
		)
		if err != nil {
			log.Fatalf("Failed to load secret: %v", err)
		}
		os.Setenv("AZURE_AD_CLIENT_SECRET", secretValue)
	}

<<<<<<< HEAD
	// Initialize authentication with the AzureAD config
	auth.InitializeAuth(cfg)

	// Setup HTTP routes
	http.HandleFunc("/", handlers.IndexHandler)
	http.Handle("/restart", auth.AuthMiddleware(http.HandlerFunc(handlers.RestartHandler)))
	http.HandleFunc("/about", handlers.AboutHandler)
	http.HandleFunc("/logout", auth.LogoutHandler)
	http.HandleFunc("/access_denied", handlers.AccessDeniedHandler)
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/auth/callback", auth.CallbackHandler)
	http.HandleFunc("/status", handlers.StatusHandler)
	http.Handle("/command", auth.AuthMiddleware(http.HandlerFunc(handlers.CommandHandler)))
	http.HandleFunc("/command-status", handlers.CommandStatusHandler)
	http.Handle("/config", auth.AuthMiddleware(http.HandlerFunc(handlers.ConfigHandler)))

	// Start web server
	address := "0.0.0.0:8080"
	log.Printf("Server started at http://%s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
=======
	// Initialize the authentication module
	auth.InitializeAuth(cfg)

	// Apply authMiddleware to protected routes
	http.HandleFunc("/", handlers.IndexHandler) // Allow public access to the index
	http.Handle("/restart", auth.AuthMiddleware(http.HandlerFunc(handlers.RestartHandler))) // Restrict restart to logged-in 
	http.HandleFunc("/about", handlers.AboutHandler)
	http.HandleFunc("/logout", auth.LogoutHandler)
	http.HandleFunc("/access_denied", handlers.AccessDeniedHandler)

	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/auth/callback", auth.CallbackHandler)
	http.HandleFunc("/status", handlers.StatusHandler)


	// Start the web server
	address := "0.0.0.0:8080"
	log.Printf("Server started at http://%s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
>>>>>>> patch-manager/main
