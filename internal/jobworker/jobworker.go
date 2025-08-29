package jobworker

import (
	"context"
	"os/exec"
	"sync"
)

type Job struct {
	ID       string   `json:"id"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
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
