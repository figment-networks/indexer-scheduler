package persistence

import (
	"context"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
)

type PDriver interface {
	GetLatest(ctx context.Context, rcp coreStructs.RunConfigParams) (structures.LatestRecord, error)
	SetLatest(ctx context.Context, rcp coreStructs.RunConfigParams, latest structures.LatestRecord) error
	GetRuns(ctx context.Context, kind, network, taskID string, limit int) (lRec []structures.LatestRecord, err error)
}

type LastDataStorageTransport struct {
	Driver PDriver
}

func NewLastDataStorageTransport(driver PDriver) *LastDataStorageTransport {
	return &LastDataStorageTransport{
		Driver: driver,
	}
}

func (s *LastDataStorageTransport) GetLatest(ctx context.Context, rcp coreStructs.RunConfigParams) (structures.LatestRecord, error) {
	return s.Driver.GetLatest(ctx, rcp)
}

func (s *LastDataStorageTransport) SetLatest(ctx context.Context, rcp coreStructs.RunConfigParams, latest structures.LatestRecord) error {
	return s.Driver.SetLatest(ctx, rcp, latest)
}

func (s *LastDataStorageTransport) GetRuns(ctx context.Context, kind, network, taskID string, limit int) (lRec []structures.LatestRecord, err error) {
	return s.Driver.GetRuns(ctx, kind, network, taskID, limit)
}
