package persistence

import (
	"context"

	"github.com/figment-networks/indexer-scheduler/structures"
	"github.com/google/uuid"
)

type CDriver interface {
	AddConfig(ctx context.Context, rc structures.RunConfig) (err error)
	GetConfigs(ctx context.Context) (rcs []structures.RunConfig, err error)

	MarkRunning(ctx context.Context, runID, configID uuid.UUID) (err error)
	MarkFinished(ctx context.Context, id uuid.UUID) (err error)

	MarkStopped(ctx context.Context, id uuid.UUID) (err error)
	RemoveStatusAllEnabled(ctx context.Context) (err error)
}

type CoreStorage struct {
	Driver CDriver
}

func (cs *CoreStorage) AddConfig(ctx context.Context, rc structures.RunConfig) (err error) {
	return cs.Driver.AddConfig(ctx, rc)
}

func (cs *CoreStorage) GetConfigs(ctx context.Context) (rcs []structures.RunConfig, err error) {
	return cs.Driver.GetConfigs(ctx)
}

func (cs *CoreStorage) MarkRunning(ctx context.Context, runID, configID uuid.UUID) (err error) {
	return cs.Driver.MarkRunning(ctx, runID, configID)
}

func (cs *CoreStorage) MarkStopped(ctx context.Context, id uuid.UUID) (err error) {
	return cs.Driver.MarkStopped(ctx, id)
}

func (cs *CoreStorage) MarkFinished(ctx context.Context, id uuid.UUID) (err error) {
	return cs.Driver.MarkFinished(ctx, id)
}

func (cs *CoreStorage) RemoveStatusAllEnabled(ctx context.Context) (err error) {
	return cs.Driver.RemoveStatusAllEnabled(ctx)
}
