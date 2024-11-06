// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ec2-restart-manager/auth"
	"ec2-restart-manager/aws"
	"ec2-restart-manager/config"
	"ec2-restart-manager/handlers"
	"ec2-restart-manager/utils"
)

func main() {
    // Initialize environments
    
    // Load the configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    if utils.Debug {
    // Pretty-print the configuration as JSON
       configJSON, _ := json.MarshalIndent(cfg, "", "  ")
       fmt.Println("Config:", string(configJSON))
    }
    
    // Load the secret into the environment if the config environment is 'test'
     err = utils.LoadSecretToEnv(
            cfg.Environment,                   // Environment (e.g., "test")
            "platform/ec2-restart-manager",      // Secret name in AWS Secrets Manager
            "AZURE_AD_CLIENT_SECRET_TEST",      // Key within the secret
            "AZURE_AD_CLIENT_SECRET",           // Environment variable to set
    )
    if err != nil {
            log.Fatalf("Failed to load secret: %v", err)
        }
    
    // Print the value of AZURE_AD_CLIENT_SECRET environment variable after retrieval
    // if utils.Debug {
    //     secretValue := os.Getenv("AZURE_AD_CLIENT_SECRET")
    //     fmt.Printf("AZURE_AD_CLIENT_SECRET: %s\n", secretValue)
    // }

    // Initialize the authentication module
    auth.InitializeAuth(cfg)

    // Initialize AWS clients
    aws.SetupAWSClients()

	// Apply authMiddleware to protected routes
	http.Handle("/", auth.AuthMiddleware(http.HandlerFunc(handlers.IndexHandler)))
    http.Handle("/restart", auth.AuthMiddleware(http.HandlerFunc(handlers.RestartHandler)))
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/auth/callback", auth.CallbackHandler)

    // Start the web server
    address := "0.0.0.0:8080"
    log.Printf("Server started at http://%s", address)
    log.Fatal(http.ListenAndServe(address, nil))
}
