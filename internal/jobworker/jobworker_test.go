package jobworker

import (
	"testing"
	"time"
)

func TestStartAndStatus(t *testing.T) {
	clear()

	id, err := Start("echo", "hello", "world")
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	if id == "" {
		t.Error("Job ID should not be empty")
	}

	time.Sleep(100 * time.Millisecond)

	status, err := GetStatus(id)
	if err != nil {
		t.Fatalf("Failed to get status %v", err)
	}
	if status != "Succeeded" {
		t.Fatalf("Job did not succeed")
	}

}

func TestStartMultipleJobs(t *testing.T) {
	clear()

	id1, _ := Start("echo", "job1")
	id2, _ := Start("echo", "job2")
	id3, _ := Start("echo", "job3")

	time.Sleep(100 * time.Millisecond)

	// Test getting individual jobs
	_, err := GetStatus(id1)
	if err != nil {
		t.Errorf("Failed to get job1: %v", err)
	}

	_, err = GetStatus(id2)
	if err != nil {
		t.Errorf("Failed to get job2: %v", err)
	}

	_, err = GetStatus(id3)
	if err != nil {
		t.Errorf("Failed to get job3: %v", err)
	}
}

func TestStopJob(t *testing.T) {
	clear()

	id, err := Start("sleep", "10")
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	err = Stop(id)
	if err != nil {
		t.Fatalf("Failed to stop job: %v", err)
	}

	// Check status
	status, err := GetStatus(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if status != "Stopped" {
		t.Errorf("Expected status %s, got %s", "Stopped", status)
	}
}

func TestGetOutput(t *testing.T) {
	clear()

	id, err := Start("echo", "hello world")
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	stdout, stderr, err := GetOutput(id)
	if err != nil {
		t.Fatalf("Failed to get output: %v", err)
	}

	if stdout != "hello world\n" {
		t.Errorf("Expected stdout 'hello world\\n', got '%s'", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got '%s'", stderr)
	}
}
