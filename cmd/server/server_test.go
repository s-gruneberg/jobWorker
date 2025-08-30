package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/s-gruneberg/jobWorker/internal/jobworker"
)

func TestHandleStartJob(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful job start",
			method:         "POST",
			body:           `{"command": "echo", "args": ["hello"]}`,
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"job_id":"1"}`,
		},
		{
			name:           "wrong method",
			method:         "GET",
			body:           `{"command": "echo", "args": ["hello"]}`,
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "invalid JSON",
			method:         "POST",
			body:           `{"command": "echo", "args": ["hello"`,
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
		{
			name:           "missing command",
			method:         "POST",
			body:           `{"args": ["hello"]}`,
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Command is required\n",
		},
		{
			name:           "empty command",
			method:         "POST",
			body:           `{"command": "", "args": ["hello"]}`,
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Command is required\n",
		},
		{
			name:           "missing auth header",
			method:         "POST",
			body:           `{"command": "echo", "args": ["hello"]}`,
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "invalid auth format",
			method:         "POST",
			body:           `{"command": "echo", "args": ["hello"]}`,
			authHeader:     "admin-token-123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "invalid token",
			method:         "POST",
			body:           `{"command": "echo", "args": ["hello"]}`,
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobworker.Clear()

			req, err := http.NewRequest(tt.method, "/jobs/start", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleStartJob)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response StartJobResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.JobID == "" {
					t.Errorf("expected job_id in response, got empty string")
				}
			} else {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}
		})
	}
}
func TestHandleGetOutput(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		jobID          string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful output retrieval",
			method:         "GET",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         "POST",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "empty job ID",
			method:         "GET",
			jobID:          "",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL - empty job ID\n",
		},
		{
			name:           "job not found",
			method:         "GET",
			jobID:          "999",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Job not found: job not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobworker.Clear()

			if tt.jobID == "1" && tt.expectedStatus == http.StatusOK {
				_, err := jobworker.Start("echo", "hello")
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(100 * time.Millisecond)
			}

			url := fmt.Sprintf("/jobs/output/%s", tt.jobID)
			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleGetOutput)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response OutputResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.Stdout == "" && response.Stderr == "" {
					t.Errorf("expected stdout or stderr in response, got empty strings")
				}
			}
		})
	}
}

func TestHandleGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		jobID          string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful status retrieval",
			method:         "GET",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         "POST",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "empty job ID",
			method:         "GET",
			jobID:          "",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL - empty job ID\n",
		},
		{
			name:           "job not found",
			method:         "GET",
			jobID:          "999",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Job not found: job not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobworker.Clear()

			if tt.jobID == "1" && tt.expectedStatus == http.StatusOK {
				_, err := jobworker.Start("echo", "hello")
				if err != nil {
					t.Fatal(err)
				}
			}

			url := fmt.Sprintf("/jobs/status/%s", tt.jobID)
			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleGetStatus)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response StatusResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.Status == "" {
					t.Errorf("expected status in response, got empty string")
				}
			}
		})
	}
}

func TestHandleStopJob(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		jobID          string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful job stop",
			method:         "PUT",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         "GET",
			jobID:          "1",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "empty job ID",
			method:         "PUT",
			jobID:          "",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL - empty job ID\n",
		},
		{
			name:           "job not found",
			method:         "PUT",
			jobID:          "999",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Job not found: job not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobworker.Clear()

			if tt.jobID == "1" && tt.expectedStatus == http.StatusOK {
				_, err := jobworker.Start("sleep", "10")
				if err != nil {
					t.Fatal(err)
				}
			}

			url := fmt.Sprintf("/jobs/stop/%s", tt.jobID)
			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleStopJob)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response StopJobResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.JobID != tt.jobID {
					t.Errorf("expected job_id %s in response, got %s", tt.jobID, response.JobID)
				}
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		action   string
		expected bool
	}{
		{"admin can start", "admin", "start", true},
		{"admin can stop", "admin", "stop", true},
		{"admin can get status", "admin", "status", true},
		{"admin can get output", "admin", "output", true},
		{"operator can start", "operator", "start", true},
		{"operator can get status", "operator", "status", true},
		{"operator can get output", "operator", "output", true},
		{"operator cannot stop", "operator", "stop", false},
		{"viewer can get status", "viewer", "status", true},
		{"viewer can get output", "viewer", "output", true},
		{"viewer cannot start", "viewer", "start", false},
		{"viewer cannot stop", "viewer", "stop", false},
		{"invalid role", "invalid", "start", false},
		{"invalid action", "admin", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthorized(tt.role, tt.action)
			if result != tt.expected {
				t.Errorf("isAuthorized(%s, %s) = %v; want %v", tt.role, tt.action, result, tt.expected)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid admin token for start",
			action:         "start",
			authHeader:     "Bearer admin-token-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid operator token for start",
			action:         "start",
			authHeader:     "Bearer operator-token-456",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "viewer cannot start",
			action:         "start",
			authHeader:     "Bearer viewer-token-789",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "missing auth header",
			action:         "start",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "invalid auth format",
			action:         "start",
			authHeader:     "admin-token-123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "empty token",
			action:         "start",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "invalid token",
			action:         "start",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			middleware := authMiddleware(tt.action)
			wrappedHandler := middleware(handler)

			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("middleware returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("middleware returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}
		})
	}
}
