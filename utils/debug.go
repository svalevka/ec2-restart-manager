// util/debug.go
package utils

import (
	"fmt"
	"os"
	"strconv"
)

// Debug is a global variable indicating whether debug mode is enabled.
var Debug bool

// init initializes the Debug variable based on the DEBUG environment variable.
func init() {
	if debugEnv := os.Getenv("DEBUG"); debugEnv != "" {
		debugVal, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Printf("Invalid value for DEBUG environment variable: %v\n", err)
			Debug = false // Default to false if parsing fails
		} else {
			Debug = debugVal
		}
	}
}
