package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/figment-networks/indexer-scheduler/conn/tray"
	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
	coreStructs "github.com/figment-networks/indexer-scheduler/structures"
	"github.com/google/uuid"

	"github.com/figment-networks/indexer-scheduler/conn"
	"go.uber.org/zap"
)

const ConnectionTypeWS = "ws"

type LastDataWSTransport struct {
	l      *zap.Logger
	ct     *tray.ConnTray
	nextID uint64
}

func NewLastDataWSTransport(l *zap.Logger, ct *tray.ConnTray) *LastDataWSTransport {
	return &LastDataWSTransport{
		l:  l,
		ct: ct,
	}
}

func (ld *LastDataWSTransport) GetLastData(ctx context.Context, t coreStructs.Target, ldReq structures.LatestDataRequest) (ldr structures.LatestDataResponse, backoff bool, err error) {
	ld.l.Info("[LastData][WS] Running LastData",
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

	sID := uuid.New()
	ch := make(chan conn.Response, 1) // todo(lukanus): make it pool
	defer rpc.CloseStream(sID.String())
	defer close(ch)

	ld.nextID++
	sent := ld.nextID
	rpc.Send(sID.String(), ch, sent, "last_data", []interface{}{ldReq})
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
			return structures.LatestDataResponse{
				LastHash:   ldReq.LastHash,
				LastHeight: ldReq.LastHeight,
				LastTime:   ldReq.LastTime,
				LastEpoch:  ldReq.LastEpoch,
				Nonce:      ldReq.Nonce,
				RetryCount: ldReq.RetryCount + 1,
			}, true, &coreStructs.RunError{Contents: fmt.Errorf("error getting response timed out")}
		}
	}

	ldrr := &structures.LatestDataResponse{}
	if len(resp.Result) != 0 {
		dec := json.NewDecoder(bytes.NewReader(resp.Result))

		if err = dec.Decode(ldrr); err != nil {
			return *ldrr, false, &coreStructs.RunError{Contents: fmt.Errorf("error decoding response:  %w", err)}
		}
	}

	// Still processing
	if ldrr.Processing {
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
