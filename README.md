# Git Runner - Lightweight CI/CD Tool

A lightweight, webhook-driven CI/CD tool that integrates with GitHub to automatically build and deploy containerized applications. Built in Go with Docker and AWS ECS support.

## 🚀 Features

- **GitHub Webhook Integration**: Secure webhook handling with signature verification
- **Docker Multi-Service Support**: Automatically discovers and builds multiple Docker services in a repository
- **AWS ECS Deployment**: Automated deployment to Amazon ECS with image registry management
- **Real-time Processing**: Asynchronous job processing with immediate webhook response
- **Flexible Configuration**: Environment variable and YAML-based configuration
- **Security First**: HMAC signature verification for webhook authenticity

## 📋 Prerequisites

- Go 1.24.3 or higher
- Docker installed and running
- AWS CLI configured (for deployment features)
- ngrok (for local webhook testing)

## 🛠️ Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd git_runner
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Server Configuration
PORT=8080
WORKING_DIR=/tmp/build-job

# Docker Configuration
DOCKER_IMAGE=golang:1.21

# GitHub Webhook Security
WEBHOOK_SECRET=your_webhook_secret_here

# AWS Configuration (for deployment)
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_ACCOUNT_ID=your_account_id
ENABLE_ECS_DEPLOY=true
```

### Deployment Configuration

Create a `deploy.yml` file for AWS ECS deployment:

```yaml
provider: aws
aws:
  region: us-west-2
  ecrRepositoryPrefix: my-app
  ecsCluster: my-cluster
  accountId: ${AWS_ACCOUNT_ID}

services:
  api:
    directory: "api"
    taskDefinition: "my-app-api"
    serviceName: "my-app-api-service"
    containerName: "api"

  worker:
    directory: "worker"
    taskDefinition: "my-app-worker"
    serviceName: "my-app-worker-service"
    containerName: "worker"
```

## 🏃‍♂️ Usage

### Local Development

1. **Start the server:**

   ```bash
   go run main.go
   ```

2. **For development with auto-reload:**

   ```bash
   # Install air for live reloading
   go install github.com/cosmtrek/air@latest
   air
   ```

3. **Expose for webhook testing:**
   ```bash
   ngrok http 8080
   ```

### Command Line Options

```bash
# Basic usage
go run main.go

# Enable deployment with custom config
go run main.go -deploy -config=./deploy.yml

# Help
go run main.go -h
```

### GitHub Webhook Setup

1. Go to your repository → Settings → Webhooks
2. Add webhook URL: `https://your-domain.com/webhook`
3. Set Content type: `application/json`
4. Add your webhook secret
5. Select "Push events"
6. Save the webhook

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitHub Repo   │───▶│   Git Runner    │───▶│   AWS ECS       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Docker Build  │
                       └─────────────────┘
```

### Project Structure

```
git_runner/
├── main.go              # Application entry point
├── config/              # Configuration management
│   ├── config.go       # Environment variables
│   └── init.go         # Deployment initialization
├── handler/             # HTTP request handlers
│   └── webhook.go      # GitHub webhook handler
├── runner/              # Job execution logic
│   ├── job.go          # Main job orchestration
│   └── docker.go       # Docker build operations
├── deploy/              # Deployment logic
│   └── deploy.go       # AWS ECS deployment
├── utils/               # Utility functions
│   └── git.go          # Git operations
└── tmp/                 # Temporary build files
```

## 🔄 Workflow

1. **Webhook Reception**: GitHub sends push event to `/webhook` endpoint
2. **Signature Verification**: HMAC-SHA256 signature validation
3. **Repository Cloning**: Downloads specific commit to temporary directory
4. **Service Discovery**: Scans for Dockerfiles in repository
5. **Docker Build**: Builds each discovered service into Docker images
6. **Deployment** (optional): Pushes images to ECR and updates ECS services

## 🐳 Docker Support

The tool automatically discovers and builds multiple Docker services:

- Scans entire repository for `Dockerfile`s
- Builds each service with unique naming
- Supports multi-service architectures
- Generates lowercase-compliant image tags

## ☁️ AWS ECS Deployment

- **ECR Integration**: Automatic image pushing to Amazon ECR
- **Task Definition Updates**: Updates ECS task definitions with new images
- **Rolling Deployments**: Performs zero-downtime deployments
- **Multi-Service Support**: Handles complex applications with multiple services

## 🔒 Security Features

- **Webhook Signature Verification**: Validates GitHub webhook authenticity
- **Environment Variable Security**: Sensitive data stored in environment variables
- **AWS Credential Management**: Secure AWS authentication
- **Constant-time Comparison**: Prevents timing attacks

## 🧪 Testing

### Test Webhook Locally

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{"repository":{"clone_url":"https://github.com/user/repo.git"},"after":"commit-sha"}'
```

### Validate Configuration

```bash
# Test deployment config
go run main.go -deploy -config=./deploy.yml
```

## 📊 Monitoring

- Server logs to stdout
- Build outputs captured and displayed
- Deployment status tracking
- Error handling and reporting

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is open source. See LICENSE file for details.

## 🚨 Production Considerations

- Use proper secrets management (AWS Secrets Manager, etc.)
- Implement proper logging and monitoring
- Set up SSL/TLS for webhook endpoints
- Configure appropriate resource limits
- Use dedicated IAM roles with minimal permissions
- Set up proper backup and disaster recovery

## 📞 Support

For questions or issues, please open a GitHub issue or contact the development team.

---

Built with ❤️ in Go for modern CI/CD workflows.
