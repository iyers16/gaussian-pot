package game

import (
	"errors"
	"sync"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

// StateMachine manages the current game state in memory.
// Persistence of state is via the rounds table; this is the in-memory gate.
type StateMachine struct {
	mu       sync.RWMutex
	state    model.GameState
	roundID  int // 0 when IDLE
}

var (
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrNotIdle           = errors.New("game is not in IDLE state")
	ErrNotRoundOpen      = errors.New("game is not in ROUND_OPEN state")
	ErrNotRoundCalled    = errors.New("game is not in ROUND_CALLED state")
	ErrNotSettling       = errors.New("game is not in SETTLING state")
)

func NewStateMachine() *StateMachine {
	return &StateMachine{state: model.StateIdle}
}

func (sm *StateMachine) State() model.GameState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

func (sm *StateMachine) RoundID() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.roundID
}

// TransitionTo moves to the next state, validating the transition.
func (sm *StateMachine) TransitionTo(next model.GameState, roundID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isValidTransition(sm.state, next) {
		return ErrInvalidTransition
	}
	sm.state = next
	sm.roundID = roundID
	return nil
}

func (sm *StateMachine) isValidTransition(from, to model.GameState) bool {
	switch from {
	case model.StateIdle:
		return to == model.StateRoundOpen
	case model.StateRoundOpen:
		return to == model.StateRoundCalled
	case model.StateRoundCalled:
		return to == model.StateSettling
	case model.StateSettling:
		return to == model.StateIdle
	default:
		return false
	}
}

// AssertState returns an error if the current state does not match expected.
func (sm *StateMachine) AssertState(expected model.GameState) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.state != expected {
		switch expected {
		case model.StateIdle:
			return ErrNotIdle
		case model.StateRoundOpen:
			return ErrNotRoundOpen
		case model.StateRoundCalled:
			return ErrNotRoundCalled
		case model.StateSettling:
			return ErrNotSettling
		}
		return ErrInvalidTransition
	}
	return nil
}
