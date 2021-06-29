package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/persistence"
	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/process"
	"github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type MonitoredRunner interface {
	process.Runner
	RegisterHandles(mux *http.ServeMux)
}

type Status string

var (
	ErrAlreadyEnabled  = errors.New("this schedule is already enabled")
	ErrAlreadyDisabled = errors.New("this schedule is already disabled")
)

type Core struct {
	ID      uuid.UUID
	run     map[uuid.UUID]structures.RunConfig
	runLock sync.RWMutex

	runners map[string]MonitoredRunner

	logger *zap.Logger

	coreStore *persistence.CoreStorage
	scheduler *process.Scheduler

	creds auth.AuthCredentials
}

func NewCore(store *persistence.CoreStorage, scheduler *process.Scheduler, creds auth.AuthCredentials, logger *zap.Logger) *Core {
	u, _ := uuid.NewRandom()
	return &Core{
		ID:        u,
		coreStore: store,
		scheduler: scheduler,
		logger:    logger,

		run:     map[uuid.UUID]structures.RunConfig{},
		runners: map[string]MonitoredRunner{},
		creds:   creds,
	}
}

func (c *Core) LoadRunner(name string, runner MonitoredRunner) {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	c.runners[name] = runner
}

func (c *Core) AddSchedules(ctx context.Context, rcs []structures.RunConfig) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()

	for _, r := range rcs {
		if r.Kind != "" && r.Network != "" && r.ChainID != "" && r.TaskID != "" {
			c.logger.Info("[Scheduler] Adding schedule config",
				zap.String("kind", r.Kind),
				zap.String("network", r.Network),
				zap.String("chain", r.ChainID),
				zap.String("task_id", r.TaskID),
			)
			r.RunID = c.ID
			err := c.coreStore.AddConfig(ctx, r)
			if err != nil && !errors.Is(err, params.ErrAlreadyRegistred) {
				return fmt.Errorf("Add config errored: %w", err)
			}
		}
	}

	return nil
}

func (c *Core) InitialLoad(ctx context.Context) error {
	return c.coreStore.RemoveStatusAllEnabled(ctx)
}

func (c *Core) LoadScheduler(ctx context.Context) error {
	defer c.logger.Sync()

	rcs, err := c.coreStore.GetConfigs(ctx)

	if err != nil {
		return err
	}
	for _, s := range rcs {
		c.runLock.Lock()
		r, ok := c.run[s.ID]
		if !ok {
			c.run[s.ID] = s
		}
		c.runLock.Unlock()

		if r.Enabled && (r.Status != structures.StateFinished && r.Status != structures.StateStopped) {
			if err := c.EnableSchedule(ctx, s.ID); err != nil {
				return fmt.Errorf("error running enableSchedule %w", err)
			}
		}

	}

	return nil
}

func (c *Core) ListSchedule(ctx context.Context) ([]structures.RunConfig, error) {
	rcs, err := c.coreStore.GetConfigs(ctx)
	return rcs, err
}

func (c *Core) EnableSchedule(ctx context.Context, sID uuid.UUID) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()

	rcs, err := c.coreStore.GetConfigs(ctx)
	if err != nil {
		return fmt.Errorf("error getting config %w", err)
	}
	var r structures.RunConfig
	for _, rconf := range rcs {
		if rconf.ID == sID {
			r = rconf
		}
	}

	if r.Network == "" {
		return fmt.Errorf("there is no such schedule ('%s') to enable", sID)
	}

	if r.Enabled && r.Status == structures.StateRunning {
		return nil
		// return ErrAlreadyEnabled
	}

	runner, ok := c.runners[r.Kind]
	if !ok {
		return fmt.Errorf("there is no such runner: %s", r.Kind)
	}

	c.logger.Info(fmt.Sprintf("[Core] Running schedule %s (%s:%s) %s in %s", runner.Name(), r.Network, r.ChainID, r.Version, r.Duration.String()))
	go c.scheduler.Run(context.Background(), sID, r.Duration,
		structures.RunConfigParams{
			Network: r.Network,
			ChainID: r.ChainID,
			TaskID:  r.TaskID,
			Version: r.Version,
			Config:  r.Config,
			Kind:    r.Kind},
		runner)

	if err := c.coreStore.MarkRunning(ctx, c.ID, sID); err != nil {
		c.logger.Error("[Core] Error setting state running", zap.Error(err))
	}

	r.Enabled = true
	r.Status = structures.StateRunning
	c.run[sID] = r

	return nil
}

