package jobworker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
)

type Job struct {
	ID       string   `json:"id"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	Status   string   `json:"status"`
	ExitCode *int     `json:"exit_code,omitempty"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	cmd      *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
}

var (
	jobs   = make(map[string]*Job)
	jobsMu sync.RWMutex
	nextID = 1
)

func Start(command string, args ...string) (string, error) {
	jobsMu.Lock()
	id := strconv.Itoa(nextID)
	nextID++
	jobsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return "", err
	}

	job := &Job{
		ID:      id,
		Command: command,
		Args:    args,
		Status:  "Running",
		cmd:     cmd,
		ctx:     ctx,
		cancel:  cancel,
	}

	jobsMu.Lock()
	jobs[id] = job
	jobsMu.Unlock()

	go func() {
		err := cmd.Wait()
		jobsMu.Lock()
		defer jobsMu.Unlock()

		job.Stdout = stdout.String()
		job.Stderr = stderr.String()

		if err != nil {
			job.Status = "Failed"
		} else {
			job.Status = "Succeeded"
		}

		if exitCode := cmd.ProcessState.ExitCode(); exitCode != -1 {
			job.ExitCode = &exitCode
		}
	}()

	return job.Status, nil
}

func GetOutput(id string) (string, string, error) {
	jobsMu.RLock()
	defer jobsMu.RUnlock()

	job, ok := jobs[id]
	if !ok {
		return "", "", fmt.Errorf("job not found")
	}
	return job.Stdout, job.Stderr, nil
}

func GetStatus(id string) (string, error) {
	jobsMu.RLock()
	defer jobsMu.RUnlock()

	job, ok := jobs[id]
	if !ok {
		return "", fmt.Errorf("job not found")
	}
	return job.Status, nil
}
