package persistence

import (
	"context"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
)

type PDriver interface {
	GetLatest(ctx context.Context, rcp coreStructs.RunConfigParams) (structures.LatestRecord, error)
	SetLatest(ctx context.Context, rcp coreStructs.RunConfigParams, latest structures.LatestRecord) error
	GetRuns(ctx context.Context, kind, network, chainID, taskID string, limit, offset uint64) (lRec []structures.LatestRecord, err error)
}

type SyncRangeStorageTransport struct {
	Driver PDriver
}

func NewLastDataStorageTransport(driver PDriver) *SyncRangeStorageTransport {
	return &SyncRangeStorageTransport{
		Driver: driver,
	}
}

func (s *SyncRangeStorageTransport) GetLatest(ctx context.Context, rcp coreStructs.RunConfigParams) (structures.LatestRecord, error) {
	return s.Driver.GetLatest(ctx, rcp)
}

func (s *SyncRangeStorageTransport) SetLatest(ctx context.Context, rcp coreStructs.RunConfigParams, latest structures.LatestRecord) error {
	return s.Driver.SetLatest(ctx, rcp, latest)
}

func (s *SyncRangeStorageTransport) GetRuns(ctx context.Context, kind, network, chainID, taskID string, limit, offset uint64) (lRec []structures.LatestRecord, err error) {
	return s.Driver.GetRuns(ctx, kind, network, chainID, taskID, limit, offset)
}
