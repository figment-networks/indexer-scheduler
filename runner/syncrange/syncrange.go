package syncrange

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/monitor"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/persistence"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
)

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

type SyncRangeTransporter interface {
	GetLastData(context.Context, structures.SyncDataRequest) (lastResponse structures.SyncDataResponse, backoff bool, err error)
}

type Client struct {
	store     *persistence.SyncRangeStorageTransport
	transport SyncRangeTransporter
	m         *monitor.Monitor
}

func NewClient(store *persistence.SyncRangeStorageTransport, transport SyncRangeTransporter) *Client {
	return &Client{
		store:     store,
		transport: transport,
		m:         monitor.NewMonitor(store),
	}
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

	resp, backoff, err := c.transport.GetLastData(ctx, structures.SyncDataRequest{
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
