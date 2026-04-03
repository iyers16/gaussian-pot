package game

import (
	"math"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

// CurvePoint is a single (x, y) coordinate on the Gaussian curve.
type CurvePoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BetOnNumberLine is a player's guess rendered on the number line.
type BetOnNumberLine struct {
	Username string  `json:"username"`
	Guess    float64 `json:"guess"`
	Wager    float64 `json:"wager"`
}

// DistributionData is the complete payload broadcast to clients.
type DistributionData struct {
	CurvePoints     []CurvePoint      `json:"curve_points"`
	BetsOnNumberLine []BetOnNumberLine `json:"bets_on_numberline"`
}

// ComputeDistribution pre-computes the Gaussian curve and numberline positions
// for the current set of bets. target is only used when calling the round;
// pass 0 during betting (curve is shown relative to guess spread).
func ComputeDistribution(bets []BetInput, target float64, mode model.BettingMode, showTarget bool) DistributionData {
	sigma := computeSigma(bets, mode)
	if sigma <= 0 {
		sigma = 5.0
	}

	// Determine the x-range for the curve: centre on guesses spread.
	center, spread := guessRange(bets, target, showTarget)
	padding := math.Max(sigma*3, spread*0.5)
	xMin := center - padding
	xMax := center + padding

	const numPoints = 100
	step := (xMax - xMin) / float64(numPoints-1)

	points := make([]CurvePoint, numPoints)
	for i := 0; i < numPoints; i++ {
		x := xMin + float64(i)*step
		// Normalised Gaussian density centred on best guess / target.
		y := math.Exp(-math.Pow(x-center, 2) / (2 * sigma * sigma))
		points[i] = CurvePoint{X: x, Y: y}
	}

	betsOnLine := make([]BetOnNumberLine, len(bets))
	for i, b := range bets {
		betsOnLine[i] = BetOnNumberLine{
			Username: b.Username,
			Guess:    b.Guess,
			Wager:    b.Wager,
		}
	}

	return DistributionData{
		CurvePoints:      points,
		BetsOnNumberLine: betsOnLine,
	}
}

// guessRange computes the centre and half-spread of the guess distribution.
func guessRange(bets []BetInput, target float64, showTarget bool) (center, spread float64) {
	if showTarget {
		center = target
	} else if len(bets) > 0 {
		sum := 0.0
		for _, b := range bets {
			sum += b.Guess
		}
		center = sum / float64(len(bets))
	} else {
		center = 0
	}

	if len(bets) == 0 {
		return center, 10
	}

	minG, maxG := bets[0].Guess, bets[0].Guess
	for _, b := range bets[1:] {
		if b.Guess < minG {
			minG = b.Guess
		}
		if b.Guess > maxG {
			maxG = b.Guess
		}
	}
	if showTarget {
		if target < minG {
			minG = target
		}
		if target > maxG {
			maxG = target
		}
	}
	spread = (maxG - minG) / 2.0
	return center, spread
}
