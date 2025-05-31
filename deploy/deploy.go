package deploy

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// DeploymentConfig represents the deployment configuration
type DeploymentConfig struct {
	Provider string                   `yaml:"provider"`
	AWS      AWSConfig                `yaml:"aws"`
	Services map[string]ServiceConfig `yaml:"services"`
}

// AWSConfig contains AWS-specific configuration
type AWSConfig struct {
	Region        string `yaml:"region"`
	ECRRepository string `yaml:"ecrRepositoryPrefix"`
	ECSCluster    string `yaml:"ecsCluster"`
	AccountID     string `yaml:"accountId,omitempty"` // Optional, can be set via environment variable
}

// ServiceConfig defines a deployable service
type ServiceConfig struct {
	Directory      string `yaml:"directory"`      // Service directory in repo
	TaskDefinition string `yaml:"taskDefinition"` // ECS task definition name
	ServiceName    string `yaml:"serviceName"`    // ECS service name
	ContainerName  string `yaml:"containerName"`  // Container name in task definition
}

// LoadConfig loads configuration with environment variable support
func LoadConfig(configPath string, workDir string) (*DeploymentConfig, error) {

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Process environment variables in the YAML
	processedData := processEnvVars(string(data))

	// Parse YAML
	var config DeploymentConfig
	if err := yaml.Unmarshal([]byte(processedData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// processEnvVars replaces ${ENV_VAR} placeholders with environment variable values
func processEnvVars(content string) string {
	re := regexp.MustCompile(`\${([A-Za-z0-9_]+)}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract env var name without ${ and }
		envVar := match[2 : len(match)-1]
		if value := os.Getenv(envVar); value != "" {
			return value
		}
		// Return original if env var not found
		return match
	})
}

// loginToECR authenticates with Amazon ECR
func loginToECR(awsConfig AWSConfig) error {
	fmt.Println("Logging in to Amazon ECR...")

	// Check if required AWS credentials are available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		return fmt.Errorf("AWS credentials not found in environment variables")
	}

	// Get ECR login password
	cmd := exec.Command("aws", "ecr", "get-login-password",
		"--region", awsConfig.Region)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ECR login password: %v", err)
	}

	// Get AWS account ID from config or environment
	accountID := awsConfig.AccountID
	if accountID == "" {
		accountID = os.Getenv("AWS_ACCOUNT_ID")
		if accountID == "" {
			return fmt.Errorf("AWS account ID not specified")
		}
	}

	// Login to Docker
	registryURL := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, awsConfig.Region)

	loginCmd := exec.Command("docker", "login",
		"--username", "AWS",
		"--password-stdin",
		registryURL)

	loginCmd.Stdin = strings.NewReader(string(output))
	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("docker login failed: %v", err)
	}

	fmt.Printf("Successfully logged in to ECR at %s\n", registryURL)
	return nil
}

// pushImagesToECR tags and pushes Docker images to Amazon ECR
func pushImagesToECR(builtImages map[string]string, config *DeploymentConfig) (map[string]string, error) {
	fmt.Println("Pushing images to ECR...")

	// Generate a unique tag for all images in this deployment
	imageTag := time.Now().Format("20060102-150405")
	pushedImages := make(map[string]string)

	// Get AWS account ID
	accountID := config.AWS.AccountID
	if accountID == "" {
		accountID = os.Getenv("AWS_ACCOUNT_ID")
	}

	// For each built image
	for serviceName, localImage := range builtImages {

		// Find the service config
		serviceConfig, found := config.Services[serviceName]
		if !found {
			fmt.Printf("Warning: No configuration found for service %s, using defaults\n", serviceName)
			serviceConfig = ServiceConfig{
				Directory:      serviceName,
				TaskDefinition: fmt.Sprintf("%s-%s", config.AWS.ECRRepository, serviceName),
				ServiceName:    fmt.Sprintf("%s-%s-service", config.AWS.ECRRepository, serviceName),
				ContainerName:  serviceName,
				// Update the config with the default serviceConfig
			}
			config.Services[serviceName] = serviceConfig
		}

		// Create ECR repository URI
		ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s/%s",
			accountID,
			config.AWS.Region,
			config.AWS.ECRRepository,
			serviceConfig.ServiceName) // Use serviceConfig.ServiceName here

		remoteImageURI := fmt.Sprintf("%s:%s", ecrRepoURI, imageTag)

		// Tag the image
		fmt.Printf("Tagging %s as %s\n", localImage, remoteImageURI)
		tagCmd := exec.Command("docker", "tag", localImage, remoteImageURI)
		if out, err := tagCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to tag image %s: %v\n%s", serviceName, err, out)
		}

		// Push the image to ECR
		fmt.Printf("Pushing %s to ECR...\n", remoteImageURI)
		pushCmd := exec.Command("docker", "push", remoteImageURI)
		if out, err := pushCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to push image %s: %v\n%s", serviceName, err, out)
		}

		// Also tag as latest
		latestTag := fmt.Sprintf("%s:latest", ecrRepoURI)
		if tagLatestCmd := exec.Command("docker", "tag", localImage, latestTag); tagLatestCmd.Run() == nil {
			exec.Command("docker", "push", latestTag).Run() // Best effort push of latest
		}

		// Store the pushed image URI for later use
		pushedImages[serviceName] = remoteImageURI
		fmt.Printf("Successfully pushed %s\n", remoteImageURI)
	}

	return pushedImages, nil
}

// updateAndDeployServices updates task definitions and deploys to ECS
func updateAndDeployServices(pushedImages map[string]string, config *DeploymentConfig) error {
	fmt.Println("Updating task definitions and deploying services...")

	for serviceName, imageURI := range pushedImages {
		serviceConfig, found := config.Services[serviceName]
		if !found {
			fmt.Printf("Warning: Skipping deployment for %s, no configuration found\n", serviceName)
			continue
		}

		// 1. Download current task definition
		taskDefFile := fmt.Sprintf("%s-task-def.json", serviceName)
		fmt.Printf("Downloading task definition for %s...\n", serviceConfig.TaskDefinition)

		describeCmd := exec.Command("aws", "ecs", "describe-task-definition",
			"--task-definition", serviceConfig.TaskDefinition,
			"--query", "taskDefinition",
			"--region", config.AWS.Region,
			"--output", "json")

		taskDefOutput, err := os.Create(taskDefFile)
		if err != nil {
			return fmt.Errorf("failed to create task definition file: %v", err)
		}

		describeCmd.Stdout = taskDefOutput
		if err := describeCmd.Run(); err != nil {
			taskDefOutput.Close()
			return fmt.Errorf("failed to download task definition: %v", err)
		}
		taskDefOutput.Close()

		// 2. Update task definition with new image
		fmt.Printf("Updating task definition with image: %s\n", imageURI)

		// Create container definitions update
		containerDefUpdate := fmt.Sprintf(`[{"name":"%s","image":"%s"}]`,
			serviceConfig.ContainerName,
			imageURI)

		registerCmd := exec.Command("aws", "ecs", "register-task-definition",
			"--cli-input-json", "file://"+taskDefFile,
			"--container-definitions", containerDefUpdate,
			"--region", config.AWS.Region)

		registerOutput, err := registerCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update task definition: %v\n%s", err, registerOutput)
		}

		// 3. Deploy updated task definition to service
		fmt.Printf("Deploying service %s with updated task definition\n", serviceConfig.ServiceName)

		updateServiceCmd := exec.Command("aws", "ecs", "update-service",
			"--cluster", config.AWS.ECSCluster,
			"--service", serviceConfig.ServiceName,
			"--task-definition", serviceConfig.TaskDefinition,
			"--force-new-deployment",
			"--region", config.AWS.Region)

		updateOutput, err := updateServiceCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update service: %v\n%s", err, updateOutput)
		}

		fmt.Printf("Successfully deployed %s\n", serviceConfig.ServiceName)
	}

	return nil
}

// DeployToECS now accepts config path as a parameter
func DeployToECS(workDir string, builtImages map[string]string, configPath string) error {

	config, err := LoadConfig(configPath, workDir)
	if err != nil {
		return fmt.Errorf("failed to load deployment config: %v", err)
	}

	if err := loginToECR(config.AWS); err != nil {
		return fmt.Errorf("failed to login to ECR: %v", err)
	}

	pushedImages, err := pushImagesToECR(builtImages, config)
	if err != nil {
		return fmt.Errorf("failed to push images to ECR: %v", err)
	}

	if err := updateAndDeployServices(pushedImages, config); err != nil {
		return fmt.Errorf("failed to update and deploy services: %v", err)
	}

	return nil
}
