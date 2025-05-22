package authHandler

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/grpcerrors"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/models"
	"github.com/gin-gonic/gin"

	"log/slog"
	"net/http"
	"strconv"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	RegisterNewUser(ctx context.Context, email, password string) (int64, error)
}

type AuthHandler struct {
	ctx         context.Context
	log         *slog.Logger
	authService AuthService
}

func New(
	ctx context.Context,
	log *slog.Logger,
	authService AuthService,
) *AuthHandler {
	return &AuthHandler{
		ctx:         ctx,
		log:         log,
		authService: authService,
	}
}

// Login godoc
// @Summary Login
// @Description Login
// @Security BearerAuth
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login"
// @Success 200 {object} response.Response "Login successful"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	const op = "handlers.auth.login"

	log := h.log.With(
		slog.String("op", op),
	)

	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind json", slog.String("op", op), sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
	}

	token, err := h.authService.Login(h.ctx, req.Email, req.Password)
	if err != nil {
		log.Error("failed to login", slog.String("op", op), sl.Error(err))

		prettyErr := grpcerrors.ParseValidationError(err)
		c.JSON(http.StatusInternalServerError, response.Error(prettyErr))
		return
	}

	log.Info("login successful")

	c.SetCookie("token", token, 360000, "/", "localhost", false, true)
	c.JSON(http.StatusOK, response.OK("login successful"))
}

// RegisterNewUser godoc
// @Summary Register new user
// @Description Register new user
// @Security BearerAuth
// @Tags auth
// @Accept json
// @Produce json
// @Param register body models.RegisterRequest true "Register"
// @Success 200 {object} response.Response "User registered successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 409 {object} response.Response "Conflict"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterNewUser(c *gin.Context) {
	const op = "handlers.auth.registerNewUser"

	log := h.log.With(
		slog.String("op", op),
	)

	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind json", slog.String("op", op), sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	userID, err := h.authService.RegisterNewUser(h.ctx, req.Email, req.Password)
	if err != nil {
		log.Error("failed to register new user", slog.String("op", op), sl.Error(err))

		prettyErr := grpcerrors.ParseValidationError(err)
		c.JSON(http.StatusInternalServerError, response.Error(prettyErr))
		return
	}

	log.Info("user registered successfully")

	c.JSON(http.StatusOK, response.OK(strconv.Itoa(int(userID))))
}
