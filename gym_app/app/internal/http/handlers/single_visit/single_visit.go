package singleVisitHandler

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
)

type SingleVisitService interface {
	AddSingleVisit(ctx context.Context, singleVisStrDate dto.SingleVisitInput) error
	GetAllSingleVisits(ctx context.Context) ([]models.SingleVisit, error)
	GetSingleVisitById(ctx context.Context, id int) (models.SingleVisit, error)
	GetSingleVisitsByDay(ctx context.Context, date string) ([]models.SingleVisit, error)
	GetSingleVisitsByPeriod(ctx context.Context, from, to string) ([]models.SingleVisit, error)
	DeleteSingleVisit(ctx context.Context, id int) error
}

type SingleVisitHandler struct {
	log                *slog.Logger
	singleVisitService SingleVisitService
}

func New(
	log *slog.Logger,
	singleVisitService SingleVisitService,
) *SingleVisitHandler {
	return &SingleVisitHandler{
		log:                log,
		singleVisitService: singleVisitService,
	}
}

func (h *SingleVisitHandler) AddSingleVisit(c *gin.Context) {
	const op = "handlers.single_visit.AddSingleVisit"
	log := h.log.With(slog.String("op", op))

	var req dto.SingleVisitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.singleVisitService.AddSingleVisit(c.Request.Context(), req); err != nil {
		log.Error("failed to add single visit", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "single visit added successfully"})
}

func (h *SingleVisitHandler) GetAllSingleVisits(c *gin.Context) {
	const op = "handlers.single_visit.GetAllSingleVisits"
	log := h.log.With(slog.String("op", op))

	singleVisits, err := h.singleVisitService.GetAllSingleVisits(c.Request.Context())
	if err != nil {
		log.Error("failed to get all single visits", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"single_visits": singleVisits})
}

func (h *SingleVisitHandler) GetSingleVisitById(c *gin.Context) {
	const op = "handlers.single_visit.GetSingleVisitById"
	log := h.log.With(slog.String("op", op))

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'id' path parameter"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("failed to parse id", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	singleVisit, err := h.singleVisitService.GetSingleVisitById(c.Request.Context(), id)
	if err != nil {
		log.Error("failed to get single visit by id", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"single_visit": singleVisit})
}

func (h *SingleVisitHandler) GetSingleVisitsByDay(c *gin.Context) {
	const op = "handlers.single_visit.GetSingleVisitsByDay"
	log := h.log.With(slog.String("op", op))

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'date' query parameter"})
		return
	}

	singleVisits, err := h.singleVisitService.GetSingleVisitsByDay(c.Request.Context(), dateStr)
	if err != nil {
		log.Error("failed to get single visits by day", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"single_visits": singleVisits})
}

func (h *SingleVisitHandler) GetSingleVisitsByPeriod(c *gin.Context) {
	const op = "handlers.single_visit.GetSingleVisitsByPeriod"
	log := h.log.With(slog.String("op", op))

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'from' or 'to' query parameter"})
		return
	}

	singleVisits, err := h.singleVisitService.GetSingleVisitsByPeriod(c.Request.Context(), fromStr, toStr)
	if err != nil {
		log.Error("failed to get single visits by period", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"single_visits": singleVisits})
}

func (h *SingleVisitHandler) DeleteSingleVisit(c *gin.Context) {
	const op = "handlers.single_visit.DeleteSingleVisit"
	log := h.log.With(slog.String("op", op))

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'id' path parameter"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("failed to parse id", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := h.singleVisitService.DeleteSingleVisit(c.Request.Context(), id); err != nil {
		log.Error("failed to delete single visit", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "single visit deleted successfully"})
}
