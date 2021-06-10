package lastdata

import (
	"context"
	"fmt"
	"net/http"

	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/runner/lastdata/monitor"
	"github.com/figment-networks/indexer-scheduler/runner/lastdata/persistence"
	"go.uber.org/zap"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
)

const RunnerName = "lastdata"

type LastDataTransporter interface {
	GetLastData(context.Context, coreStructs.Target, structures.LatestDataRequest) (lastResponse structures.LatestDataResponse, backoff bool, err error)
}

type TargetGetter interface {
	Get(nv coreStructs.NVCKey) (t coreStructs.Target, ok bool)
}

type Client struct {
	store     *persistence.LastDataStorageTransport
	transport map[string]LastDataTransporter
	dest      TargetGetter
	logger    *zap.Logger
	m         *monitor.Monitor
}

func NewClient(logger *zap.Logger, store *persistence.LastDataStorageTransport, ac auth.AuthCredentials, dest TargetGetter) *Client {
	return &Client{
		store:     store,
		dest:      dest,
		logger:    logger,
		transport: make(map[string]LastDataTransporter),
		m:         monitor.NewMonitor(store, ac),
	}
}

func (c *Client) AddTransport(typeS string, tr LastDataTransporter) {
	c.transport[typeS] = tr
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
		LastTime:   latest.LastTime,
		Nonce:      latest.Nonce,
		RetryCount: latest.RetryCount,
	}

	t, ok := c.dest.Get(coreStructs.NVCKey{Network: rcp.Network, Version: rcp.Version, ChainID: rcp.ChainID})
	if !ok {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error getting response:  %w", coreStructs.ErrNoDestinationAvailable)}
	}

	tr, ok := c.transport[t.ConnType]
	if !ok {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("no such transport of lastdata as :  %s", t.ConnType)}
	}

	resp, backoff, err := tr.GetLastData(ctx, t, structures.LatestDataRequest{
		Network: rcp.Network,
		ChainID: rcp.ChainID,
		Version: rcp.Version,
		TaskID:  rcp.TaskID,

		LastHeight: latest.Height,
		LastHash:   latest.Hash,
		LastTime:   latest.LastTime,
		Nonce:      latest.Nonce,
		RetryCount: latest.RetryCount,
	})
	lrec.RetryCount = resp.RetryCount

	if resp.LastHeight > 0 || !(resp.LastTime.IsZero() || resp.LastTime.Unix() == 0) {
		lrec = structures.LatestRecord{
			Hash:       resp.LastHash,
			Height:     resp.LastHeight,
			LastTime:   resp.LastTime,
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
		lrec.RetryCount++
	}

	if err != nil {
		lrec.Error = []byte(err.Error())
		backoff = true
		lrec.RetryCount++
	}

	c.logger.Info("[LastData] Response ",
		zap.String("runner", "lastdata"),
		zap.String("network", rcp.Network),
		zap.String("chain_id", rcp.ChainID),
		zap.String("task_id", rcp.TaskID),
		zap.Uint64("req_last_height", latest.Height),
		zap.Uint64("resp_last_height", resp.LastHeight),
		zap.String("error", string(lrec.Error)),
	)

	if err2 := c.store.SetLatest(ctx, rcp, lrec); err2 != nil {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error writing last record SetLatest [%s]:  %w", RunnerName, err2)}
	}

	if err != nil {
		return backoff, &coreStructs.RunError{Contents: fmt.Errorf("error getting data from GetLastData [%s]:  %w", RunnerName, err)}
	}

	return backoff, nil
}

func (c *Client) RegisterHandles(mux *http.ServeMux) {
	c.m.RegisterHandles(mux)
}
