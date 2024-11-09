// main.go
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
	"ec2-restart-manager/utils"
)

func main() {

	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AWS configuration and clients
	if err := aws.InitAWSConfig(); err != nil {
		log.Fatalf("Failed to initialize AWS configuration: %v", err)
	}
	aws.SetupS3Client() // No need to handle error here since SetupS3Client does not return a value

	// Debug print configuration if enabled
	if utils.Debug {
		configJSON, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println("Config:", string(configJSON))
	}


	// Load the Azure app secret into the environment if the config environment is 'test'
	if cfg.Environment == "test" {
		secretValue, err := utils.LoadSecretToEnv(
			cfg.Environment,                   // Environment (e.g., "test")
			"platform/ec2-restart-manager",    // Secret name in AWS Secrets Manager
			"AZURE_AD_CLIENT_SECRET_TEST",     // Key within the secret
		)
		if err != nil {
			log.Fatalf("Failed to load secret: %v", err)
		}
		os.Setenv("AZURE_AD_CLIENT_SECRET", secretValue)
	}

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

	// Start the web server
	address := "0.0.0.0:8080"
	log.Printf("Server started at http://%s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
