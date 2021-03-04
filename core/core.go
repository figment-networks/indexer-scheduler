package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	ErrAlreadyEnabled = errors.New("this schedule is already enabled")

	StatusEnabled Status = "enabled"
	StatusChanged Status = "changed"
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

	c.runLock.Lock()
	defer c.runLock.Unlock()
	rcs, err := c.coreStore.GetConfigs(ctx)

	if err != nil {
		return err
	}
	for _, s := range rcs {
		runner, ok := c.runners[s.Kind]
		if !ok {
			c.logger.Error(fmt.Sprintf("[Core] There is no such type as %s", s.Kind))
			continue
		}

		r, ok := c.run[s.ID]
		if !ok {
			r = &RunInfo{
				RunConfig: s,
			}

		} else {
			if r.Duration != s.Duration || r.RunID != s.RunID {
				c.logger.Info(fmt.Sprintf("[Core] Record changed reloading %s (%s:%s) %s in %s", runner.Name(), r.Network, r.ChainID, r.Version, r.Duration.String()))
				if r.CFunc != nil {
					r.CFunc()
				}
				r.Status = StatusChanged
			}

		}

		if r.Status == StatusEnabled {
			continue
		}

		// In fact run scheduler
		c.logger.Info(fmt.Sprintf("[Core] Running schedule %s (%s:%s) %s in %s", runner.Name(), r.Network, r.ChainID, r.Version, r.Duration.String()))
		var cCtx context.Context
		cCtx, r.CFunc = context.WithCancel(ctx)
		go c.scheduler.Run(cCtx, s.ID.String(), r.Duration, structures.RunConfigParams{Network: r.Network, ChainID: r.ChainID, TaskID: r.TaskID, Version: r.Version, Kind: r.Kind}, runner)

		if err := c.coreStore.MarkRunning(ctx, s.RunID, s.ID); err != nil {
			c.logger.Error("[Core] Error setting state running", zap.Error(err))
		}

		r.Status = StatusEnabled
		c.run[s.ID] = r
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

	runner, _ := c.runners[r.Kind]
	go c.scheduler.Run(ctx, sID.String(), r.Duration, structures.RunConfigParams{Network: r.Network, ChainID: r.ChainID, TaskID: r.TaskID, Version: r.Version, Kind: r.Kind}, runner)
	err := c.coreStore.MarkRunning(ctx, c.ID, sID)
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
}

func (c *Core) handlerListSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := c.ListSchedule()
	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(schedule)
}

func (c *Core) handlerEnableSchedule(w http.ResponseWriter, r *http.Request) {

	schedule := c.ListSchedule()
	enc := json.NewEncoder(w)
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(schedule)
}
