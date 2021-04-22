package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/figment-networks/indexer-scheduler/conn/tray"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"

	"github.com/figment-networks/indexer-scheduler/conn"
	"go.uber.org/zap"
)

const ConnectionTypeWS = "ws"

type SyncRangeWSTransport struct {
	l      *zap.Logger
	ct     *tray.ConnTray
	nextID uint64
}

func NewSyncRangeWSTransport(l *zap.Logger, ct *tray.ConnTray) *SyncRangeWSTransport {
	return &SyncRangeWSTransport{
		l:  l,
		ct: ct,
	}
}

func (ld *SyncRangeWSTransport) GetLastData(ctx context.Context, t coreStructs.Target, ldReq structures.SyncDataRequest) (ldr structures.SyncDataResponse, backoff bool, err error) {
	ld.l.Info("Running LastData",
		zap.String("network", ldReq.Network),
		zap.String("chain_id", ldReq.ChainID),
		zap.String("address", t.Address),
		zap.String("address", ldReq.TaskID),
		zap.Uint64("last_height", ldReq.LastHeight),
		zap.Uint64("retry_count", ldReq.RetryCount),
	)

	rpc, err := ld.ct.Get(ConnectionTypeWS, t.Address)
	if err != nil {
		return ldr, true, &coreStructs.RunError{Contents: fmt.Errorf("error getting connection:  %w", err)}
	}

	ch := make(chan conn.Response, 1)
	defer close(ch)

	ld.nextID++
	sent := ld.nextID
	rpc.Send(ch, sent, "sync_range", []interface{}{ldReq})
	var resp conn.Response

WAIT_FOR_MESSAGE:
	for {
		select {
		case resp = <-ch:
			if resp.ID == sent {
				if resp.Error != nil {
					return ldr, true, &coreStructs.RunError{Contents: fmt.Errorf("error getting response:  %w", resp.Error)}
				}
				break WAIT_FOR_MESSAGE
			}
			ld.l.Warn("Outstanding message passed", zap.Any("response", resp))
		case <-time.After(time.Minute * 5):
			return structures.SyncDataResponse{
				LastHash:   ldReq.LastHash,
				LastHeight: ldReq.LastHeight,
				LastTime:   ldReq.LastTime,
				LastEpoch:  ldReq.LastEpoch,
				Nonce:      ldReq.Nonce,
				RetryCount: ldReq.RetryCount + 1,
			}, true, &coreStructs.RunError{Contents: fmt.Errorf("error getting response timed out")}
		}
	}

	ldrr := &structures.SyncDataResponse{}
	dec := json.NewDecoder(bytes.NewReader(resp.Result))

	if err = dec.Decode(ldrr); err != nil {
		return *ldrr, false, &coreStructs.RunError{Contents: fmt.Errorf("error decoding response:  %w", err)}
	}
	// Still processing
	if ldrr.Processing {
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
