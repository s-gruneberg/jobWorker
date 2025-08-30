package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			// Clear any existing jobs before each test
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

			// For successful requests, check if job_id is present
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
