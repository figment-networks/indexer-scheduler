package destination

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"
)

type Targets struct {
	l sync.RWMutex
	T []structures.Target

	next  int
	nextL sync.Mutex
	Len   int
}

func (t *Targets) inc() int {
	t.nextL.Lock()
	defer t.nextL.Unlock()
	if t.next == t.Len-1 {
		t.next = 0
	} else {
		t.next++
	}

	return t.next
}

func (trgs *Targets) Add(t structures.Target) bool {
	trgs.l.Lock()
	defer trgs.l.Unlock()
	for _, v := range trgs.T {
		if v.Address == t.Address {
			return false
		}
	}

	trgs.T = append(trgs.T, t)
	trgs.Len = len(trgs.T)
	return true
}

func (trgs *Targets) Count() uint64 {
	return uint64(trgs.Len)
}

func (trgs *Targets) Remove(t structures.Target) {
	trgs.l.Lock()
	defer trgs.l.Unlock()

	var nT []structures.Target
	for _, lt := range trgs.T {
		if lt.Address != t.Address {
			nT = append(nT, lt)
		}
	}
	trgs.T = nT
	trgs.Len = len(trgs.T)
}

func (trgs *Targets) GetNext() (t structures.Target) {
	trgs.l.RLock()
	defer trgs.l.RUnlock()

	return trgs.T[trgs.inc()]
}

type Scheme struct {
	targets    map[structures.NVCKey]*Targets
	targetLock sync.RWMutex

	creds  auth.AuthCredentials
	logger *zap.Logger
}

func NewScheme(logger *zap.Logger, creds auth.AuthCredentials) *Scheme {
	return &Scheme{
		logger:  logger,
		creds:   creds,
		targets: make(map[structures.NVCKey]*Targets),
	}
}

func (s *Scheme) Add(t structures.Target) {
	s.targetLock.Lock()
	defer s.targetLock.Unlock()

	i, ok := s.targets[structures.NVCKey{t.Network, t.Version, t.ChainID}]
	if !ok {
		i = &Targets{}
	}

	if added := i.Add(t); added {
		s.logger.Info("[Scheduler] Adding destination config", zap.String("connection_type", t.ConnType), zap.String("network", t.Network), zap.String("chain_id", t.ChainID))
		s.targets[structures.NVCKey{t.Network, t.Version, t.ChainID}] = i
	}
}

func (s *Scheme) Get(nv structures.NVCKey) (t structures.Target, ok bool) {
	s.targetLock.RLock()
	defer s.targetLock.RUnlock()

	d, ok := s.targets[nv]
	if !ok {
		return t, false
	}
	return d.GetNext(), ok
}

func (s *Scheme) Remove(t structures.Target) {
	s.targetLock.Lock()
	defer s.targetLock.Unlock()

	key := structures.NVCKey{t.Network, t.Version, t.ChainID}
	targ, ok := s.targets[key]
	if !ok {
		return
	}
	targ.Remove(t)

	if targ.Count() == 0 {
		delete(s.targets, key)
	}
}

type schemeOutp struct {
	Destinations map[string][]structures.Target `json:"destinations"`
}

func (s *Scheme) handlerListDestination(w http.ResponseWriter, r *http.Request) {
	if err := auth.BasicAuth(s.creds, w, r); err != nil {
		return
	}

	s.targetLock.RLock()
	defer s.targetLock.RUnlock()

	enc := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	so := schemeOutp{Destinations: make(map[string][]structures.Target)}

	for k, v := range s.targets {
		so.Destinations[k.Network+":"+k.ChainID+":"+k.Version] = v.T
	}
	if err := enc.Encode(so); err != nil {
		s.logger.Error("[Scheme] Error encoding data http ", zap.Error(err))
	}
}

func (s *Scheme) RegisterHandles(smux *http.ServeMux) {
	smux.HandleFunc("/scheduler/destination/list", s.handlerListDestination)
}
