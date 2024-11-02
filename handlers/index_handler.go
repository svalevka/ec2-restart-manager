// handlers/index_handler.go
package handlers

import (
	"fmt"
	"net/http"
)

// IndexHandler handles requests to the root URL ("/")
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the EC2 Restart Manager!")
}
