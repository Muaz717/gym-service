package person_sub

import (
	"context"
	"errors"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/models"
	personSubService "github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"github.com/gin-gonic/gin"

	"io"
	"log/slog"
	"net/http"
)

type PersonSubService interface {
	AddPersonSub(ctx context.Context, personSubStrDate models.PersonSubStrDate) (string, error)
	GetPersonSubByNumber(ctx context.Context, number string) (models.PersonSubStrDate, error)
	GetAllPersonSubs(ctx context.Context) ([]models.PersonSubStrDate, error)
	DeletePersonSub(ctx context.Context, number string) error
	FindPersonSubByPersonName(ctx context.Context, name string) ([]models.PersonSubStrDate, error)
	//UpdatePersonSub(ctx context.Context, number string, personSubStrDate models.PersonSubStrDate) error
}

type PersonSubHandler struct {
	ctx              context.Context
	log              *slog.Logger
	personSubService PersonSubService
}

func New(ctx context.Context, log *slog.Logger, personSubService PersonSubService) *PersonSubHandler {
	return &PersonSubHandler{
		ctx:              ctx,
		log:              log,
		personSubService: personSubService,
	}
}

// AddPersonSub godoc
// @Summary      Добавить абонемент
// @Description  Добавляет новый абонемент
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Param        person_sub  body     models.PersonSubStrDate  true  "Абонемент"
// @Success      200   {object}  response.Response "Абонемент добавлен"
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      409   {object}  response.Response "Конфликт"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub/add [post]
func (h *PersonSubHandler) AddPersonSub(c *gin.Context) {

	const op = "handlers.personSub.addPersonSub"

	log := h.log.With(
		slog.String("op", op),
	)

	var personSubStrDate models.PersonSubStrDate

	if err := c.ShouldBindJSON(&personSubStrDate); err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			c.JSON(http.StatusBadRequest, response.Error("empty request"))
			return
		}

		log.Error("failed to decode request body", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	if err := personSubStrDate.Validate(); err != nil {
		log.Error("failed to validate person subscription", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	personSubNumber, err := h.personSubService.AddPersonSub(h.ctx, personSubStrDate)
	if err != nil {

		if errors.Is(err, personSubService.ErrSubExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "subscription with that number already exists"})
			return
		}

		if errors.Is(err, personSubService.ErrPersonNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "person with this id not found"})
			return
		}

		log.Error("failed to add person subscription", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to add person subscription"))
		return
	}

	log.Info("person subscription registered", "number", personSubNumber)
	c.JSON(http.StatusOK, response.OK(personSubNumber))

}

// DeletePersonSub godoc
// @Summary      Удалить абонемент
// @Description  Удаляет абонемент по номеру
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Param        number  path     string  true  "Номер абонемента"
// @Success      200   {object}  response.Response "Абонемент удален"
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Абонемент не найден"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub/delete/{number} [delete]
func (h *PersonSubHandler) DeletePersonSub(c *gin.Context) {
	const op = "handlers.personSub.deletePersonSub"

	log := h.log.With(
		slog.String("op", op),
	)

	number := c.Param("number")

	if err := h.personSubService.DeletePersonSub(h.ctx, number); err != nil {

		if errors.Is(err, personSubService.ErrSubNotFound) {
			log.Error("subscription not found", sl.Error(err))
			c.JSON(http.StatusNotFound, response.Error("subscription not found"))
			return
		}

		log.Error("failed to delete person subscription", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to delete person subscription"))
		return
	}

	log.Info("person subscription deleted", "number", number)
	c.JSON(http.StatusOK, response.OK("person subscription deleted"))
}

// FindPersonSubByNumber godoc
// @Summary      Получить абонементы по номеру
// @Description  Возвращает список абонементов клиента по номеру
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Param        number  path     string  true  "Номер абонемента"
// @Success      200   {array}   models.PersonSubscription
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Router       /person_sub/find/{number} [get]
func (h *PersonSubHandler) FindPersonSubByNumber(c *gin.Context) {
	const op = "handlers.personSub.getPersonSubByNumber"

	log := h.log.With(
		slog.String("op", op),
	)

	number := c.Param("number")

	personSubStrDate, err := h.personSubService.GetPersonSubByNumber(h.ctx, number)
	if err != nil {
		log.Error("failed to get person subscription by number", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to get person subscription"))
		return
	}

	c.JSON(http.StatusOK, personSubStrDate)
}

// FindAllPersonSubs godoc
// @Summary      Получить все абонементы
// @Description  Возвращает список всех абонементов
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Success      200   {array}   models.PersonSubStrDate
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub [get]
func (h *PersonSubHandler) FindAllPersonSubs(c *gin.Context) {
	const op = "handlers.personSub.getAllPersonSubs"

	log := h.log.With(
		slog.String("op", op),
	)

	personSubs, err := h.personSubService.GetAllPersonSubs(h.ctx)
	if err != nil {
		log.Error("failed to get all person subscriptions", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to get all person subscriptions"))
		return
	}

	log.Info("all person subscriptions found")
	c.JSON(http.StatusOK, personSubs)
}

// FindPersonSubByPersonName godoc
// @Summary      Получить абонементы по имени
// @Description  Возвращает список абонементов клиента по имени
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Param        name  query     string  true  "Имя клиента"
// @Success      200   {array}   models.PersonSubStrDate
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Абонемент не найден"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub/find [get]
func (h *PersonSubHandler) FindPersonSubByPersonName(c *gin.Context) {
	const op = "handlers.personSub.getPersonSubByPersonName"

	log := h.log.With(
		slog.String("op", op),
	)

	name := c.Query("name")
	if name == "" {
		log.Error("name parameter is missing")
		c.JSON(http.StatusBadRequest, response.Error("name parameter is required"))
		return
	}

	personSubs, err := h.personSubService.FindPersonSubByPersonName(h.ctx, name)
	if err != nil {
		log.Error("failed to find person subscription by person name", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to find person subscription"))
		return
	}

	log.Info("person subscription found by person name", slog.String("name", name))
	c.JSON(http.StatusOK, personSubs)
}

//func (h *PersonSubHandler) UpdatePersonSub(c *gin.Context) {
//	const op = "handlers.personSub.UpdatePersonSub"
//
//	log := h.log.With(
//		slog.String("op", op),
//	)
//
//	number := c.Param("number")
//	if number == "" {
//		log.Error("number parameter is missing")
//		c.JSON(http.StatusBadRequest, response.Error("number parameter is required"))
//		return
//	}
//
//	var personSubStrDate models.PersonSubStrDate
//	if err := c.ShouldBindJSON(&personSubStrDate); err != nil {
//		if errors.Is(err, io.EOF) {
//			log.Error("request body is empty")
//			c.JSON(http.StatusBadRequest, response.Error("empty request"))
//			return
//		}
//
//		log.Error("failed to decode request body", sl.Error(err))
//		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
//		return
//	}
//
//	err := h.personSubService.UpdatePersonSub(h.ctx, number, personSubStrDate)
//	if err != nil {
//		if errors.Is(err, personSubService.ErrSubNotFound) {
//			log.Error("subscription not found", sl.Error(err))
//			c.JSON(http.StatusNotFound, response.Error("subscription not found"))
//			return
//		}
//
//		log.Error("failed to update person subscription", sl.Error(err))
//		c.JSON(http.StatusInternalServerError, response.Error("failed to update person subscription"))
//		return
//	}
//
//	log.Info("person subscription updated", "number", number)
//	c.JSON(http.StatusOK, response.OK("person subscription updated"))
//
//}
