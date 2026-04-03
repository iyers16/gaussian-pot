package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iyers16/gaussian-pot/backend/internal/game"
	"github.com/iyers16/gaussian-pot/backend/internal/model"
	"github.com/iyers16/gaussian-pot/backend/internal/repository"
)

// RoundHandler handles bet placement.
type RoundHandler struct {
	betRepo   *repository.BetRepository
	debtRepo  *repository.DebtRepository
	userRepo  *repository.UserRepository
	roundRepo *repository.RoundRepository
	sm        *game.StateMachine
	hub       *Hub
}

func NewRoundHandler(
	betRepo *repository.BetRepository,
	debtRepo *repository.DebtRepository,
	userRepo *repository.UserRepository,
	roundRepo *repository.RoundRepository,
	sm *game.StateMachine,
	hub *Hub,
) *RoundHandler {
	return &RoundHandler{
		betRepo:   betRepo,
		debtRepo:  debtRepo,
		userRepo:  userRepo,
		roundRepo: roundRepo,
		sm:        sm,
		hub:       hub,
	}
}

type placeBetRequest struct {
	Guess float64 `json:"guess" binding:"required"`
	Wager float64 `json:"wager" binding:"required"`
}

// PlaceBet handles POST /api/round/bet
func (h *RoundHandler) PlaceBet(c *gin.Context) {
	sess := sessionFromContext(c)
	if sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	if err := h.sm.AssertState(model.StateRoundOpen); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "round is not open for betting"})
		return
	}

	// Block players with unsettled debts.
	hasPending, err := h.debtRepo.HasPendingDebts(sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if hasPending {
		c.JSON(http.StatusForbidden, gin.H{"error": "you have unsettled debts — settle them before betting"})
		return
	}

	var req placeBetRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Wager <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "guess and wager (> 0) are required"})
		return
	}

	// Verify the player has enough credits.
	user, err := h.userRepo.GetByID(sess.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	if user.Credits < req.Wager {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits"})
		return
	}

	roundID := h.sm.RoundID()

	// Prevent double-betting.
	existing, err := h.betRepo.GetByUserAndRound(sess.UserID, roundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already placed a bet this round"})
		return
	}

	bet, err := h.betRepo.Create(roundID, sess.UserID, sess.Username, req.Guess, req.Wager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to place bet"})
		return
	}

	// Deduct wager from credits immediately.
	if err := h.userRepo.AddCredits(sess.UserID, -req.Wager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct credits"})
		return
	}

	// Broadcast updated ticker to all clients.
	allBets, _ := h.betRepo.GetByRound(roundID)
	h.hub.Broadcast("bet_ticker", gin.H{"bets": allBets})

	// Broadcast updated distribution curve.
	inputs := betsToInputs(allBets)
	round, _ := h.roundRepo.GetByID(roundID)
	if round != nil {
		dist := game.ComputeDistribution(inputs, 0, round.Mode, false)
		h.hub.Broadcast("distribution_update", dist)
	}

	c.JSON(http.StatusOK, bet)
}

// GetCurrentRound returns the current round state (without target if not called).
func (h *RoundHandler) GetCurrentRound(c *gin.Context) {
	round, err := h.roundRepo.GetCurrent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if round == nil {
		c.JSON(http.StatusOK, gin.H{"state": string(model.StateIdle)})
		return
	}

	// Strip target_value unless round is called.
	resp := gin.H{
		"id":            round.ID,
		"question_text": round.QuestionText,
		"unit":          round.Unit,
		"mode":          round.Mode,
		"state":         round.State,
		"opened_at":     round.OpenedAt,
	}
	if round.State == model.StateRoundCalled || round.State == model.StateSettling {
		resp["target_value"] = round.TargetValue
		resp["called_at"] = round.CalledAt
	}
	c.JSON(http.StatusOK, resp)
}

func betsToInputs(bets []model.Bet) []game.BetInput {
	inputs := make([]game.BetInput, len(bets))
	for i, b := range bets {
		inputs[i] = game.BetInput{
			UserID:   b.UserID,
			Username: b.Username,
			Guess:    b.Guess,
			Wager:    b.Wager,
		}
	}
	return inputs
}
