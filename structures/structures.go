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
type LatestRecord struct {
	TaskID     string    `json:"task_id"`
	Hash       string    `json:"hash"`
	Height     uint64    `json:"height"`
	Time       time.Time `json:"time"`
	Nonce      []byte    `json:"nonce"`
	RetryCount uint64    `json:"retry_count"`
	Error      []byte    `json:"error"`
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

type LatestDataRequest struct {
	Network string `json:"network"`
	ChainID string `json:"chain_id"`
	Version string `json:"version"`
	TaskID  string `json:"task_id"`

	LastHash   string    `json:"last_hash"`
	LastEpoch  string    `json:"last_epoch"`
	LastHeight uint64    `json:"last_height"`
	LastTime   time.Time `json:"last_time"`
	RetryCount uint64    `json:"retry_count"`
	Nonce      []byte    `json:"nonce"`

	SelfCheck bool `json:"selfCheck"`
}

type LatestDataResponse struct {
	LastHash   string    `json:"last_hash"`
	LastHeight uint64    `json:"last_height"`
	LastTime   time.Time `json:"last_time"`
	LastEpoch  string    `json:"last_epoch"`
	RetryCount uint64    `json:"retry_count"`
	Nonce      []byte    `json:"nonce"`
	Error      []byte    `json:"error"`

	Processing bool `json:"processing"`
}
