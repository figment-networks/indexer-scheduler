package structures

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDoesNotExists      = errors.New("does not exists")
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
	Duration string `json:"duration"`
	Kind     string `json:"kind"`
	TaskID   string `json:"task_id"`
}

type LatestRecord struct {
	TaskID string    `json:"task_id"`
	Hash   string    `json:"hash"`
	Height uint64    `json:"height"`
	Time   time.Time `json:"time"`
	From   string    `json:"from"`
	Nonce  []byte    `json:"nonce"`
	Retry  uint64    `json:"retry"`
	Error  []byte    `json:"error"`
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

	LastHash   string    `json:"lastHash"`
	LastEpoch  string    `json:"lastEpoch"`
	LastHeight uint64    `json:"lastHeight"`
	LastTime   time.Time `json:"lastTime"`
	Retry      uint64    `json:"retry"`
	Nonce      []byte    `json:"nonce"`

	SelfCheck bool `json:"selfCheck"`
}

type LatestDataResponse struct {
	LastHash   string    `json:"lastHash"`
	LastHeight uint64    `json:"lastHeight"`
	LastTime   time.Time `json:"lastTime"`
	LastEpoch  string    `json:"lastEpoch"`
	Retry      uint64    `json:"retry"`
	Nonce      []byte    `json:"nonce"`
	Error      []byte    `json:"error"`

	Processing bool `json:"processing"`
}
