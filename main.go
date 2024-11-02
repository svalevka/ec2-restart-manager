// main.go
package main

import (
    "log"
    "net/http"

    "ec2-restart-manager/aws"
    "ec2-restart-manager/handlers"
)

func main() {
    // Initialize AWS clients
    aws.SetupAWSClients()

    // Define HTTP routes
    http.HandleFunc("/", handlers.IndexHandler)
    http.HandleFunc("/restart", handlers.RestartHandler)

    // Start the web server
    address := "0.0.0.0:8080"
    log.Printf("Server started at http://%s", address)
    log.Fatal(http.ListenAndServe(address, nil))
}
