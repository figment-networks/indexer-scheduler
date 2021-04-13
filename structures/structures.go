package structures

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type State string

var (
	StateAdded    State = "added"
	StateFinished State = "finished"
	StateStopped  State = "stopped"
	StateRunning  State = "running"
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

	Enabled bool                   `json:"enabled"`
	Status  State                  `json:"status"`
	Config  map[string]interface{} `json:"config"`
}

type RunConfigParams struct {
	Network  string `json:"network"`
	ChainID  string `json:"chain_id"`
	Interval string `json:"interval"`
	Kind     string `json:"kind"`
	TaskID   string `json:"task_id"`

	Config map[string]interface{} `json:"config"`

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

type Target struct {
	ChainID          string                 `json:"chain_id"`
	Network          string                 `json:"network"`
	Version          string                 `json:"version"`
	Address          string                 `json:"address"`
	ConnType         string                 `json:"conn_type"`
	AdditionalConfig map[string]interface{} `json:"additional"`
}

type TargetConfig struct {
	Target
	Type string `json:"type"`
}

type NVCKey struct {
	Network string
	Version string
	ChainID string
}

func (nv NVCKey) String() string {
	return fmt.Sprintf("%s:%s (%s) %s", nv.Network, nv.ChainID, nv.Version)
}
