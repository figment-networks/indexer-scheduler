package process

import (
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/figment-networks/indexer-scheduler/structures"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Marker interface {
	MarkFinished(ctx context.Context, id uuid.UUID) error
}

type Runner interface {
	Run(ctx context.Context, rcp structures.RunConfigParams) (backoff bool, err error)
	Name() string
}

type Running struct {
	Name string

	CancelFunc context.CancelFunc
}

type Scheduler struct {
	running map[string]Running
	runlock sync.Mutex
	logger  *zap.Logger
}

func NewScheduler(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		running: make(map[string]Running),
		logger:  logger,
	}
}

func (s *Scheduler) Run(ctx context.Context, name string, d time.Duration, rcp structures.RunConfigParams, r Runner) {
	cCtx, cancel := context.WithCancel(ctx)
	tckr := time.NewTicker(d)

	s.runlock.Lock()
	s.running[name] = Running{
		Name:       name,
		CancelFunc: cancel,
	}
	s.runlock.Unlock()

	var backoffCounter uint64
RunLoop:
	for {
		select {
		case <-tckr.C:
			backoff, err := r.Run(cCtx, rcp)

			if err != nil && err == io.EOF { // finish on end of processing
				tckr.Stop()
				break RunLoop
			}

			if backoff {
				backoffCounter++
				dur := calcBackoff(d, backoffCounter)
				s.logger.Info("[Process] Setting backoff", zap.Duration("duration", dur))
				tckr.Reset(dur)
			} else if backoffCounter > 0 {
				s.logger.Info("[Process] Resetting backoff", zap.Duration("duration", d))
				backoffCounter = 0
				tckr.Reset(d)
			}

			if err != nil {
				var rErr *structures.RunError
				s.logger.Error("[Process] Error running task", zap.String("name", name), zap.String("network", rcp.Network), zap.String("chain_id", rcp.ChainID), zap.String("task_id", rcp.TaskID), zap.String("version", rcp.Version), zap.Error(err))
				if errors.As(err, &rErr) {
					if !rErr.IsRecoverable() {
						tckr.Stop()
						break RunLoop
					}
				}
			}

		case <-cCtx.Done():
			tckr.Stop()
			break RunLoop
		case <-ctx.Done():
			tckr.Stop()
			break RunLoop
		}
	}

	s.runlock.Lock()
	delete(s.running, name)
	s.runlock.Unlock()
}

func (s *Scheduler) Stop(ctx context.Context, name string) {
	s.runlock.Lock()
	defer s.runlock.Unlock()

	r, ok := s.running[name]
	if ok {
		r.CancelFunc()
	}
}

var backoffsMultipliers = []float64{.5, 1, 1, 1.5, 2, 4, 4, 8, 8, 16, 16, 32}

func calcBackoff(initialDuration time.Duration, backoffIteration uint64) (finalDuration time.Duration) {

	bl := len(backoffsMultipliers)

	multiplier := backoffsMultipliers[bl-1]
	if backoffIteration < uint64(bl-1) {
		multiplier = backoffsMultipliers[backoffIteration]
	}

	newDur := float64(initialDuration.Milliseconds()) * multiplier
	newDur += newDur * rand.Float64()

	return time.Duration(math.Ceil(newDur * float64(time.Millisecond)))
}
