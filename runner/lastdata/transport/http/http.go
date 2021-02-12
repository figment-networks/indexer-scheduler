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
)

type LastDataHTTPTransport struct {
	client *http.Client
	dest   *destination.Scheme
}

func NewLastDataHTTPTransport(dest *destination.Scheme) *LastDataHTTPTransport {
	return &LastDataHTTPTransport{
		dest: dest,
		client: &http.Client{
			Timeout: time.Second * 40,
		},
	}
}

func (ld LastDataHTTPTransport) GetLastData(ctx context.Context, ldReq structures.LatestDataRequest) (ldr structures.LatestDataResponse, err error) {
	t, ok := ld.dest.Get(destination.NVCKey{Network: ldReq.Network, Version: ldReq.Version, ChainID: ldReq.ChainID})
	if !ok {
		return ldr, &structures.RunError{Contents: fmt.Errorf("error getting response:  %w", structures.ErrNoWorkersAvailable)}
	}

	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	if err := enc.Encode(&ldReq); err != nil {
		return ldr, &structures.RunError{Contents: fmt.Errorf("error encoding request: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.Address+"/scrape_latest", b)
	if err != nil {
		return ldr, &structures.RunError{Contents: fmt.Errorf("error creating response: %w", err)}
	}

	resp, err := ld.client.Do(req)
	if err != nil {
		return ldr, &structures.RunError{Contents: fmt.Errorf("error getting response:  %w", err)}
	}

	ldrr := &structures.LatestDataResponse{}

	dec := json.NewDecoder(resp.Body)

	defer resp.Body.Close()

	if err = dec.Decode(ldrr); err != nil {
		return *ldrr, &structures.RunError{Contents: fmt.Errorf("error decoding response:  %w", err)}
	}

	// Still processing
	if resp.StatusCode == http.StatusProcessing || ldrr.Processing {
		return structures.LatestDataResponse{
			LastHash:   ldReq.LastHash,
			LastHeight: ldReq.LastHeight,
			LastTime:   ldReq.LastTime,
			LastEpoch:  ldReq.LastEpoch,
			Nonce:      ldReq.Nonce,
			Retry:      ldReq.Retry + 1,
		}, nil
	}

	return *ldrr, nil
}
