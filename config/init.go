package config

import (
	"fmt"
	"log"
	"os"
)

// InitDeployment initializes deployment configuration
func InitDeployment(configFlag *string, deployEnabled bool) (string, bool) {
	// If deployment is enabled but no config file specified, try to find one
	if deployEnabled && *configFlag == "" {
		log.Println("Deployment enabled but no config file specified. Searching for config files...")

		configPath := FindConfigFile()
		if configPath != "" {
			*configFlag = configPath
			return *configFlag, true
		}

		// Handle missing config
		return *configFlag, promptForDeployment()
	}

	return *configFlag, deployEnabled
}

// FindConfigFile searches for deployment config files in common locations
func FindConfigFile() string {
	// Common config file locations to search
	configCandidates := []string{
		"deploy.yml", "deploy.yaml",
		".deploy/config.yml", ".deploy/config.yaml",
		"config/deploy.yml", "config/deploy.yaml",
		".github/workflows/deploy-config.yml",
	}

	// Check each candidate path
	for _, candidate := range configCandidates {
		if _, err := os.Stat(candidate); err == nil {
			log.Printf("Found config file: %s\n", candidate)
			return candidate
		}
	}

	log.Println("No deployment config file found.")
	return ""
}

// promptForDeployment asks the user whether to disable deployment
func promptForDeployment() bool {
	fmt.Print("Deployment requires a config file. Disable deployment? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "" {
		log.Println("Deployment disabled.")
		return false
	}

	log.Fatal("Cannot proceed with deployment without a config file. Please provide one using -config flag.")
	return false // Never reached due to log.Fatal
}
