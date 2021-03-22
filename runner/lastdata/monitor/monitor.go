package monitor

import (
	"encoding/json"
	"net/http"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata/persistence"
	"github.com/figment-networks/indexer-scheduler/runner/lastdata/structures"
)

type Monitor struct {
	store persistence.PDriver
}

func NewMonitor(store persistence.PDriver) *Monitor {
	return &Monitor{store}
}

func (m *Monitor) RegisterHandles(mux *http.ServeMux) {
	mux.HandleFunc("/scheduler/runner/lastdata/listRunning", m.handlerListRunning)
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
	enc := json.NewEncoder(w)
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
		runs = []structures.LatestRecord{}
	}
	enc.Encode(runs)
}

/*
func (c *Monitor) handlerListScheduleFor(w http.ResponseWriter, r *http.Request) {
	s, err := c.ListScheduleFor(r.Context(), "lastdata", "skale", "", 1000)
	if err != nil {
		log.Println("error", err)
	}
	enc := json.NewEncoder(w)

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(s)
}

func (c *Monitor) ListScheduleFor(ctx context.Context, kind, network, taskID string, limit int) ([]structures.LatestRecord, error) {
	l, err := c.pStore.GetRuns(ctx, kind, network, taskID, limit)
	return l, err
}
*/
