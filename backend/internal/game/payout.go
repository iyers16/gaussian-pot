package game

import (
	"math"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

// BetInput carries the data needed for payout calculation.
type BetInput struct {
	UserID   int
	Username string
	Guess    float64
	Wager    float64
}

// PayoutResult is the computed payout for one player.
type PayoutResult struct {
	UserID   int
	Username string
	Guess    float64
	Wager    float64
	Score    float64
	Share    float64
	Payout   float64
	Net      float64
}

// ComputePayouts calculates payout shares for all bets given the target value and mode.
// Returns results sorted by payout descending (rank 1 = index 0).
func ComputePayouts(bets []BetInput, target float64, mode model.BettingMode) []PayoutResult {
	if len(bets) == 0 {
		return nil
	}

	sigma := computeSigma(bets, mode)

	results := make([]PayoutResult, len(bets))
	totalScore := 0.0
	totalPot := 0.0

	for i, b := range bets {
		d := math.Abs(b.Guess - target)
		score := gaussianScore(b, d, sigma, mode)
		results[i] = PayoutResult{
			UserID:   b.UserID,
			Username: b.Username,
			Guess:    b.Guess,
			Wager:    b.Wager,
			Score:    score,
		}
		totalScore += score
		totalPot += b.Wager
	}

	// Normalize into shares and compute payouts.
	for i := range results {
		if totalScore > 0 {
			results[i].Share = results[i].Score / totalScore
		} else {
			// Degenerate case: equal split.
			results[i].Share = 1.0 / float64(len(bets))
		}
		results[i].Payout = results[i].Share * totalPot
		results[i].Net = results[i].Payout - results[i].Wager
	}

	// Sort by payout descending.
	sortByPayoutDesc(results)
	return results
}

// gaussianScore computes the unnormalized Gaussian kernel score per mode.
func gaussianScore(b BetInput, distance, sigma float64, mode model.BettingMode) float64 {
	kernel := math.Exp(-(distance * distance) / (2 * sigma * sigma))
	switch mode {
	case model.ModeSniper, model.ModeSocial:
		// Wager amplifies the score.
		return b.Wager * kernel
	case model.ModeVolatile:
		// Wager is ignored for curve shaping; pure proximity score.
		return kernel
	default:
		return b.Wager * kernel
	}
}

// computeSigma derives the curve width based on the betting mode.
func computeSigma(bets []BetInput, mode model.BettingMode) float64 {
	switch mode {
	case model.ModeSniper:
		// Sigma shrinks as total wager grows → winner-take-most.
		totalWager := totalWagerMass(bets)
		// Base sigma 10, reduced by wager mass.
		sigma := 10.0 / (1.0 + math.Log1p(totalWager/100.0))
		return math.Max(sigma, 0.5)

	case model.ModeSocial:
		// Sigma grows with total wager → flat, generous curve.
		totalWager := totalWagerMass(bets)
		sigma := 5.0 * (1.0 + math.Log1p(totalWager/100.0))
		return math.Min(sigma, 50.0)

	case model.ModeVolatile:
		// Sigma is the standard deviation of all guesses.
		return guessStdDev(bets)

	default:
		return 10.0
	}
}

func totalWagerMass(bets []BetInput) float64 {
	total := 0.0
	for _, b := range bets {
		total += b.Wager
	}
	return total
}

func guessStdDev(bets []BetInput) float64 {
	if len(bets) < 2 {
		return 5.0 // fallback for single-player edge case
	}
	mean := 0.0
	for _, b := range bets {
		mean += b.Guess
	}
	mean /= float64(len(bets))

	variance := 0.0
	for _, b := range bets {
		d := b.Guess - mean
		variance += d * d
	}
	variance /= float64(len(bets))
	std := math.Sqrt(variance)
	if std < 0.5 {
		return 0.5 // prevent division by zero
	}
	return std
}

func sortByPayoutDesc(results []PayoutResult) {
	// Simple insertion sort (small N — max 8 players).
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Payout < key.Payout {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
