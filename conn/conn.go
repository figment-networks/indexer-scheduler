package conn

import "encoding/json"

type Response struct {
	ID     uint64
	Error  error
	Type   string
	Result json.RawMessage
}

type RPCConnector interface {
	Send(ch chan Response, id uint64, method string, params []interface{})
}
