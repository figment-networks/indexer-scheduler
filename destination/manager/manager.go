package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/figment-networks/indexer-scheduler/conn"
	"github.com/figment-networks/indexer-scheduler/conn/tray"
	"github.com/figment-networks/indexer-scheduler/structures"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TargetAdder interface {
	Add(t structures.Target)
	Remove(t structures.Target)
}

type StreamState int

const (
	StreamUnknown StreamState = iota
	StreamOnline
	StreamError
	StreamReconnecting
	StreamClosing
	StreamOffline
)

type WorkerNetworkStatic struct {
	Workers map[string]WorkerInfoStatic `json:"workers"`
	All     int                         `json:"all"`
	Active  int                         `json:"active"`
}

type WorkerInfoStatic struct {
	NodeSelfID     string             `json:"node_id"`
	Type           string             `json:"type"`
	ChainID        string             `json:"chain_id"`
	State          StreamState        `json:"state"`
	ConnectionInfo []WorkerConnection `json:"connection"`
	LastCheck      time.Time          `json:"last_check"`
}

type WorkerConnection struct {
	Version   string          `json:"version"`
	Type      string          `json:"type"`
	Addresses []WorkerAddress `json:"addresses"`
}

type WorkerAddress struct {
	IP      net.IP `json:"ip"`
	Address string `json:"address"`
}

type Manager struct {
	logger *zap.Logger
	ta     TargetAdder
	nodes  map[string]WorkerInfoStatic // NodeSelfID
}

func NewManager(logger *zap.Logger, ta TargetAdder) *Manager {
	return &Manager{
		logger: logger,
		ta:     ta,
		nodes:  make(map[string]WorkerInfoStatic),
	}
}

func (m *Manager) Load(ctx context.Context, t structures.Target, ct *tray.ConnTray) {

	rcpconn, err := ct.Get(t.ConnType, t.Address)
	if err != nil {
		m.logger.Error("error getting connection", zap.Error(err))
	}

	sID := uuid.New()
	ch := make(chan conn.Response, 10)
	defer rcpconn.CloseStream(sID.String())
	defer close(ch)

	readr := new(bytes.Reader)
	dec := json.NewDecoder(readr)

	tckr := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tckr.C:
			rcpconn.Send(sID.String(), ch, 0, "get_workers", nil)
		case resp := <-ch:
			if resp.Error != nil {
				m.logger.Error("error getting workers", zap.Error(err))
				continue
			}

			wns := map[string]WorkerNetworkStatic{}
			readr.Reset(resp.Result)
			err := dec.Decode(&wns)
			if resp.Error != nil {
				m.logger.Error("error decoding worker response", zap.Error(err))
				continue
			}

			for network, sub := range wns {
				for _, w := range sub.Workers {
					n, ok := m.nodes[w.NodeSelfID]
					if !ok && w.State == StreamOnline {
						m.nodes[w.NodeSelfID] = w
						for _, ci := range w.ConnectionInfo {
							m.ta.Add(structures.Target{
								Network:          network,
								Version:          ci.Version,
								ChainID:          w.ChainID,
								Address:          t.Address,
								ConnType:         t.ConnType,
								AdditionalConfig: t.AdditionalConfig,
							})
						}
						continue
					}

					if n.State == StreamOnline && w.State != StreamOnline {
						for _, ci := range w.ConnectionInfo {
							m.ta.Remove(structures.Target{
								Network:  network,
								Version:  ci.Version,
								ChainID:  w.ChainID,
								Address:  t.Address,
								ConnType: t.ConnType,
							})
						}
					}

					m.nodes[w.NodeSelfID] = w
				}

				for k, n := range m.nodes {
					if _, ok := sub.Workers[k]; !ok {
						for _, ci := range n.ConnectionInfo {
							m.ta.Remove(structures.Target{
								Network:  network,
								Version:  ci.Version,
								ChainID:  n.ChainID,
								Address:  t.Address,
								ConnType: t.ConnType,
							})
						}
						delete(m.nodes, k)
					}
				}

			}
		}
	}
}
