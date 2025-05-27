package authService

import (
	"context"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	ssov1 "github.com/Muaz717/gym_app/app/pkg/sso"

	"log/slog"
)

type SSOClient interface {
	Login(ctx context.Context, appId int32, email, password string) (string, error)
	RegisterNewUser(ctx context.Context, email, password string) (int64, error)
	CheckToken(ctx context.Context, appID int32, token string) (*ssov1.CheckTokenResponse, error)
}
type AuthService struct {
	log       *slog.Logger
	appId     int32
	ssoClient SSOClient
}

func New(
	log *slog.Logger,
	ssoClient SSOClient,
	appId int32,
) *AuthService {
	return &AuthService{
		log:       log,
		ssoClient: ssoClient,
		appId:     appId,
	}
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "services.auth.login"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("logging in", slog.String("email", email))

	token, err := a.ssoClient.Login(ctx, a.appId, email, password)
	if err != nil {
		log.Error("failed to login", slog.String("email", email), sl.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("login successful")
	return token, nil
}

func (a *AuthService) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
	const op = "services.auth.registerNewUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("registering new user", slog.String("email", email))

	userId, err := a.ssoClient.RegisterNewUser(ctx, email, password)
	if err != nil {
		log.Error("failed to register new user", slog.String("email", email), sl.Error(err))
		return 0, err
	}

	log.Info("user registered successfully")
	return userId, nil
}

func (a *AuthService) CheckToken(ctx context.Context, token string) (models.User, error) {
	const op = "services.auth.checkToken"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("checking token", slog.String("token", token))

	resp, err := a.ssoClient.CheckToken(ctx, a.appId, token)
	if err != nil {
		log.Error("failed to check token", slog.String("token", token), sl.Error(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("token is valid")
	return CheckTokenResponseToUser(resp), nil
}

func CheckTokenResponseToUser(resp *ssov1.CheckTokenResponse) models.User {
	return models.User{
		UserID: resp.UserId,
		Email:  resp.Email,
		Roles:  resp.Roles,
	}
}
