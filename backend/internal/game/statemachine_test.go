package game

import (
	"testing"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

func TestStateMachine_InitialStateIsIdle(t *testing.T) {
	sm := NewStateMachine()
	if sm.State() != model.StateIdle {
		t.Errorf("expected initial state IDLE, got %s", sm.State())
	}
}

func TestStateMachine_ValidTransitions(t *testing.T) {
	sm := NewStateMachine()

	steps := []struct {
		to      model.GameState
		roundID int
	}{
		{model.StateRoundOpen, 1},
		{model.StateRoundCalled, 1},
		{model.StateSettling, 1},
		{model.StateIdle, 0},
	}

	for _, step := range steps {
		if err := sm.TransitionTo(step.to, step.roundID); err != nil {
			t.Errorf("unexpected error transitioning to %s: %v", step.to, err)
		}
		if sm.State() != step.to {
			t.Errorf("expected state %s, got %s", step.to, sm.State())
		}
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := NewStateMachine()
	// Cannot jump from IDLE directly to ROUND_CALLED.
	if err := sm.TransitionTo(model.StateRoundCalled, 1); err == nil {
		t.Error("expected error for invalid transition IDLE→ROUND_CALLED")
	}
}

func TestStateMachine_CannotSkipSettling(t *testing.T) {
	sm := NewStateMachine()
	sm.TransitionTo(model.StateRoundOpen, 1)
	sm.TransitionTo(model.StateRoundCalled, 1)
	// Cannot go directly from ROUND_CALLED to IDLE.
	if err := sm.TransitionTo(model.StateIdle, 0); err == nil {
		t.Error("expected error for transition ROUND_CALLED→IDLE")
	}
}

func TestStateMachine_AssertState_Pass(t *testing.T) {
	sm := NewStateMachine()
	if err := sm.AssertState(model.StateIdle); err != nil {
		t.Errorf("expected no error for asserting IDLE on IDLE machine: %v", err)
	}
}

func TestStateMachine_AssertState_Fail(t *testing.T) {
	sm := NewStateMachine()
	if err := sm.AssertState(model.StateRoundOpen); err == nil {
		t.Error("expected error asserting ROUND_OPEN on IDLE machine")
	}
}

func TestStateMachine_RoundID(t *testing.T) {
	sm := NewStateMachine()
	sm.TransitionTo(model.StateRoundOpen, 42)
	if sm.RoundID() != 42 {
		t.Errorf("expected round ID 42, got %d", sm.RoundID())
	}
}
