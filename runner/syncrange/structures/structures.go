package structures

import (
	"time"
)

type SyncRecord struct {
	TaskID     string    `json:"task_id"`
	Time       time.Time `json:"time"`
	Hash       string    `json:"hash"`
	Height     uint64    `json:"height"`
	LastTime   time.Time `json:"last_time"`
	Nonce      []byte    `json:"nonce"`
	RetryCount uint64    `json:"retry_count"`
	Error      []byte    `json:"error"`
}

type SyncDataRequest struct {
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

type SyncDataResponse struct {
	LastHash   string    `json:"last_hash"`
	LastHeight uint64    `json:"last_height"`
	LastTime   time.Time `json:"last_time"`
	LastEpoch  string    `json:"last_epoch"`
	RetryCount uint64    `json:"retry_count"`
	Nonce      []byte    `json:"nonce"`
	Error      []byte    `json:"error"`

	Processing bool `json:"processing"`
}
