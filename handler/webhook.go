package handler

import (
	"encoding/json"
	"git-runner/runner"
	"net/http"
)

type PushPayload struct {
	Repository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	After string `json:"after"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload PushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	go runner.RunJob(payload.Repository.CloneURL, payload.After)
	w.WriteHeader(http.StatusAccepted)
}
