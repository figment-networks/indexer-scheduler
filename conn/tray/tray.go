package tray

import (
	"context"
	"errors"
	"sync"

	"github.com/figment-networks/indexer-scheduler/conn"
	"github.com/figment-networks/indexer-scheduler/conn/ws"
	"go.uber.org/zap"
)

type PAKey struct {
	Protocol string
	Address  string
}

type ConnTray struct {
	logger *zap.Logger
	l      sync.RWMutex
	conns  map[PAKey]conn.RPCConnector
}

func NewConnTray(logger *zap.Logger) *ConnTray {
	return &ConnTray{logger: logger, conns: make(map[PAKey]conn.RPCConnector)}
}

func (c *ConnTray) Get(protocol, address string) (conn.RPCConnector, error) {
	c.l.Lock()
	defer c.l.Unlock()
	cg, ok := c.conns[PAKey{protocol, address}]
	if ok {
		return cg, nil
	}

	switch protocol {
	case "ws":
		wsConn := ws.NewConn(c.logger)
		go wsConn.Run(context.Background(), address)
		c.conns[PAKey{protocol, address}] = wsConn
		return wsConn, nil
	case "http": // todo implement http
		return nil, errors.New("unknown protocol")
	default:
		return nil, errors.New("unknown protocol")

	}

	//ch := make(chan ws.Response, 10)
	//wsConn.Send(ch, 1, "get_workers", []interface{}{1232, "34543543"})

}
