package subFreezeHandler

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type SubFreezeService interface {
	FreezeSubscription(ctx context.Context, subscriptionNumber string, freezeStart time.Time) error
	UnfreezeSubscription(ctx context.Context, subscriptionNumber string, unfreezeDate time.Time) error
	GetAllActiveFreeze(ctx context.Context) ([]models.SubscriptionFreeze, error)
}

type SubFreezeHandler struct {
	ctx              context.Context
	log              *slog.Logger
	subFreezeService SubFreezeService
}

func New(
	ctx context.Context,
	log *slog.Logger,
	subFreezeService SubFreezeService,
) *SubFreezeHandler {
	return &SubFreezeHandler{
		ctx:              ctx,
		log:              log,
		subFreezeService: subFreezeService,
	}
}

func (h *SubFreezeHandler) FreezeSubscription(c *gin.Context) {
	log := h.log.With(slog.String("op", "handlers.sub_freeze.FreezeSubscription"))

	var req models.SubscriptionFreeze
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.SubscriptionNumber == "" || req.FreezeStart.IsZero() {
		log.Error("missing required fields in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	err := h.subFreezeService.FreezeSubscription(c.Request.Context(), req.SubscriptionNumber, req.FreezeStart)
	if err != nil {
		log.Error("failed to freeze subscription", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription frozen successfully"})
}

type UnfreezeRequest struct {
	SubscriptionNumber string    `json:"subscription_number"`
	UnfreezeDate       time.Time `json:"unfreeze_date"`
}

func (h *SubFreezeHandler) UnfreezeSubscription(c *gin.Context) {
	log := h.log.With(slog.String("op", "handlers.sub_freeze.UnfreezeSubscription"))

	var req UnfreezeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.SubscriptionNumber == "" || req.UnfreezeDate.IsZero() {
		log.Error("missing required fields in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	err := h.subFreezeService.UnfreezeSubscription(c.Request.Context(), req.SubscriptionNumber, req.UnfreezeDate)
	if err != nil {
		log.Error("failed to unfreeze subscription", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription unfrozen successfully"})
}

func (h *SubFreezeHandler) GetAllActiveFreeze(c *gin.Context) {

	log := h.log.With(slog.String("op", "handlers.sub_freeze.GetActiveFreeze"))

	freezedSubs, err := h.subFreezeService.GetAllActiveFreeze(c.Request.Context())
	if err != nil {
		log.Error("failed to get active freezedSubs", slog.Any("error", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "no active freezedSubs found"})
		return
	}

	c.JSON(http.StatusOK, freezedSubs)
}
