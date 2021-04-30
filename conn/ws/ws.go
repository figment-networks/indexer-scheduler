package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/figment-networks/indexer-scheduler/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var ErrConnectionClosed = errors.New("connection closed")

type JsonRPCRequest struct {
	ID      uint64        `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JsonRPCSend struct {
	JsonRPCRequest
	Sid    string
	RespCH chan conn.Response
}

type JsonRPCError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type JsonRPCResponse struct {
	ID      uint64          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Error   *JsonRPCError   `json:"error,omitempty"`
	Result  json.RawMessage `json:"result"`
}
type ResponseStore struct {
	ID     uint64 `json:"id"` // originalID
	Sid    string
	Type   string
	RespCH chan conn.Response
}

type Conn struct {
	l        *zap.Logger
	Requests chan JsonRPCSend
	Closes   chan JsonRPCSend
	Closed   bool
}

type LockedResponseMap struct {
	Map map[uint64]ResponseStore
	L   sync.RWMutex
}

func NewConn(l *zap.Logger) *Conn {
	return &Conn{
		l:        l,
		Closes:   make(chan JsonRPCSend),
		Requests: make(chan JsonRPCSend),
	}
}

func (co *Conn) CloseStream(sid string) error {
	if !co.Closed {
		resp := make(chan conn.Response)
		co.Closes <- JsonRPCSend{
			Sid:    sid,
			RespCH: resp,
		}
		r := <-resp
		close(resp)

		return r.Error
	}
	return nil
}

// Send is there just because of mock, it doesn't make much sense otherwise
func (co *Conn) Send(sid string, ch chan conn.Response, id uint64, method string, params []interface{}) {
	co.Requests <- JsonRPCSend{
		Sid:            sid,
		RespCH:         ch,
		JsonRPCRequest: JsonRPCRequest{ID: id, Method: method, Params: params},
	}
}

func (co *Conn) recv(ctx context.Context, c *websocket.Conn, done chan struct{}, resps *LockedResponseMap) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			co.l.Error("error reading next message", zap.Error(err))
			return
		}

		res := &JsonRPCResponse{}
		err = json.Unmarshal(message, res)
		if err != nil {
			co.l.Error("error unmarshaling jsonrpc response", zap.Error(err))
			continue
		}

		resps.L.Lock()
		ch := resps.Map[res.ID]

		response := conn.Response{
			ID:     ch.ID,
			Type:   ch.Type,
			Result: res.Result,
		}

		if res.Error != nil {
			response.Error = fmt.Errorf("error in service %s", res.Error.Message)
		}

		ch.RespCH <- response
		delete(resps.Map, res.ID)
		resps.L.Unlock()
	}
}

func (conn *Conn) Run(ctx context.Context, addr string) {
	f := make(chan struct{}, 1)
	multipliers := []int{1, 1, 1, 2, 3, 4, 6, 10}
	var i int
	go conn.run(ctx, addr, f)
	for {
		select { // reconnects respecting context
		case <-ctx.Done():
			return
		case <-f:
			var tryM int
			if len(multipliers) <= i {
				tryM = multipliers[7]
			} else {
				tryM = multipliers[i]
			}

			<-time.After(time.Second * time.Duration(tryM))

			go conn.run(ctx, addr, f)
			i++
		}
	}
}

func (co *Conn) run(ctx context.Context, addr string, f chan struct{}) {
	defer co.l.Sync()
	var nextMessageID uint64

	responseMap := &LockedResponseMap{Map: make(map[uint64]ResponseStore)}

	urlHost := url.URL{Scheme: "ws", Host: addr, Path: "ws"}
	co.l.Info("[API] Connecting to websocket ", zap.String("host", urlHost.String()))
	c, _, err := websocket.DefaultDialer.DialContext(ctx, urlHost.String(), nil)
	if err != nil {
		co.l.Error("[API] Error connecting to websocket ", zap.String("host", addr), zap.Error(err))
		f <- struct{}{}
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go co.recv(ctx, c, done, responseMap)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	buff := new(bytes.Buffer)
	enc := json.NewEncoder(buff)
WSLOOP:
	for {
		select {
		case cl := <-co.Closes:
			responseMap.L.Lock()
		MapClean:
			for k, rR := range responseMap.Map {
				if rR.Sid == cl.Sid {
					delete(responseMap.Map, k)
					break MapClean
				}
			}
			responseMap.L.Unlock()
			cl.RespCH <- conn.Response{}
		case <-ctx.Done():
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				co.l.Error("[API] Error closing websocket ", zap.Error(err))
				break WSLOOP
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			break WSLOOP
		case <-done:
			break WSLOOP
		case req := <-co.Requests:

			originalID := req.ID
			req.ID = nextMessageID
			if req.JSONRPC == "" {
				req.JSONRPC = "2.0"
			}

			nextMessageID++
			if err := enc.Encode(req.JsonRPCRequest); err != nil {
				req.RespCH <- conn.Response{
					ID:    originalID,
					Type:  req.Method,
					Error: fmt.Errorf("error encoding message: %w ", err),
				}
				continue WSLOOP
			}
			responseMap.L.Lock()
			responseMap.Map[req.ID] = ResponseStore{
				ID:     originalID,
				Sid:    req.Sid,
				Type:   req.Method,
				RespCH: req.RespCH,
			}
			responseMap.L.Unlock()
			err = c.WriteMessage(websocket.TextMessage, buff.Bytes())
			buff.Reset()
			if err != nil {
				co.l.Error("[API] Error sending data websocket ", zap.Error(err))
				break WSLOOP
			}
		}
	}
	co.Closed = true
	responseMap.L.RLock()
	for _, resp := range responseMap.Map {
		// send on closed
		resp.RespCH <- conn.Response{
			ID:    resp.ID,
			Type:  resp.Type,
			Error: ErrConnectionClosed,
		}
	}
	responseMap.L.RUnlock()

	co.l.Info("[API] Websocket listener finished", zap.String("host", addr))
	f <- struct{}{}
}
