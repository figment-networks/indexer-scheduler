package destination

import (
	"context"

	"github.com/figment-networks/indexer-scheduler/conn/tray"
	"github.com/figment-networks/indexer-scheduler/destination/manager"
	"github.com/figment-networks/indexer-scheduler/structures"
	"go.uber.org/zap"
)

type Destination interface {
	Load(structures.Target, *tray.ConnTray)
}

type Container struct {
	logger       *zap.Logger
	destinations map[string]Destination
}

func NewContainer(logger *zap.Logger) (c *Container) {
	return &Container{
		logger:       logger,
		destinations: make(map[string]Destination),
	}
}

func (c *Container) Add(ctx context.Context, t structures.TargetConfig, ct *tray.ConnTray, ta manager.TargetAdder) error {
	switch t.Type {
	case "manager":
		m := manager.NewManager(c.logger, ta)
		go m.Load(ctx, t.Target, ct)
	default:
		ta.Add(t.Target)
	}

	return nil
}