func (c *Core) DisableSchedule(ctx context.Context, sID uuid.UUID) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()

	rcs, err := c.coreStore.GetConfigs(ctx)
	if err != nil {
		return fmt.Errorf("error getting config %w", err)
	}
	var r structures.RunConfig
	for _, rconf := range rcs {
		if rconf.ID == sID {
			r = rconf
		}
	}

	if r.Network == "" {
		return fmt.Errorf("there is no such schedule ('%s') to enable", sID)
	}

	if !r.Enabled {
		return ErrAlreadyDisabled
	}

	c.scheduler.Stop(ctx, sID)

	if err := c.coreStore.MarkStopped(ctx, sID); err != nil {
		c.logger.Error("[Core] Error setting state stopped", zap.Error(err))
	}

	r.Enabled = false
	r.Status = structures.StateStopped
	c.run[sID] = r

	return nil

}

func (c *Core) RegisterHandles(smux *http.ServeMux) {
	smux.HandleFunc("/scheduler/core/list", c.handlerListSchedule)
	smux.HandleFunc("/scheduler/core/enable/", c.handlerEnableSchedule)
	smux.HandleFunc("/scheduler/core/disable/", c.handlerDisableSchedule)
	smux.HandleFunc("/scheduler/core/addTask/", c.handlerAddSchedule)
}

func (c *Core) handlerListSchedule(w http.ResponseWriter, r *http.Request) {
	if err := auth.BasicAuth(c.creds, w, r); err != nil {
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	schedule, err := c.ListSchedule(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	enc.Encode(schedule)
}

func (c *Core) handlerEnableSchedule(w http.ResponseWriter, r *http.Request) {
	if err := auth.BasicAuth(c.creds, w, r); err != nil {
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	enc := json.NewEncoder(w)
	sIDs := strings.Replace(r.URL.Path, "/scheduler/core/enable/", "", -1)
	sID, err := uuid.Parse(sIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}

	if err := c.EnableSchedule(r.Context(), sID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	enc.Encode(string(`{"status":"ok"}`))
}

func (c *Core) handlerDisableSchedule(w http.ResponseWriter, r *http.Request) {
	if err := auth.BasicAuth(c.creds, w, r); err != nil {
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	sIDs := strings.Replace(r.URL.Path, "/scheduler/core/disable/", "", -1)
	sID, err := uuid.Parse(sIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}
	if err := c.DisableSchedule(r.Context(), sID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	enc.Encode(string(`{"status":"ok"}`))
}

type RunConfigAddRequest struct {
	Network  string `json:"network"`
	ChainID  string `json:"chain_id"`
	TaskID   string `json:"task_id"`
	Interval string `json:"interval"`
	Kind     string `json:"kind"`

	Config map[string]interface{} `json:"config"`
}

func (c *Core) handlerAddSchedule(w http.ResponseWriter, r *http.Request) {
	if err := auth.BasicAuth(c.creds, w, r); err != nil {
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	rcar := RunConfigAddRequest{}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&rcar); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}

	if rcar.Network == "" || rcar.ChainID == "" ||
		rcar.TaskID == "" || rcar.Interval == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(string(`{"error": "all parameters are required"}`))
		return

	}

	interval, err := time.ParseDuration(rcar.Interval)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}

	runConfig := structures.RunConfig{
		Network:  rcar.Network,
		ChainID:  rcar.ChainID,
		Version:  "0.0.1",
		TaskID:   rcar.TaskID,
		Duration: interval,
		Kind:     rcar.Kind,
		Enabled:  false,
		Config:   rcar.Config,
		Status:   structures.StateAdded,
	}

	if err := c.coreStore.AddConfig(r.Context(), runConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(string(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	enc.Encode(string(`{"status":"ok"}`))
}
