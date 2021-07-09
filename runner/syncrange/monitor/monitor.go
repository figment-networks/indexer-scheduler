package monitor

import (
	"encoding/json"
	"net/http"

	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/persistence"
	"github.com/figment-networks/indexer-scheduler/runner/syncrange/structures"
	"github.com/figment-networks/indexer-scheduler/utils"
)

type Monitor struct {
	store persistence.PDriver
	creds auth.AuthCredentials
}

func NewMonitor(store persistence.PDriver, creds auth.AuthCredentials) *Monitor {
	return &Monitor{store, creds}
}

func (m *Monitor) RegisterHandles(mux *http.ServeMux) {
	mux.HandleFunc("/scheduler/runner/syncrange/listRunning", m.handlerListRunning)
}

type ListRunningRequestPayload struct {
	Kind    string `json:"kind"`
	Network string `json:"network"`
	TaskID  string `json:"task_id"`
	ChainID string `json:"chain_id"`
	Limit   uint64 `json:"limit"`
	Offset  uint64 `json:"offset"`
}

func (m *Monitor) handlerListRunning(w http.ResponseWriter, r *http.Request) {

	if err := auth.BasicAuth(m.creds, w, r); err != nil {
		return
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	enc := json.NewEncoder(w)
	utils.SetupResponse(&w, r)
	w.Header().Add("Content-type", "application/json")

	dec := json.NewDecoder(r.Body)
	lrrp := ListRunningRequestPayload{}

	if err := dec.Decode(&lrrp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(`{"error": "error decoding payload"}`)
	}

	runs, err := m.store.GetRuns(r.Context(), lrrp.Kind, lrrp.Network, lrrp.ChainID, lrrp.TaskID, lrrp.Limit, lrrp.Offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(`{"error": "error getting runs"}`)
	}

	w.WriteHeader(http.StatusOK)

	if runs == nil {
		runs = []structures.SyncRecord{}
	}
	enc.Encode(runs)
}
