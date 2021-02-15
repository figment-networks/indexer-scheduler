package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Target struct {
	ChainID  string `json:"chain_id"`
	Network  string `json:"network"`
	Version  string `json:"version"`
	Address  string `json:"address"`
	ConnType string `json:"conn_type"`
}

type NVCKey struct {
	Network  string
	Version  string
	ChainID  string
	ConnType string
}

func (nv NVCKey) String() string {
	return fmt.Sprintf("%s:%s (%s) %s", nv.Network, nv.ChainID, nv.Version, nv.ConnType)
}

type Scheme struct {
	destinations    map[NVCKey][]Target
	destinationLock sync.RWMutex

	logger *zap.Logger
}

type WorkerNetworkStatic struct {
	Workers map[string]WorkerInfoStatic `json:"workers"`
	All     int                         `json:"all"`
	Active  int                         `json:"active"`
}

type WorkerInfoStatic struct {
	NodeSelfID     string             `json:"node_id"`
	Type           string             `json:"type"`
	State          int64              `json:"state"`
	ConnectionInfo []WorkerConnection `json:"connection"`
	LastCheck      time.Time          `json:"last_check"`
}

type WorkerConnection struct {
	Version string `json:"version"`
	ChainID string `json:"chain_id"`
	Type    string `json:"type"`
}

func NewScheme(logger *zap.Logger) *Scheme {
	return &Scheme{
		logger:       logger,
		destinations: make(map[NVCKey][]Target),
	}
}

func (s *Scheme) Add(t Target) {
	s.destinationLock.Lock()
	defer s.destinationLock.Unlock()

	s.logger.Info("[Scheduler] Adding destination config", zap.String("connection_tpe", t.ConnType), zap.String("network", t.Network), zap.String("chain", t.ChainID))

	i, ok := s.destinations[NVCKey{t.Network, t.Version, t.ChainID, t.ConnType}]
	if !ok {
		i = []Target{}
	}
	i = append(i, t)

	s.destinations[NVCKey{t.Network, t.Version, t.ChainID, t.ConnType}] = i
}

func (s *Scheme) Get(nv NVCKey) (t Target, ok bool) {
	s.destinationLock.RLock()
	defer s.destinationLock.RUnlock()

	d, ok := s.destinations[nv]
	if !ok {
		return t, false
	}
	return d[0], ok
}

func (s *Scheme) Refresh(ctx context.Context) error {

	s.destinationLock.Lock()
	defer s.destinationLock.Unlock()

	return nil
}

type schemeOutp struct {
	Destinations map[string][]Target `json:"destinations"`
}

func (s *Scheme) handlerListDestination(w http.ResponseWriter, r *http.Request) {
	s.destinationLock.RLock()
	defer s.destinationLock.RUnlock()

	enc := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	so := schemeOutp{
		Destinations: make(map[string][]Target),
	}

	for k, v := range s.destinations {
		so.Destinations[k.String()] = v
	}
	if err := enc.Encode(so); err != nil {
		s.logger.Error("[Scheme] Error encoding data http ", zap.Error(err))
	}

}

func (s *Scheme) RegisterHandles(smux *http.ServeMux) {
	smux.HandleFunc("/scheduler/destination/list", s.handlerListDestination)
}
