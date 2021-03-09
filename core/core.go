package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

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

	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
	StatusFinished Status = "finished"
	StatusChanged  Status = "changed"
)

type RunInfo struct {
	structures.RunConfig

	Status Status             `json:"status"`
	CFunc  context.CancelFunc `json:"-"`
}

type Core struct {
	ID      uuid.UUID
	run     map[uuid.UUID]*RunInfo
	runLock sync.RWMutex

	runners map[string]MonitoredRunner

	logger *zap.Logger

	coreStore *persistence.CoreStorage
	scheduler *process.Scheduler
}

func NewCore(store *persistence.CoreStorage, scheduler *process.Scheduler, logger *zap.Logger) *Core {
	u, _ := uuid.NewRandom()
	return &Core{
		ID:        u,
		coreStore: store,
		scheduler: scheduler,
		logger:    logger,

		run:     map[uuid.UUID]*RunInfo{},
		runners: map[string]MonitoredRunner{},
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
				return fmt.Errorf("Add Config errored: %w", err)
			}
		}
	}

	return nil
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
			r = &RunInfo{
				RunConfig: s,
			}
			c.run[s.ID] = r
		}
		defer c.runLock.Unlock()

		if r.Status == StatusEnabled || r.Status == StatusFinished {
			continue
		}

		if err := c.EnableSchedule(ctx, s.ID); err != nil {
			return fmt.Errorf("error running enableSchedule %w", err)
		}
	}

	return nil
}

func (c *Core) ListSchedule() []RunInfo {
	c.runLock.RLock()
	defer c.runLock.RUnlock()

	m := make([]RunInfo, len(c.run))
	for _, v := range c.run {
		m = append(m, *v)
	}

	return m
}

func (c *Core) EnableSchedule(ctx context.Context, sID uuid.UUID) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()

	r, ok := c.run[sID]
	if !ok {
		return errors.New("there is no such schedule to enable")
	}

	if r.Status == StatusEnabled {
		return ErrAlreadyEnabled
	}

	runner, ok := c.runners[r.Kind]
	if !ok {
		return fmt.Errorf("there is no such runner: %s", r.Kind)
	}

	c.logger.Info(fmt.Sprintf("[Core] Running schedule %s (%s:%s) %s in %s", runner.Name(), r.Network, r.ChainID, r.Version, r.Duration.String()))
	go c.scheduler.Run(ctx, sID.String(), r.Duration, structures.RunConfigParams{Network: r.Network, ChainID: r.ChainID, TaskID: r.TaskID, Version: r.Version, Kind: r.Kind}, runner)
	err := c.coreStore.MarkRunning(ctx, c.ID, sID)
	if err != nil {
		c.logger.Error("[Core] Error setting state running", zap.Error(err))
	}

	r.Status = StatusEnabled
	c.run[sID] = r

	return nil
}

func (c *Core) DisableSchedule(ctx context.Context, sID uuid.UUID) error {
	c.runLock.Lock()
	defer c.runLock.Unlock()

	r, ok := c.run[sID]
	if !ok {
		return errors.New("there is no such schedule to disable")
	}

	if r.Status == StatusDisabled || r.Status == StatusFinished {
		return ErrAlreadyDisabled
	}

	c.scheduler.Stop(ctx, sID.String())
	err := c.coreStore.MarkStopped(ctx, sID)
	if err != nil {
		c.logger.Error("[Core] Error setting state running", zap.Error(err))
	}

	r.Status = StatusEnabled
	c.run[sID] = r

	return nil

}

func (c *Core) RegisterHandles(smux *http.ServeMux) {
	smux.HandleFunc("/scheduler/core/list", c.handlerListSchedule)
	smux.HandleFunc("/scheduler/core/enable/", c.handlerEnableSchedule)
	smux.HandleFunc("/scheduler/core/disable/", c.handlerDisableSchedule)
}

func (c *Core) handlerListSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := c.ListSchedule()
	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(schedule)
}

func (c *Core) handlerEnableSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")

	enc := json.NewEncoder(w)
	sIDs := strings.Replace(r.URL.Path, "/scheduler/core/enable/", "", -1)
	sID, err := uuid.Parse(sIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	if err := c.EnableSchedule(r.Context(), sID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	enc.Encode([]byte(`{"status":"ok"}`))
}

func (c *Core) handlerDisableSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")

	enc := json.NewEncoder(w)
	sIDs := strings.Replace(r.URL.Path, "/scheduler/core/disable/", "", -1)
	sID, err := uuid.Parse(sIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	if err := c.DisableSchedule(r.Context(), sID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	enc.Encode([]byte(`{"status":"ok"}`))
}
