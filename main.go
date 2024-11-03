// main.go
package main

import (
    "log"
    "net/http"

    "ec2-restart-manager/aws"
    "ec2-restart-manager/handlers"
    "ec2-restart-manager/auth"
    "ec2-restart-manager/config"
)

func main() {
    // Initialize AWS clients
    // Load the configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

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
