package main

import (
	"flag"
	"git-runner/config"
	"git-runner/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	// Parse command line flags
	configFlag := flag.String("config", "", "Path to deployment configuration file")
	deployFlag := flag.Bool("deploy", false, "Enable deployment after build")
	flag.Parse()

	// Check environment variables
	deployEnabled := *deployFlag
	if os.Getenv("ENABLE_ECS_DEPLOY") == "true" {
		deployEnabled = true
	}

	// Initialize deployment configuration
	configPath, deployEnabled := config.InitDeployment(configFlag, deployEnabled)

	// Store values in config package
	config.DeployEnabled = deployEnabled
	config.DeployConfigPath = configPath

	// Set up HTTP server
	http.HandleFunc("/webhook", handler.WebhookHandler)
	log.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
