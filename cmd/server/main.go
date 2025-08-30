package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/s-gruneberg/jobWorker/internal/jobworker"
)

type User struct {
	Token string
	Role  string
}

var tokens = map[string]User{
	"admin-token-123":    {Token: "admin-token-123", Role: "admin"},
	"operator-token-456": {Token: "operator-token-456", Role: "operator"},
	"viewer-token-789":   {Token: "viewer-token-789", Role: "viewer"},
}

var rolePermissions = map[string]map[string]bool{
	"admin": {
		"start":  true,
		"stop":   true,
		"status": true,
		"output": true,
	},
	"operator": {
		"start":  true,
		"status": true,
		"output": true,
	},
	"viewer": {
		"status": true,
		"output": true,
	},
}

type StartJobRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type StartJobResponse struct {
	JobID string `json:"job_id"`
}

type StopJobResponse struct {
	JobID string `json:"job_id"`
}

type OutputResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func handleStartJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	token := auth[7:]

	role := ""
	if t, exists := tokens[token]; exists {
		role = t.Role
	}

	if role == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := jobworker.Start(req.Command, req.Args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start job: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StartJobResponse{JobID: id})
}

func handleGetOutput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 { // base/jobs/output/id
		http.Error(w, "Invalid URL - missing job ID", http.StatusBadRequest)
		return
	}
	// base/jobs/output/id
	id := pathParts[3]
	if id == "" {
		http.Error(w, "Invalid URL - empty job ID", http.StatusBadRequest)
		return
	}

	stdout, stderr, err := jobworker.GetOutput(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Job not found: %v", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(OutputResponse{Stdout: stdout, Stderr: stderr})
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 { // base/jobs/output/id
		http.Error(w, "Invalid URL - missing job ID", http.StatusBadRequest)
		return
	}
	// base/jobs/output/id
	id := pathParts[3]
	if id == "" {
		http.Error(w, "Invalid URL - empty job ID", http.StatusBadRequest)
		return
	}
	status, err := jobworker.GetStatus(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Job not found: %v", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatusResponse{Status: status})

}

func handleStopJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 { // base/jobs/output/id
		http.Error(w, "Invalid URL - missing job ID", http.StatusBadRequest)
		return
	}
	// base/jobs/output/id
	id := pathParts[3]
	if id == "" {
		http.Error(w, "Invalid URL - empty job ID", http.StatusBadRequest)
		return
	}
	err := jobworker.Stop(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Job not found: %v", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StopJobResponse{JobID: id})
}

func isAuthorized(role, action string) bool {
	actions, ok := rolePermissions[role]
	if !ok {
		return false
	}
	return actions[action]
}

func authMiddleware(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := auth[7:]
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if _, exists := tokens[token]; !exists {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			role := ""
			if t, exists := tokens[token]; exists {
				role = t.Role
			}

			if role == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if !isAuthorized(role, action) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/jobs/start", authMiddleware("start")(http.HandlerFunc(handleStartJob)))
	mux.Handle("/jobs/stop/", authMiddleware("stop")(http.HandlerFunc(handleStopJob)))
	mux.Handle("/jobs/status/", authMiddleware("status")(http.HandlerFunc(handleGetStatus)))
	mux.Handle("/jobs/output/", authMiddleware("output")(http.HandlerFunc(handleGetOutput)))

	if _, err := os.Stat("cert.pem"); err != nil {
		panic("cert.pem not found - HTTPS is required")
	}
	if _, err := os.Stat("key.pem"); err != nil {
		panic("key.pem not found - HTTPS is required")
	}

	fmt.Println("Server starting on https://localhost:8080")
	if err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", mux); err != nil {
		panic(err)
	}
}
