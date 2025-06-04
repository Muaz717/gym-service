package authHandler

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/lib/api/response"
	"github.com/Muaz717/gym_app/app/internal/lib/grpcerrors"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/gin-gonic/gin"

	"log/slog"
	"net/http"
	"strconv"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	RegisterNewUser(ctx context.Context, email, password string) (int64, error)
	CheckToken(ctx context.Context, token string) (dto.User, error)
}

type AuthHandler struct {
	log         *slog.Logger
	authService AuthService
}

func New(
	log *slog.Logger,
	authService AuthService,
) *AuthHandler {
	return &AuthHandler{
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

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind json", slog.String("op", op), sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
	}

	token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("failed to login", slog.String("op", op), sl.Error(err))

		prettyErr := grpcerrors.ParseValidationError(err)
		c.JSON(http.StatusInternalServerError, response.Error(prettyErr))
		return
	}

	log.Info("login successful")

	c.SetCookie("token", token, 360000, "/", "", false, true)
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

	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to bind json", slog.String("op", op), sl.Error(err))
		c.JSON(http.StatusBadRequest, response.Error("failed to decode request"))
		return
	}

	userID, err := h.authService.RegisterNewUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("failed to register new user", slog.String("op", op), sl.Error(err))

		prettyErr := grpcerrors.ParseValidationError(err)
		c.JSON(http.StatusInternalServerError, response.Error(prettyErr))
		return
	}

	log.Info("user registered successfully")

	c.JSON(http.StatusOK, response.OK(strconv.Itoa(int(userID))))
}

// Me godoc
// @Summary Get current user info
// @Description Returns info about the authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} models.User "Current user info"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /auth/me [get]
// @Security BearerAuth
func (h *AuthHandler) Me(c *gin.Context) {
	cookie, err := c.Request.Cookie("token")
	if err != nil || cookie.Value == "" {
		c.JSON(http.StatusUnauthorized, response.Error("unauthorized"))
		return
	}
	token := cookie.Value

	user, err := h.authService.CheckToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("invalid token"))
		return
	}

	c.JSON(http.StatusOK, struct {
		UserID int64    `json:"user_id"`
		Email  string   `json:"email"`
		Roles  []string `json:"roles"`
	}{
		UserID: user.UserID,
		Email:  user.Email,
		Roles:  user.Roles,
	})
}
