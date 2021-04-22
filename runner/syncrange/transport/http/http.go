package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/figment-networks/indexer-scheduler/runner/syncrange"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"
)

const ConnectionTypeHTTP = "http"

type AdditionalConfig struct {
	Endpoint string `json:"endpoint"`
}

func setAdditionalConfig(in interface{}) (ac AdditionalConfig) {
	i, ok := in.(map[string]interface{})

	if !ok {
		return ac
	}

	endp, ok := i["endpoint"]
	if ok {
		if s, isstring := endp.(string); isstring {
			ac.Endpoint = s
		}
	}

	return ac
}

type SyncrangeHTTPTransport struct {
	client *http.Client
	l      *zap.Logger
}

func NewSyncrangeHTTPTransport(l *zap.Logger) *SyncrangeHTTPTransport {
	return &SyncrangeHTTPTransport{
		l: l,
		client: &http.Client{
			Timeout: time.Second * 40,
		},
	}
}

func (ld *SyncrangeHTTPTransport) GetLastData(ctx context.Context, t coreStructs.Target, ldReq structures.SyncDataRequest) (ldr structures.SyncDataResponse, backoff bool, err error) {

	var adc AdditionalConfig
	ad, ok := t.AdditionalConfig[syncrange.RunnerName]
	if ok {
		adc = setAdditionalConfig(ad)
	}

	ld.l.Info("Running SyncRange",
		zap.String("network", ldReq.Network),
		zap.String("chain_id", ldReq.ChainID),
		zap.String("address", t.Address),
		zap.Uint64("last_height", ldReq.LastHeight),
		zap.Uint64("retry_count", ldReq.RetryCount),
	)

	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	if err := enc.Encode(&ldReq); err != nil {
		return ldr, false, &coreStructs.RunError{Contents: fmt.Errorf("error encoding request: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.Address+adc.Endpoint, b)
	if err != nil {
		return ldr, false, &coreStructs.RunError{Contents: fmt.Errorf("error creating response: %w", err)}
	}

	resp, err := ld.client.Do(req)
	if err != nil {
		return ldr, true, &coreStructs.RunError{Contents: fmt.Errorf("error getting response:  %w", err)}
	}

	ldrr := &structures.SyncDataResponse{}
	dec := json.NewDecoder(resp.Body)

	defer resp.Body.Close()

	if err = dec.Decode(ldrr); err != nil {
		return *ldrr, false, &coreStructs.RunError{Contents: fmt.Errorf("error decoding response:  %w", err)}
	}

	// Still processing
	if resp.StatusCode == http.StatusProcessing || ldrr.Processing {
		return structures.SyncDataResponse{
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
