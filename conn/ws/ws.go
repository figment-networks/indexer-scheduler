package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/figment-networks/indexer-scheduler/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ConnStatus uint8

const (
	StateOffline ConnStatus = iota
	StateOnline
)

var ErrConnectionClosed = errors.New("connection closed")
var ErrRequestTimedout = errors.New("request timedout")

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
	Time   time.Time
	RespCH chan conn.Response
}

type Conn struct {
	l        *zap.Logger
	Requests chan JsonRPCSend
	Closes   chan JsonRPCSend
	Closed   bool

	resetConnections map[string]context.CancelFunc

	statusLock    sync.RWMutex
	status        map[string]ConnStatus
	statusEvenOne ConnStatus
}

type LockedResponseMap struct {
	Map map[uint64]ResponseStore
	L   sync.RWMutex
}

func NewConn(l *zap.Logger) *Conn {
	return &Conn{
		l:                l,
		Closes:           make(chan JsonRPCSend),
		Requests:         make(chan JsonRPCSend),
		resetConnections: make(map[string]context.CancelFunc),
		status:           make(map[string]ConnStatus),
	}
}

func (conn *Conn) getStatusEvenOne() ConnStatus {
	conn.statusLock.RLock()
	defer conn.statusLock.RUnlock()
	return conn.statusEvenOne
}

func (conn *Conn) getStatus() ConnStatus {
	conn.statusLock.RLock()
	defer conn.statusLock.RUnlock()
	for _, s := range conn.status {
		if s == StateOnline {
			return s
		}
	}

	return StateOffline
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
func (co *Conn) HealthCheck(ctx context.Context, tick time.Duration, healthCheckRequestFunc func(id uint64) JsonRPCRequest, healthCheckResponseFunc func(conn.Response) bool) {

	ch := make(chan conn.Response, 10)
	tckr := time.NewTicker(tick)

	var unHealthyRequests uint8

	var id uint64
	for {
		select {
		case <-tckr.C:
			status := co.getStatus()
			if status != StateOnline {
				co.l.Warn("[API] Connection is NOT Online")
				continue
			}

			id++
			// Check case when you cannot send anything (receivers are blocked)
			select {
			case co.Requests <- JsonRPCSend{RespCH: ch, JsonRPCRequest: healthCheckRequestFunc(id)}:
			case <-time.After(30 * time.Second):
				// FATAL ERROR
				os.Exit(1)
			}

			select {
			case a := <-ch:
				if !healthCheckResponseFunc(a) {
					co.l.Warn("[API] Bad Healthcheck")
					unHealthyRequests++
				} else {
					unHealthyRequests = 0
				}
			case <-time.After(30 * time.Second):
				co.l.Warn("[API] Response timed out")
				unHealthyRequests++
			}

			if unHealthyRequests == 10 {
				for _, canc := range co.resetConnections {
					canc()
				}
			}
			if unHealthyRequests > 20 {
				// FATAL ERROR
				os.Exit(1)
			}
		}

	}
}

// Send is there just because of mock, it doesn't make much sense otherwise
func (co *Conn) Send(streamID string, ch chan conn.Response, id uint64, method string, params []interface{}) error {
	if co.getStatusEvenOne() == StateOffline {
		return ErrConnectionClosed
	}

	co.Requests <- JsonRPCSend{
		RespCH:         ch,
		Sid:            streamID,
		JsonRPCRequest: JsonRPCRequest{ID: id, Method: method, Params: params},
	}
	return nil
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
		ch, ok := resps.Map[res.ID]
		if ok {
			response := conn.Response{
				ID:     ch.ID,
				Type:   ch.Type,
				Result: res.Result,
			}

			if res.Error != nil {
				response.Error = fmt.Errorf("error in service %s", res.Error.Message)
			}

			select {
			case ch.RespCH <- response:
			case <-time.After(time.Second * 5):
				co.l.Error("error unmarshaling jsonrpc response", zap.Error(err))
			}

			delete(resps.Map, res.ID)
		}

		resps.L.Unlock()
	}
}

func (conn *Conn) Run(ctx context.Context, addr string, connTimeout time.Duration) {
	f := make(chan struct{}, 1)
	multipliers := []int{1, 1, 1, 2, 3, 4, 6, 10}
	var i int

	cctx, close := context.WithCancel(ctx)
	conn.statusLock.Lock()
	conn.resetConnections[addr] = close
	conn.statusLock.Unlock()

	go conn.run(cctx, addr, f, connTimeout)
	for {
		select { // reconnects respecting context
		case <-ctx.Done():
			return
		case <-f:
			conn.statusLock.Lock()
			conn.status[addr] = StateOffline
			if reset, ok := conn.resetConnections[addr]; ok {
				reset()
			}
			cctx, close = context.WithCancel(ctx)
			conn.resetConnections[addr] = close
			conn.statusLock.Unlock()

			var tryM int
			if len(multipliers) <= i {
				tryM = multipliers[7]
			} else {
				tryM = multipliers[i]
			}

			<-time.After(time.Second * time.Duration(tryM))

			go conn.run(cctx, addr, f, connTimeout)
			i++
		}
	}
}

func (co *Conn) run(ctx context.Context, addr string, f chan struct{}, timeout time.Duration) {
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
	go co.timeoutChecker(ctx, timeout, done, responseMap)

	buff := new(bytes.Buffer)
	enc := json.NewEncoder(buff)
	co.statusLock.Lock()
	co.status[addr] = StateOnline
	co.statusEvenOne = StateOnline
	co.statusLock.Unlock()

WSLOOP:
	for {
		select {
		case cl := <-co.Closes:
			responseMap.L.Lock()
		MapClean:
			for k, rR := range responseMap.Map {
				if rR.Sid == cl.Sid {
					co.l.Info("[API] cleaning websocket ", zap.Uint64("id", k), zap.String("sid", cl.Sid))
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
			co.l.Info("[API] done websocket ")
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
				Time:   time.Now(),
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

func (co *Conn) timeoutChecker(ctx context.Context, timout time.Duration, done chan struct{}, responseMap *LockedResponseMap) {
	tckr := time.NewTicker(time.Second * 10)
	defer tckr.Stop()
	for {
		select { // reconnects respecting context
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-tckr.C:
			responseMap.L.Lock()
			for k, resp := range responseMap.Map {
				if time.Since(resp.Time) > timout {
					resp.RespCH <- conn.Response{
						ID:    resp.ID,
						Type:  resp.Type,
						Error: ErrRequestTimedout,
					}
					delete(responseMap.Map, k)
				}
			}
			responseMap.L.Unlock()
		}
	}
}
