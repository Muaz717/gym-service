package person_sub

import (
	"context"
	"errors"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	personSubService "github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"github.com/gin-gonic/gin"
	"strconv"

	"io"
	"log/slog"
	"net/http"
)

type PersonSubService interface {
	AddPersonSub(ctx context.Context, personSubStrDate dto.PersonSubInput) (string, error)
	GetPersonSubByNumber(ctx context.Context, number string) (dto.PersonSubResponse, error)
	GetAllPersonSubs(ctx context.Context) ([]dto.PersonSubResponse, error)
	DeletePersonSub(ctx context.Context, number string) error
	FindPersonSubByPersonName(ctx context.Context, name string) ([]dto.PersonSubResponse, error)
	FindPersonSubByPersonId(ctx context.Context, personID int) ([]dto.PersonSubResponse, error)
}

type PersonSubHandler struct {
	log              *slog.Logger
	personSubService PersonSubService
}

func New(log *slog.Logger, personSubService PersonSubService) *PersonSubHandler {
	return &PersonSubHandler{
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
// @Param        person_sub  body  models.PersonSubscription  true  "Абонемент"
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

	var personSub dto.PersonSubInput

	if err := c.ShouldBindJSON(&personSub); err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			c.JSON(http.StatusBadRequest, response.Error("empty request"))
			return
		}

		log.Error("failed to decode request body", sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	log.Info("", personSub)

	if err := personSub.Validate(); err != nil {
		log.Error("failed to validate person subscription", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	personSubNumber, err := h.personSubService.AddPersonSub(c.Request.Context(), personSub)
	if err != nil {

		if errors.Is(err, personSubService.ErrSubExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Абонемент с таким номером уже существует"})
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

	if err := h.personSubService.DeletePersonSub(c.Request.Context(), number); err != nil {

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
// @Success      200   {array}   dto.PersonSubResponse
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Router       /person_sub/find/{number} [get]
func (h *PersonSubHandler) FindPersonSubByNumber(c *gin.Context) {
	const op = "handlers.personSub.getPersonSubByNumber"

	log := h.log.With(
		slog.String("op", op),
	)

	number := c.Param("number")

	personSub, err := h.personSubService.GetPersonSubByNumber(c.Request.Context(), number)
	if err != nil {
		log.Error("failed to get person subscription by number", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to get person subscription"))
		return
	}

	c.JSON(http.StatusOK, personSub)
}

// FindAllPersonSubs godoc
// @Summary      Получить все абонементы
// @Description  Возвращает список всех абонементов
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Success      200   {array}    dto.PersonSubResponse
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub [get]
func (h *PersonSubHandler) FindAllPersonSubs(c *gin.Context) {
	const op = "handlers.personSub.getAllPersonSubs"

	log := h.log.With(
		slog.String("op", op),
	)

	personSubs, err := h.personSubService.GetAllPersonSubs(c.Request.Context())
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
// @Success      200   {array}   dto.PersonSubResponse
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Клиент не найден"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub/find_by_name [get]
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

	personSubs, err := h.personSubService.FindPersonSubByPersonName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, personSubService.ErrSubNotFound) {
			log.Info("no subscriptions found for this person", slog.String("name", name))
			// Возвращаем пустой массив и 200 OK (это не ошибка для фронта!)
			c.JSON(http.StatusOK, []interface{}{})
			return
		}

		if errors.Is(err, personSubService.ErrPersonNotFound) {
			log.Error("person with this name not found", sl.Error(err))
			c.JSON(http.StatusNotFound, response.Error("person with this name not found"))
			return
		}

		log.Error("failed to find person subscription by person name", sl.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("failed to find person subscription"))
		return
	}

	log.Info("person subscription found by person name", slog.String("name", name))
	c.JSON(http.StatusOK, personSubs)
}

// FindPersonSubByPersonId godoc
// @Summary      Получить абонементы по ID клиента
// @Description  Возвращает список абонементов клиента по ID
// @Security BearerAuth
// @Tags         person_sub
// @Accept       json
// @Produce      json
// @Param        id  path     int  true  "ID клиента"
// @Success      200   {array}    dto.PersonSubResponse
// @Failure      400   {object}  response.Response "Ошибка валидации"
// @Failure      404   {object}  response.Response "Клиент не найден"
// @Failure      500   {object}  response.Response "Внутренняя ошибка сервера"
// @Router       /person_sub/find/{id} [get]
func (h *PersonSubHandler) FindPersonSubByPersonId(c *gin.Context) {
	personIDStr := c.Param("id")
	if personIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id is required"})
		return
	}

	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id"})
		return
	}

	subs, err := h.personSubService.FindPersonSubByPersonId(c.Request.Context(), personID)
	if err != nil {
		// Если человек не найден — возвращаем 404.
		if errors.Is(err, personSubService.ErrPersonNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "person with that id not found"})
			return
		}
		// Если у человека нет абонементов — возвращаем пустой массив (OK).
		if errors.Is(err, personSubService.ErrSubNotFound) {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		// Другие ошибки — 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subs)
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
