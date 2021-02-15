package lastdata

import (
	"context"
	"fmt"

	"github.com/figment-networks/indexer-scheduler/persistence"
	"github.com/figment-networks/indexer-scheduler/structures"
)

const RunnerName = "lastdata"

type LastDataTransporter interface {
	GetLastData(context.Context, structures.LatestDataRequest) (lastResponse structures.LatestDataResponse, backoff bool, err error)
}

type Client struct {
	store     *persistence.Storage
	transport LastDataTransporter
}

func NewClient(store *persistence.Storage, transport LastDataTransporter) *Client {
	return &Client{
		store:     store,
		transport: transport,
	}
}
func (c *Client) Name() string {
	return RunnerName
}

func (c *Client) Run(ctx context.Context, network, chainID, taskID, version string) (backoff bool, err error) {
	latest, err := c.store.GetLatest(ctx, RunnerName, network, chainID, taskID, version)
	if err != nil && err != structures.ErrDoesNotExists {
		return false, &structures.RunError{Contents: fmt.Errorf("error getting data from store GetLatest [%s]:  %w", RunnerName, err)}
	}

	lrec := structures.LatestRecord{
		Hash:   latest.Hash,
		Height: latest.Height,
		Time:   latest.Time,
		Nonce:  latest.Nonce,
		Retry:  latest.Retry,
	}

	resp, backoff, err := c.transport.GetLastData(ctx, structures.LatestDataRequest{
		Network: network,
		ChainID: chainID,
		Version: version,
		TaskID:  taskID,

		LastHeight: latest.Height,
		LastHash:   latest.Hash,
		LastTime:   latest.Time,
		Nonce:      latest.Nonce,
		Retry:      latest.Retry,
	})
	lrec.Retry = resp.Retry

	if resp.LastHeight > 0 || !(resp.LastTime.IsZero() || resp.LastTime.Unix() == 0) {
		lrec = structures.LatestRecord{
			Hash:   resp.LastHash,
			Height: resp.LastHeight,
			Time:   resp.LastTime,
			Nonce:  resp.Nonce,
			Retry:  resp.Retry,
			Error:  resp.Error,
		}
	}

	// do not proceed on error
	if len(resp.Error) != 0 {
		lrec.Height = latest.Height
		lrec.Error = resp.Error
	}

	if err != nil {
		lrec.Error = []byte(err.Error())
		lrec.Retry++
	}

	if err2 := c.store.SetLatest(ctx, RunnerName, network, chainID, version, taskID, lrec); err2 != nil {
		return false, &structures.RunError{Contents: fmt.Errorf("error writing last record SetLatest [%s]:  %w", RunnerName, err2)}
	}

	if err != nil {
		return backoff, &structures.RunError{Contents: fmt.Errorf("error getting data from GetLastData [%s]:  %w", RunnerName, err)}
	}

	return backoff, nil
}
