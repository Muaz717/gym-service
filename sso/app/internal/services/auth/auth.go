package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/sso/app/internal/domain/models"
	"github.com/Muaz717/sso/app/internal/lib/jwt"
	"github.com/Muaz717/sso/app/internal/lib/logger/sl"
	"github.com/Muaz717/sso/app/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
		role string,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, []string, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	Logout(ctx context.Context, email string) error
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

const userRole = "user"

// New creates a new Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (token string, err error) {

	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login")

	user, roles, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Error(err))

			return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", sl.Error(err))

		return "", fmt.Errorf("%s : %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid password", sl.Error(err))

		return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Error(err))

			return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get app", sl.Error(err))

		return "", fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err = jwt.NewToken(user, app, a.tokenTTL, roles)
	if err != nil {
		log.Error("failed to generate token", sl.Error(err))

		return "", fmt.Errorf("%s : %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash password", sl.Error(err))

		return 0, fmt.Errorf("%s : %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash, userRole)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Error(err))

			return 0, fmt.Errorf("%s : %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Error(err))

		return 0, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user registered", slog.Int64("userID", id))

	return id, err
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Error(err))

			return false, fmt.Errorf("%s : %w", op, ErrInvalidAppID)
		}

		log.Error("failed to check if user is admin", sl.Error(err))

		return false, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user is admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) Logout(ctx context.Context, token string, appID int32) error {
	const op = "auth.Logout"

	log := a.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("logging out user")

	app, err := a.appProvider.App(ctx, int(appID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Error(err))

			return fmt.Errorf("%s : %w", op, ErrInvalidAppID)
		}

		log.Error("failed to get app", sl.Error(err))

		return fmt.Errorf("%s : %w", op, err)
	}

	claims, err := jwt.ParseToken(token, app)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	err = a.userProvider.Logout(ctx, claims.Email)
	if err != nil {
		log.Error("failed to logout user", sl.Error(err))

		return fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user logged out successfully")

	return nil
}

func (a *Auth) CheckToken(ctx context.Context, token string, appID int32) (*jwt.Claims, error) {
	const op = "auth.CheckToken"

	log := a.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("checking token")

	app, err := a.appProvider.App(ctx, int(appID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Error(err))

			return nil, fmt.Errorf("%s : %w", op, ErrInvalidAppID)
		}

		log.Error("failed to get app", sl.Error(err))

		return nil, fmt.Errorf("%s : %w", op, err)
	}

	claims, err := jwt.ParseToken(token, app)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("token is valid")

	return claims, nil
}
