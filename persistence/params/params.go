package params

import "errors"

var (
	ErrNotFound         = errors.New("record not found")
	ErrAlreadyRegistred = errors.New("already registred")
)

type State string

var (
	StateFinished State = "finished"
	StateRunning  State = "running"
)
