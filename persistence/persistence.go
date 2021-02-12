package persistence

import (
	"context"

	"github.com/figment-networks/indexer-scheduler/structures"
)

type PDriver interface {
	GetLatest(ctx context.Context, kind, network, chainID, taskID, version string) (structures.LatestRecord, error)
	SetLatest(ctx context.Context, kind, network, chainID, taskID, version string, latest structures.LatestRecord) error
	GetRuns(ctx context.Context, kind, network, taskID string, limit int) (lRec []structures.LatestRecord, err error)
}

type Storage struct {
	Driver PDriver
}

func (s *Storage) GetLatest(ctx context.Context, kind, network, chainID, taskID, version string) (structures.LatestRecord, error) {
	return s.Driver.GetLatest(ctx, kind, network, chainID, taskID, version)
}

func (s *Storage) SetLatest(ctx context.Context, kind, network, chainID, taskID, version string, latest structures.LatestRecord) error {
	return s.Driver.SetLatest(ctx, kind, network, chainID, version, taskID, latest)
}

func (s *Storage) GetRuns(ctx context.Context, kind, network, taskID string, limit int) (lRec []structures.LatestRecord, err error) {
	return s.Driver.GetRuns(ctx, kind, network, taskID, limit)
}
