package config

import "os"

var (
	Port             = getEnv("PORT", "8080")
	WorkingDir       = getEnv("WORKING_DIR", "/tmp/build-job")
	DockerImage      = getEnv("DOCKER_IMAGE", "golang:1.21")
	WebhookSecret    = getEnv("WEBHOOK_SECRET", "")
	DeployEnabled    = false
	DeployConfigPath = ""
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
