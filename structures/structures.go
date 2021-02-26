package structures

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoWorkersAvailable = errors.New("no workers available")
)

type RunConfig struct {
	ID uuid.UUID `json:"id"`

	RunID   uuid.UUID `json:"run_id"`
	Network string    `json:"network"`
	ChainID string    `json:"chain_id"`
	Version string    `json:"version"`

	TaskID string `json:"task_id"`

	Duration time.Duration `json:"duration"`
	Kind     string        `json:"kind"`
}

type RunConfigParams struct {
	Network  string `json:"network"`
	ChainID  string `json:"chain_id"`
	Interval string `json:"interval"`
	Kind     string `json:"kind"`
	TaskID   string `json:"task_id"`

	Version string `json:"version"`
}

type RunError struct {
	Contents      error
	Unrecoverable bool
}

func (re *RunError) Error() string {
	return fmt.Sprintf("error in runner: %s , unrecoverable: %t", re.Contents.Error(), re.Unrecoverable)
}

func (re *RunError) IsRecoverable() bool {
	return !re.Unrecoverable
}
