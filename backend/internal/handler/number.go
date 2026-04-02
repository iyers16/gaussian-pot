package handler

import (
	"math/rand"
	"net/http"

	"github.com/iyers16/gaussian-pot/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type NumberHandler struct {
	repo *repository.NumberRepository
}

func NewNumberHandler(repo *repository.NumberRepository) *NumberHandler {
	return &NumberHandler{repo: repo}
}

func (h *NumberHandler) Generate(c *gin.Context) {
	value := rand.Intn(100) + 1
	number, err := h.repo.Save(value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, number)
}

func (h *NumberHandler) GetAll(c *gin.Context) {
	numbers, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, numbers)
}
