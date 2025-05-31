package runner

import (
	"fmt"
	"git-runner/deploy"
	"git-runner/utils"
	"os"
)

// RunJob clones and builds the repository, optionally deploying it
func RunJob(repoURL, commitSHA string, deployEnabled bool, configPath string) {
	dir := "/tmp/build-job"
	os.RemoveAll(dir)

	err := utils.CloneAndCheckout(repoURL, commitSHA, dir)
	if err != nil {
		fmt.Println("Git error", err)
		return
	}

	fmt.Println("Cloned. Starting Docker job...")

	// Track built images if deployment is needed
	builtImages := make(map[string]string)

	// Modified RunDockerJob to return built images
	builtImages = RunDockerJob(dir)

	// Deploy if enabled
	if deployEnabled {
		fmt.Println("Deploying to ECS...")
		if err := deploy.DeployToECS(dir, builtImages, configPath); err != nil {
			fmt.Printf("Deployment failed: %v\n", err)
		}
	}
}
