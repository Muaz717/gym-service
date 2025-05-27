package statistics

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type StatService interface {
	TotalClients(ctx context.Context) (int, error)
	NewClients(ctx context.Context, from, to time.Time) (int, error)
	TotalIncome(ctx context.Context) (float64, error)
	Income(ctx context.Context, from, to time.Time) (float64, error)
	TotalSoldSubscriptions(ctx context.Context) (int, error)
	SoldSubscriptions(ctx context.Context, from, to time.Time) (int, error)
}

type StatHandler struct {
	ctx         context.Context
	log         *slog.Logger
	statService StatService
}

func New(
	ctx context.Context,
	log *slog.Logger,
	statService StatService,
) *StatHandler {
	return &StatHandler{
		ctx:         ctx,
		log:         log,
		statService: statService,
	}
}

func (h *StatHandler) TotalClients(c *gin.Context) {
	const op = "handlers.statistics.totalClients"

	log := h.log.With(
		slog.String("op", op),
	)

	total, err := h.statService.TotalClients(h.ctx)
	if err != nil {
		log.Error("failed to get total clients", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *StatHandler) NewClients(c *gin.Context) {
	const op = "handlers.statistics.newClients"

	log := h.log.With(
		slog.String("op", op),
	)

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing 'from' or 'to' date"))
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		log.Error("failed to parse 'from' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'from' date format"))
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		log.Error("failed to parse 'to' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'to' date format"))
		return
	}

	if from.After(to) {
		c.JSON(http.StatusBadRequest, response.Error("'from' date must be before 'to' date"))
		return
	}

	total, err := h.statService.NewClients(h.ctx, from, to)
	if err != nil {
		log.Error("failed to get new clients", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *StatHandler) TotalIncome(c *gin.Context) {
	const op = "handlers.statistics.totalIncome"

	log := h.log.With(
		slog.String("op", op),
	)

	total, err := h.statService.TotalIncome(h.ctx)
	if err != nil {
		log.Error("failed to get total income", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *StatHandler) Income(c *gin.Context) {
	const op = "handlers.statistics.income"

	log := h.log.With(
		slog.String("op", op),
	)

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing 'from' or 'to' date"))
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		log.Error("failed to parse 'from' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'from' date format"))
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		log.Error("failed to parse 'to' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'to' date format"))
		return
	}

	if from.After(to) {
		c.JSON(http.StatusBadRequest, response.Error("'from' date must be before 'to' date"))
		return
	}

	income, err := h.statService.Income(h.ctx, from, to)
	if err != nil {
		log.Error("failed to get income", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"income": income})
}

func (h *StatHandler) TotalSoldSubscriptions(c *gin.Context) {
	const op = "handlers.statistics.totalSoldSubscriptions"

	log := h.log.With(
		slog.String("op", op),
	)

	total, err := h.statService.TotalSoldSubscriptions(h.ctx)
	if err != nil {
		log.Error("failed to get total sold subscriptions", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *StatHandler) SoldSubscriptions(c *gin.Context) {
	const op = "handlers.statistics.soldSubscriptions"

	log := h.log.With(
		slog.String("op", op),
	)

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing 'from' or 'to' date"))
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		log.Error("failed to parse 'from' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'from' date format"))
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		log.Error("failed to parse 'to' date", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("Invalid 'to' date format"))
		return
	}

	if from.After(to) {
		c.JSON(http.StatusBadRequest, response.Error("'from' date must be before 'to' date"))
		return
	}

	total, err := h.statService.SoldSubscriptions(h.ctx, from, to)
	if err != nil {
		log.Error("failed to get sold subscriptions", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("Internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
