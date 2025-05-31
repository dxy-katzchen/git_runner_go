package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunDockerJob(workDir string) map[string]string {
	fmt.Println("Scanning for Docker services in repository...")

	// Track found services and built images
	foundServices := 0
	builtImages := make(map[string]string)

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
			imageName := buildDockerImage(serviceDir, path, serviceName)

			// Store successful builds
			if imageName != "" {
				builtImages[serviceName] = imageName
			}
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

	return builtImages
}

func buildDockerImage(servicePath, dockerfilePath, serviceName string) string {
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

	// Return the image name on success, empty string on failure
	if err != nil {
		fmt.Printf("Docker build failed for %s: %v\n", serviceName, err)
		return ""
	} else {
		fmt.Printf("Successfully built %s service\n", serviceName)
		imageName := fmt.Sprintf("project-%s:local", lowerServiceName)
		fmt.Printf("Image available as: %s\n", imageName)
		return imageName
	}

}
