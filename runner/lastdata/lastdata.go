package lastdata

import (
	"context"
	"fmt"

	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/runner/lastdata/persistence"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
)

const RunnerName = "lastdata"

type LastDataTransporter interface {
	GetLastData(context.Context, structures.LatestDataRequest) (lastResponse structures.LatestDataResponse, backoff bool, err error)
}

type Client struct {
	store     *persistence.LastDataStorageTransport
	transport LastDataTransporter
}

func NewClient(store *persistence.LastDataStorageTransport, transport LastDataTransporter) *Client {
	return &Client{
		store:     store,
		transport: transport,
	}
}
func (c *Client) Name() string {
	return RunnerName
}

func (c *Client) Run(ctx context.Context, rcp coreStructs.RunConfigParams) (backoff bool, err error) {
	latest, err := c.store.GetLatest(ctx, rcp)
	if err != nil && err != params.ErrNotFound {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error getting data from store GetLatest [%s]:  %w", RunnerName, err)}
	}

	lrec := structures.LatestRecord{
		Hash:       latest.Hash,
		Height:     latest.Height,
		Time:       latest.Time,
		Nonce:      latest.Nonce,
		RetryCount: latest.RetryCount,
	}

	resp, backoff, err := c.transport.GetLastData(ctx, structures.LatestDataRequest{
		Network: rcp.Network,
		ChainID: rcp.ChainID,
		Version: rcp.Version,
		TaskID:  rcp.TaskID,

		LastHeight: latest.Height,
		LastHash:   latest.Hash,
		LastTime:   latest.Time,
		Nonce:      latest.Nonce,
		RetryCount: latest.RetryCount,
	})
	lrec.RetryCount = resp.RetryCount

	if resp.LastHeight > 0 || !(resp.LastTime.IsZero() || resp.LastTime.Unix() == 0) {
		lrec = structures.LatestRecord{
			Hash:       resp.LastHash,
			Height:     resp.LastHeight,
			Time:       resp.LastTime,
			Nonce:      resp.Nonce,
			RetryCount: resp.RetryCount,
			Error:      resp.Error,
		}
	}

	// do not proceed on error
	if len(resp.Error) != 0 {
		lrec.Height = latest.Height
		lrec.Error = resp.Error
		backoff = true
	}

	if err != nil {
		lrec.Error = []byte(err.Error())
		backoff = true
		lrec.RetryCount++
	}

	if err2 := c.store.SetLatest(ctx, rcp, lrec); err2 != nil {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error writing last record SetLatest [%s]:  %w", RunnerName, err2)}
	}

	if err != nil {
		return backoff, &coreStructs.RunError{Contents: fmt.Errorf("error getting data from GetLastData [%s]:  %w", RunnerName, err)}
	}

	return backoff, nil
}
