package membershipHandler

import (
	"context"
	"errors"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	personSubService "github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"github.com/gin-gonic/gin"

	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type SubscriptionService interface {
	AddSubscription(ctx context.Context, subscription models.Subscription) (int, error)
	FindAllSubscriptions(ctx context.Context) ([]models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription models.Subscription, subID int) (int, error)
	DeleteSubscription(ctx context.Context, subID int) error
}

type SubscriptionHandler struct {
	ctx                 context.Context
	log                 *slog.Logger
	subscriptionService SubscriptionService
}

func New(
	ctx context.Context,
	log *slog.Logger,
	subscriptionService SubscriptionService,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		ctx:                 ctx,
		log:                 log,
		subscriptionService: subscriptionService,
	}
}

// AddSubscription godoc
// @Summary      Добавить абонемент
// @Description  Добавляет новый абонемент
// @Security BearerAuth
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        subscription  body     models.Subscription  true  "Абонемент"
// @Success      200   {object}  response.Response "Абонемент добавлен"
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      409   {object}  response.Response "Конфликт"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /subscription/add [post]
func (h *SubscriptionHandler) AddSubscription(c *gin.Context) {
	const op = "handlers.subscription.addSubscription"

	log := h.log.With(
		slog.String("op", op),
	)

	var subscription models.Subscription

	if err := c.ShouldBindJSON(&subscription); err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			c.JSON(http.StatusBadRequest, response.Error("empty request"))
			return
		}

		log.Error("failed to decode request body", sl.Error(err))

		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	subId, err := h.subscriptionService.AddSubscription(h.ctx, subscription)
	if err != nil {
		log.Error("failed to add subscription", sl.Error(err))

		c.JSON(http.StatusInternalServerError, response.Error("failed to add subscription"))
		return
	}

	log.Info("Subscription added", slog.Int("Subscription_id", subId))
	c.JSON(http.StatusOK, response.OK("Subscription added"))
}

// UpdateSubscription godoc
// @Summary      Обновить абонемент
// @Description  Обновляет существующий абонемент
// @Security BearerAuth
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        id           path     int                  true  "ID абонемента"
// @Param        subscription body     models.Subscription  true  "Абонемент"
// @Success      200   {object}  response.Response "Абонемент обновлен"
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Не найдено"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /subscription/update/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	const op = "handlers.subscription.updateSubscription"

	log := h.log.With(
		slog.String("op", op),
	)

	subscriptionIdStr := c.Param("id")
	if subscriptionIdStr == "" {
		log.Error("subscription ID is empty")
		c.JSON(http.StatusBadRequest, response.Error("subscription ID is required"))
		return
	}
	subscriptionID, err := strconv.Atoi(subscriptionIdStr)
	if err != nil {
		log.Error("failed to parse subscription ID", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("invalid subscription ID"))
		return
	}

	var subscription models.Subscription

	if err := c.ShouldBindJSON(&subscription); err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			c.JSON(http.StatusBadRequest, response.Error("empty request"))
			return
		}

		log.Error("failed to decode request body", sl.Error(err))

		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	subId, err := h.subscriptionService.UpdateSubscription(h.ctx, subscription, subscriptionID)
	if err != nil {
		log.Error("failed to update subscription", sl.Error(err))

		c.JSON(http.StatusInternalServerError, response.Error("failed to update subscription"))
		return
	}

	log.Info("Subscription updated", slog.Int("Subscription_id", subId))
	c.JSON(http.StatusOK, response.OK("Subscription updated"))
}

// DeleteSubscription godoc
// @Summary      Удалить абонемент
// @Description  Удаляет абонемент по номеру
// @Security BearerAuth
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        id  path     int  true  "ID абонемента"
// @Success      200   {object}  response.Response "Абонемент удален"
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Абонемент не найден"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /subscription/delete/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	const op = "handlers.subscription.deleteSubscription"

	log := h.log.With(
		slog.String("op", op),
	)

	subscriptionIdStr := c.Param("id")
	if subscriptionIdStr == "" {
		log.Error("subscription ID is empty")
		c.JSON(http.StatusBadRequest, response.Error("subscription ID is required"))
		return
	}
	subscriptionID, err := strconv.Atoi(subscriptionIdStr)
	if err != nil {
		log.Error("failed to parse subscription ID", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("invalid subscription ID"))
		return
	}

	err = h.subscriptionService.DeleteSubscription(h.ctx, subscriptionID)
	if err != nil {

		if errors.Is(err, personSubService.ErrSubNotFound) {
			c.JSON(http.StatusNotFound, response.Error("subscription not found"))
			return
		}

		log.Error("failed to delete subscription", sl.Error(err))

		c.JSON(http.StatusInternalServerError, response.Error("failed to delete subscription"))
		return
	}

	log.Info("Subscription deleted")
	c.JSON(http.StatusOK, response.OK("Subscription deleted"))
}

// FindAllSubscriptions godoc
// @Summary      Получить все абонементы
// @Description  Возвращает список всех абонементов
// @Security BearerAuth
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Success      200   {object}  []models.Subscription "Список абонементов"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /subscription [get]
func (h *SubscriptionHandler) FindAllSubscriptions(c *gin.Context) {
	const op = "handlers.subscription.findAllSubscriptions"

	log := h.log.With(
		slog.String("op", op),
	)

	subscriptions, err := h.subscriptionService.FindAllSubscriptions(h.ctx)
	if err != nil {
		log.Error("failed to get Subscriptions", sl.Error(err))

		c.JSON(http.StatusInternalServerError, response.Error("failed to get Subscriptions"))
		return
	}

	log.Info("Subscription found")

	c.JSON(http.StatusOK, subscriptions)
}
