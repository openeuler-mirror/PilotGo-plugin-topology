package service

import (
	"fmt"
	"sync/atomic"
)

// State describes the state of a service.
type State int64

// MarshalJSON marshal the connection state to JSON
func (s *State) MarshalJSON() ([]byte, error) {
	switch *s {
	case StartingState:
		return []byte("\"starting\""), nil
	case RunningState:
		return []byte("\"running\""), nil
	case StoppingState:
		return []byte("\"stopping\""), nil
	case StoppedState:
		return []byte("\"stopped\""), nil
	}
	return nil, fmt.Errorf("Invalid state: %d", s)
}

// Store atomatically stores the state
func (s *State) Store(state State) {
	atomic.StoreInt64((*int64)(s), int64(state))
}

// Load atomatically loads and returns the state
func (s *State) Load() State {
	return State(atomic.LoadInt64((*int64)(s)))
}

// CompareAndSwap executes the compare-and-swap operation for a state
func (s *State) CompareAndSwap(old, new State) bool {
	return atomic.CompareAndSwapInt64((*int64)(s), int64(old), int64(new))
}

const (
	// StoppedState service stopped
	StoppedState State = iota + 1
	// StartingState service starting
	StartingState
	// RunningState service running
	RunningState
	// StoppingState service stopping
	StoppingState
)
