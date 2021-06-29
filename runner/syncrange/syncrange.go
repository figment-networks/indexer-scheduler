package syncrange

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/monitor"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/persistence"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"
)

type SyncRangeTransporter interface {
	GetLastData(context.Context, coreStructs.Target, structures.SyncDataRequest) (lastResponse structures.SyncDataResponse, backoff bool, err error)
}

type TargetGetter interface {
	Get(nv coreStructs.NVCKey) (t coreStructs.Target, ok bool)
}

const RunnerName = "syncrange"

type SyncRangeConfig struct {
	HeightFrom uint64 `json:"height_from"`
	HeightTo   uint64 `json:"height_to"`
}

func SyncRangeFromMapInterface(a map[string]interface{}) (src SyncRangeConfig, ok bool) {
	src = SyncRangeConfig{}
	var err error
	if hf, ok := a["height_from"]; ok {
		if hff, ok := hf.(string); ok {
			src.HeightFrom, err = strconv.ParseUint(hff, 10, 64)
			if err != nil {
				return src, false
			}
		} else {
			return src, false
		}
	} else {
		return src, false
	}
	if hf, ok := a["height_to"]; ok {
		if hff, ok := hf.(string); ok {
			src.HeightTo, err = strconv.ParseUint(hff, 10, 64)
			if err != nil {
				return src, false
			}
		} else {
			return src, false
		}
	} else {
		return src, false
	}

	return src, true
}

type Client struct {
	transport map[string]SyncRangeTransporter
	dest      TargetGetter

	store  *persistence.SyncRangeStorageTransport
	logger *zap.Logger
	m      *monitor.Monitor
}

func NewClient(logger *zap.Logger, store *persistence.SyncRangeStorageTransport, creds auth.AuthCredentials, dest TargetGetter) *Client {
	return &Client{
		store:     store,
		dest:      dest,
		transport: make(map[string]SyncRangeTransporter),
		logger:    logger,
		m:         monitor.NewMonitor(store, creds),
	}
}

func (c *Client) AddTransport(typeS string, tr SyncRangeTransporter) {
	c.transport[typeS] = tr
}

func (c *Client) Name() string {
	return RunnerName
}

func (c *Client) RegisterHandles(mux *http.ServeMux) {
	c.m.RegisterHandles(mux)
}

func (c *Client) Run(ctx context.Context, rcp coreStructs.RunConfigParams) (backoff bool, err error) {

	mi, ok := SyncRangeFromMapInterface(rcp.Config)
	if !ok {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error parsing syncrange config:  %+v", rcp.Config)}
	}
	latest, err := c.store.GetLatest(ctx, rcp)
	if err != nil && err != params.ErrNotFound {
		return false, &coreStructs.RunError{Contents: fmt.Errorf("error getting data from store GetLatest [%s]:  %w", RunnerName, err)}
	}

	if latest.Height != 0 && latest.Height >= mi.HeightTo { // finished
		return false, io.EOF
	}

	lrec := structures.SyncRecord{
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

	startHeight := latest.Height
	if latest.Height == 0 {
		startHeight = mi.HeightFrom
	}

	resp, backoff, err := tr.GetLastData(ctx, t, structures.SyncDataRequest{
		Network: rcp.Network,
		ChainID: rcp.ChainID,
		Version: rcp.Version,
		TaskID:  rcp.TaskID,

		LastHeight:  startHeight,
		FinalHeight: mi.HeightTo,

		LastHash:   latest.Hash,
		LastTime:   latest.LastTime,
		Nonce:      latest.Nonce,
		RetryCount: latest.RetryCount,
	})
	lrec.RetryCount = resp.RetryCount

	if resp.LastHeight > 0 || !(resp.LastTime.IsZero() || resp.LastTime.Unix() == 0) {
		lrec = structures.SyncRecord{
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

	c.logger.Info("[SyncData] Response ",
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
