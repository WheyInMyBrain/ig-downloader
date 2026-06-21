package main

import (
	"fmt"
	"os"
)

func main() {
	config, err := ParseCLIProfileURL()
	if err != nil {
		fmt.Printf("[CLI Error] %v\n", err)
		os.Exit(1)
	}

	// Route control flow context based on flags parsed
	if config.ServeUI {
		StartLocalWebServer()
	} else {
		OrchestrateEngine()
	}
}