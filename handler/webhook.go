package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"git-runner/config"
	"git-runner/runner"
	"io"
	"net/http"
	"strings"
)

type PushPayload struct {
	Repository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	After string `json:"after"`
}

// verifySignature validates the webhook payload against the GitHub signature header
func verifySignature(payload []byte, signatureHeader string, secret string) bool {
	// GitHub signature header format is "sha256=<hex digest>"
	const signaturePrefix = "sha256="
	if !strings.HasPrefix(signatureHeader, signaturePrefix) {
		return false
	}

	// Extract signature hex string from header
	signatureHex := strings.TrimPrefix(signatureHeader, signaturePrefix)

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(signatureHex), []byte(expectedMAC))
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received webhook request")

	// Check event type
	eventType := r.Header.Get("X-GitHub-Event")
	fmt.Println("Event type:", eventType)

	// Handle ping events (sent when webhook is first set up)
	if eventType == "ping" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Pong!")
		return
	}

	// Continue with push event handling
	if eventType != "push" {
		http.Error(w, "unsupported event type", http.StatusBadRequest)
		return
	}

	// Read and store the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	// Verify signature if webhook secret is configured
	if config.WebhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" || !verifySignature(bodyBytes, signature, config.WebhookSecret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var payload PushPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	go runner.RunJob(payload.Repository.CloneURL, payload.After,
		config.DeployEnabled, config.DeployConfigPath)
	w.WriteHeader(http.StatusAccepted)
}
