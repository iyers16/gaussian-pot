package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iyers16/gaussian-pot/backend/internal/game"
	"github.com/iyers16/gaussian-pot/backend/internal/model"
	"github.com/iyers16/gaussian-pot/backend/internal/repository"
)

// DebtHandler handles debt confirmation endpoints.
type DebtHandler struct {
	debtRepo  *repository.DebtRepository
	roundRepo *repository.RoundRepository
	sm        *game.StateMachine
	hub       *Hub
}

func NewDebtHandler(debtRepo *repository.DebtRepository, roundRepo *repository.RoundRepository, sm *game.StateMachine, hub *Hub) *DebtHandler {
	return &DebtHandler{debtRepo: debtRepo, roundRepo: roundRepo, sm: sm, hub: hub}
}

// ConfirmPaid handles POST /api/debts/:id/confirm-paid
func (h *DebtHandler) ConfirmPaid(c *gin.Context) {
	sess := sessionFromContext(c)
	if sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	debtID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt id"})
		return
	}

	debt, err := h.debtRepo.GetByID(debtID)
	if err != nil || debt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
		return
	}
	if debt.PayerID != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the payer of this debt"})
		return
	}

	updated, err := h.debtRepo.ConfirmPaid(debtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	h.hub.Broadcast("debt_update", updated)
	h.checkAllSettled()

	c.JSON(http.StatusOK, updated)
}

// ConfirmReceived handles POST /api/debts/:id/confirm-received
func (h *DebtHandler) ConfirmReceived(c *gin.Context) {
	sess := sessionFromContext(c)
	if sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	debtID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt id"})
		return
	}

	debt, err := h.debtRepo.GetByID(debtID)
	if err != nil || debt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
		return
	}
	if debt.PayeeID != sess.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the payee of this debt"})
		return
	}

	updated, err := h.debtRepo.ConfirmReceived(debtID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	h.hub.Broadcast("debt_update", updated)
	h.checkAllSettled()

	c.JSON(http.StatusOK, updated)
}

// GetMyDebts handles GET /api/debts
func (h *DebtHandler) GetMyDebts(c *gin.Context) {
	sess := sessionFromContext(c)
	if sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	debts, err := h.debtRepo.GetByUser(sess.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if debts == nil {
		debts = []model.Debt{}
	}
	c.JSON(http.StatusOK, debts)
}

// checkAllSettled transitions the game back to IDLE when every debt is settled.
func (h *DebtHandler) checkAllSettled() {
	if h.sm.State() != model.StateSettling {
		return
	}
	hasPending, err := h.debtRepo.HasAnyPending()
	if err != nil || hasPending {
		return
	}
	h.sm.TransitionTo(model.StateIdle, 0)
	h.hub.Broadcast("round_settled", gin.H{})
}
