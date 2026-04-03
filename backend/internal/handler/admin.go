package handler

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iyers16/gaussian-pot/backend/internal/game"
	"github.com/iyers16/gaussian-pot/backend/internal/model"
	"github.com/iyers16/gaussian-pot/backend/internal/questions"
	"github.com/iyers16/gaussian-pot/backend/internal/repository"
)

// AdminHandler handles host-only actions: open/call round, replenish credits.
type AdminHandler struct {
	roundRepo *repository.RoundRepository
	betRepo   *repository.BetRepository
	debtRepo  *repository.DebtRepository
	userRepo  *repository.UserRepository
	sm        *game.StateMachine
	hub       *Hub
}

func NewAdminHandler(
	roundRepo *repository.RoundRepository,
	betRepo *repository.BetRepository,
	debtRepo *repository.DebtRepository,
	userRepo *repository.UserRepository,
	sm *game.StateMachine,
	hub *Hub,
) *AdminHandler {
	return &AdminHandler{
		roundRepo: roundRepo,
		betRepo:   betRepo,
		debtRepo:  debtRepo,
		userRepo:  userRepo,
		sm:        sm,
		hub:       hub,
	}
}

type openRoundRequest struct {
	QuestionID int    `json:"question_id" binding:"required"`
	Mode       string `json:"mode" binding:"required"`
}

// OpenRound handles POST /api/admin/round/open
func (h *AdminHandler) OpenRound(c *gin.Context) {
	if err := h.sm.AssertState(model.StateIdle); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot open round: current state is " + string(h.sm.State())})
		return
	}

	// Check no pending debts.
	hasPending, err := h.debtRepo.HasAnyPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if hasPending {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot open round: unsettled debts remain"})
		return
	}

	var req openRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question_id and mode are required"})
		return
	}

	mode := model.BettingMode(req.Mode)
	if mode != model.ModeSniper && mode != model.ModeSocial && mode != model.ModeVolatile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be sniper, social, or volatile"})
		return
	}

	q := questionByID(req.QuestionID)
	if q == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question not found"})
		return
	}

	round, err := h.roundRepo.Create(q.ID, q.Text, q.TargetValue, q.Unit, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create round"})
		return
	}

	if err := h.sm.TransitionTo(model.StateRoundOpen, round.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "state transition failed"})
		return
	}

	h.hub.Broadcast("round_opened", gin.H{
		"question_text": round.QuestionText,
		"mode":          round.Mode,
		"unit":          round.Unit,
	})

	c.JSON(http.StatusOK, gin.H{"round_id": round.ID, "state": model.StateRoundOpen})
}

// CallRound handles POST /api/admin/round/call
func (h *AdminHandler) CallRound(c *gin.Context) {
	if err := h.sm.AssertState(model.StateRoundOpen); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "round is not open"})
		return
	}

	roundID := h.sm.RoundID()
	round, err := h.roundRepo.GetByID(roundID)
	if err != nil || round == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "round not found"})
		return
	}

	// Transition to ROUND_CALLED.
	if err := h.sm.TransitionTo(model.StateRoundCalled, roundID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "state transition failed"})
		return
	}
	h.roundRepo.UpdateState(roundID, model.StateRoundCalled)

	// Fetch all bets and compute payouts.
	bets, err := h.betRepo.GetByRound(roundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}

	inputs := betsToInputs(bets)
	results := game.ComputePayouts(inputs, round.TargetValue, round.Mode)

	// Persist payouts and build rankings payload.
	type rankEntry struct {
		Rank     int     `json:"rank"`
		Username string  `json:"username"`
		Guess    float64 `json:"guess"`
		Wager    float64 `json:"wager"`
		Payout   float64 `json:"payout"`
		Net      float64 `json:"net"`
	}
	rankings := make([]rankEntry, len(results))
	for i, res := range results {
		// Update payout in DB and credit user.
		h.betRepo.UpdatePayout(findBetID(bets, res.UserID), res.Payout)
		h.userRepo.AddCredits(res.UserID, res.Payout)
		rankings[i] = rankEntry{
			Rank:     i + 1,
			Username: res.Username,
			Guess:    res.Guess,
			Wager:    res.Wager,
			Payout:   res.Payout,
			Net:      res.Net,
		}
	}

	// Resolve P2P debts.
	nets := make([]game.PlayerNet, len(results))
	for i, res := range results {
		nets[i] = game.PlayerNet{
			UserID:   res.UserID,
			Username: res.Username,
			Net:      res.Net,
		}
	}
	edges := game.ResolveDebts(nets)

	// Persist debts.
	debtPayloads := []gin.H{}
	for _, e := range edges {
		debt, err := h.debtRepo.Create(roundID, e.PayerID, e.PayeeID, e.PayerUsername, e.PayeeUsername, e.Amount)
		if err != nil {
			continue
		}
		debtPayloads = append(debtPayloads, gin.H{
			"id":             debt.ID,
			"payer_username": debt.PayerUsername,
			"payee_username": debt.PayeeUsername,
			"amount":         debt.Amount,
			"status":         debt.Status,
		})
	}

	// Transition to SETTLING.
	h.sm.TransitionTo(model.StateSettling, roundID)
	h.roundRepo.UpdateState(roundID, model.StateSettling)

	// Compute final distribution with target revealed.
	dist := game.ComputeDistribution(inputs, round.TargetValue, round.Mode, true)

	h.hub.Broadcast("round_called", gin.H{
		"target_value": round.TargetValue,
		"unit":         round.Unit,
		"rankings":     rankings,
		"debts":        debtPayloads,
		"distribution": dist,
	})

	c.JSON(http.StatusOK, gin.H{
		"target_value": round.TargetValue,
		"rankings":     rankings,
		"debts":        debtPayloads,
	})
}

type replenishRequest struct {
	Username string  `json:"username" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
}

// ReplenishCredits handles POST /api/admin/credits/replenish
func (h *AdminHandler) ReplenishCredits(c *gin.Context) {
	var req replenishRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and amount (> 0) required"})
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := h.userRepo.AddCredits(user.ID, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update credits"})
		return
	}

	updated, _ := h.userRepo.GetByID(user.ID)
	c.JSON(http.StatusOK, gin.H{"username": updated.Username, "credits": updated.Credits})
}

// RandomQuestion handles POST /api/admin/question/random
func (h *AdminHandler) RandomQuestion(c *gin.Context) {
	q := questions.Bank[rand.Intn(len(questions.Bank))]
	c.JSON(http.StatusOK, q)
}

// ListQuestions handles GET /api/admin/questions
func (h *AdminHandler) ListQuestions(c *gin.Context) {
	c.JSON(http.StatusOK, questions.Bank)
}

// ListUsers handles GET /api/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func questionByID(id int) *questions.Question {
	for i := range questions.Bank {
		if questions.Bank[i].ID == id {
			return &questions.Bank[i]
		}
	}
	return nil
}

func findBetID(bets []model.Bet, userID int) int {
	for _, b := range bets {
		if b.UserID == userID {
			return b.ID
		}
	}
	return 0
}
