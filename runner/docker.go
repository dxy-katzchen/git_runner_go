package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunDockerJob(workDir string) {
	fmt.Println("Scanning for Docker services in repository...")

	// Track found services
	foundServices := 0

	// Walk through the repository to find Dockerfiles
	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Look for Dockerfile
		if !info.IsDir() && info.Name() == "Dockerfile" {
			// Get directory containing the Dockerfile
			serviceDir := filepath.Dir(path)

			// Generate service name from directory path
			relPath, _ := filepath.Rel(workDir, serviceDir)
			serviceName := strings.Replace(relPath, string(filepath.Separator), "-", -1)
			if serviceName == "." {
				serviceName = "root"
			}

			// Build Docker image
			foundServices++
			buildDockerImage(serviceDir, path, serviceName)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning repository: %v\n", err)
	}

	if foundServices == 0 {
		fmt.Println("No Dockerfiles found in the repository.")
	} else {
		fmt.Printf("Found and processed %d Docker services\n", foundServices)
	}
}

func buildDockerImage(servicePath, dockerfilePath, serviceName string) {
	fmt.Printf("Building service from %s...\n", servicePath)

	// Create a lowercase tag name for Docker (Docker requires lowercase repository names)
	lowerServiceName := strings.ToLower(serviceName)

	// Build image using the service's Dockerfile
	buildCmd := exec.Command("docker", "build",
		"-t", fmt.Sprintf("project-%s:local", lowerServiceName),
		"-f", dockerfilePath,
		servicePath)

	output, err := buildCmd.CombinedOutput()
	fmt.Printf("=== Build output for %s ===\n", serviceName)
	fmt.Println(string(output))

	if err != nil {
		fmt.Printf("Docker build failed for %s: %v\n", serviceName, err)
	} else {
		fmt.Printf("Successfully built %s service\n", serviceName)
		fmt.Printf("Image available as: project-%s:local\n", lowerServiceName)
	}
	fmt.Println("-----------------------------------")
}
