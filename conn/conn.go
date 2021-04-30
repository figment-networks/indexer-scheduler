package conn

import "encoding/json"

type Response struct {
	ID     uint64
	Error  error
	Type   string
	Result json.RawMessage
}

type RPCConnector interface {
	Send(streamID string, ch chan Response, id uint64, method string, params []interface{})
	CloseStream(streamID string) error
}
