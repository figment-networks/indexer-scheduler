package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/figment-networks/indexer-scheduler/destination"
	"github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"
)

const ConnectionTypeHTTP = "http"

type LastDataHTTPTransport struct {
	client *http.Client
	dest   *destination.Scheme
	l      *zap.Logger
}

func NewLastDataHTTPTransport(dest *destination.Scheme, l *zap.Logger) *LastDataHTTPTransport {
	return &LastDataHTTPTransport{
		dest: dest,
		l:    l,
		client: &http.Client{
			Timeout: time.Second * 40,
		},
	}
}

func (ld LastDataHTTPTransport) GetLastData(ctx context.Context, ldReq structures.LatestDataRequest) (ldr structures.LatestDataResponse, backoff bool, err error) {

	t, ok := ld.dest.Get(destination.NVCKey{Network: ldReq.Network, Version: ldReq.Version, ChainID: ldReq.ChainID, ConnType: ConnectionTypeHTTP})
	if !ok {
		return ldr, false, &structures.RunError{Contents: fmt.Errorf("error getting response:  %w", structures.ErrNoWorkersAvailable)}
	}

	ld.l.Info("Running LastData",
		zap.String("network", ldReq.Network),
		zap.String("chain_id", ldReq.ChainID),
		zap.String("address", t.Address),
		zap.Uint64("last_height", ldReq.LastHeight),
		zap.Uint64("retry_count", ldReq.RetryCount),
	)

	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	if err := enc.Encode(&ldReq); err != nil {
		return ldr, false, &structures.RunError{Contents: fmt.Errorf("error encoding request: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.Address+"/scrape_latest", b)
	if err != nil {
		return ldr, false, &structures.RunError{Contents: fmt.Errorf("error creating response: %w", err)}
	}

	resp, err := ld.client.Do(req)
	if err != nil {
		return ldr, true, &structures.RunError{Contents: fmt.Errorf("error getting response:  %w", err)}
	}

	ldrr := &structures.LatestDataResponse{}

	dec := json.NewDecoder(resp.Body)

	defer resp.Body.Close()

	if err = dec.Decode(ldrr); err != nil {
		return *ldrr, false, &structures.RunError{Contents: fmt.Errorf("error decoding response:  %w", err)}
	}

	// Still processing
	if resp.StatusCode == http.StatusProcessing || ldrr.Processing {
		return structures.LatestDataResponse{
			LastHash:   ldReq.LastHash,
			LastHeight: ldReq.LastHeight,
			LastTime:   ldReq.LastTime,
			LastEpoch:  ldReq.LastEpoch,
			Nonce:      ldReq.Nonce,
			RetryCount: ldReq.RetryCount + 1,
		}, true, nil
	}

	return *ldrr, false, nil
}
