package game

import (
	"math"
	"testing"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

func TestComputePayouts_SumIsConserved(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "alice", Guess: 10, Wager: 100},
		{UserID: 2, Username: "bob", Guess: 15, Wager: 200},
		{UserID: 3, Username: "carol", Guess: 20, Wager: 150},
	}
	target := 12.0

	for _, mode := range []model.BettingMode{model.ModeSniper, model.ModeSocial, model.ModeVolatile} {
		results := ComputePayouts(bets, target, mode)
		totalPot := 100.0 + 200.0 + 150.0
		totalPayout := 0.0
		for _, r := range results {
			totalPayout += r.Payout
		}
		if math.Abs(totalPayout-totalPot) > 0.01 {
			t.Errorf("mode %s: payout sum %.2f != pot %.2f", mode, totalPayout, totalPot)
		}
	}
}

func TestComputePayouts_CloserGuessPaysMore(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "close", Guess: 10, Wager: 100},
		{UserID: 2, Username: "far", Guess: 50, Wager: 100},
	}
	target := 10.0
	results := ComputePayouts(bets, target, model.ModeSniper)

	// results are sorted by payout desc; close should be rank 1
	if results[0].Username != "close" {
		t.Errorf("expected 'close' to rank first, got %s", results[0].Username)
	}
	if results[0].Payout <= results[1].Payout {
		t.Errorf("closer guess should yield higher payout: %.2f vs %.2f", results[0].Payout, results[1].Payout)
	}
}

func TestComputePayouts_NetSumsToZero(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "a", Guess: 5, Wager: 50},
		{UserID: 2, Username: "b", Guess: 10, Wager: 100},
		{UserID: 3, Username: "c", Guess: 20, Wager: 75},
	}
	for _, mode := range []model.BettingMode{model.ModeSniper, model.ModeSocial, model.ModeVolatile} {
		results := ComputePayouts(bets, 10, mode)
		netSum := 0.0
		for _, r := range results {
			netSum += r.Net
		}
		if math.Abs(netSum) > 0.01 {
			t.Errorf("mode %s: net sum %.4f should be ~0", mode, netSum)
		}
	}
}

func TestComputePayouts_EmptyBets(t *testing.T) {
	results := ComputePayouts(nil, 10, model.ModeSniper)
	if results != nil {
		t.Error("expected nil result for empty bets")
	}
}

func TestComputePayouts_SingleBetGetsFullPot(t *testing.T) {
	bets := []BetInput{
		{UserID: 1, Username: "solo", Guess: 99, Wager: 300},
	}
	results := ComputePayouts(bets, 10, model.ModeSocial)
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}
	if math.Abs(results[0].Payout-300) > 0.01 {
		t.Errorf("single player should get full pot 300, got %.2f", results[0].Payout)
	}
}

func TestSigmaSniper_ShrinksWithWager(t *testing.T) {
	small := []BetInput{{UserID: 1, Username: "a", Guess: 1, Wager: 10}}
	large := []BetInput{{UserID: 1, Username: "a", Guess: 1, Wager: 10000}}

	sigmaSmall := computeSigma(small, model.ModeSniper)
	sigmaLarge := computeSigma(large, model.ModeSniper)

	if sigmaSmall <= sigmaLarge {
		t.Errorf("sniper: sigma should shrink with larger wager; small=%.3f large=%.3f", sigmaSmall, sigmaLarge)
	}
}

func TestSigmaSocial_GrowsWithWager(t *testing.T) {
	small := []BetInput{{UserID: 1, Username: "a", Guess: 1, Wager: 10}}
	large := []BetInput{{UserID: 1, Username: "a", Guess: 1, Wager: 10000}}

	sigmaSmall := computeSigma(small, model.ModeSocial)
	sigmaLarge := computeSigma(large, model.ModeSocial)

	if sigmaSmall >= sigmaLarge {
		t.Errorf("social: sigma should grow with larger wager; small=%.3f large=%.3f", sigmaSmall, sigmaLarge)
	}
}

func TestSigmaVolatile_UsesGuessSpread(t *testing.T) {
	tight := []BetInput{
		{UserID: 1, Username: "a", Guess: 10, Wager: 100},
		{UserID: 2, Username: "b", Guess: 11, Wager: 100},
	}
	wide := []BetInput{
		{UserID: 1, Username: "a", Guess: 10, Wager: 100},
		{UserID: 2, Username: "b", Guess: 90, Wager: 100},
	}
	sigmaTight := computeSigma(tight, model.ModeVolatile)
	sigmaWide := computeSigma(wide, model.ModeVolatile)

	if sigmaTight >= sigmaWide {
		t.Errorf("volatile: sigma should grow with wider guess spread; tight=%.3f wide=%.3f", sigmaTight, sigmaWide)
	}
}
