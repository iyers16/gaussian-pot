package game

import (
	"testing"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

func TestComputeDistribution_CorrectPointCount(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "a", Guess: 10, Wager: 100},
		{UserID: 2, Username: "b", Guess: 20, Wager: 100},
	}
	dist := ComputeDistribution(bets, 0, model.ModeSniper, false)
	if len(dist.CurvePoints) != 100 {
		t.Errorf("expected 100 curve points, got %d", len(dist.CurvePoints))
	}
}

func TestComputeDistribution_BetsOnLine(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "alice", Guess: 5, Wager: 50},
		{UserID: 2, Username: "bob", Guess: 15, Wager: 75},
	}
	dist := ComputeDistribution(bets, 0, model.ModeSocial, false)
	if len(dist.BetsOnNumberLine) != 2 {
		t.Errorf("expected 2 bets on line, got %d", len(dist.BetsOnNumberLine))
	}
	if dist.BetsOnNumberLine[0].Username != "alice" {
		t.Errorf("expected alice first, got %s", dist.BetsOnNumberLine[0].Username)
	}
}

func TestComputeDistribution_YValuesInUnitRange(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "a", Guess: 10, Wager: 100},
	}
	dist := ComputeDistribution(bets, 0, model.ModeSniper, false)
	for _, p := range dist.CurvePoints {
		if p.Y < 0 || p.Y > 1.0001 {
			t.Errorf("y value %.4f out of [0,1] range at x=%.2f", p.Y, p.X)
		}
	}
}

func TestComputeDistribution_EmptyBets(t *testing.T) {
	dist := ComputeDistribution(nil, 10, model.ModeSniper, true)
	if len(dist.CurvePoints) != 100 {
		t.Errorf("expected 100 points even for empty bets, got %d", len(dist.CurvePoints))
	}
	if len(dist.BetsOnNumberLine) != 0 {
		t.Errorf("expected 0 bets on line, got %d", len(dist.BetsOnNumberLine))
	}
}

func TestComputeDistribution_XRangeIsMonotonic(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "a", Guess: 30, Wager: 100},
		{UserID: 2, Username: "b", Guess: 70, Wager: 100},
	}
	dist := ComputeDistribution(bets, 50, model.ModeVolatile, true)
	for i := 1; i < len(dist.CurvePoints); i++ {
		if dist.CurvePoints[i].X <= dist.CurvePoints[i-1].X {
			t.Errorf("x values are not strictly increasing at index %d", i)
		}
	}
}
