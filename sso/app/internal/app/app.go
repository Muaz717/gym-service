package app

import (
	"context"
	grpcapp "github.com/Muaz717/sso/app/internal/app/grpc"
	"github.com/Muaz717/sso/app/internal/config"
	"github.com/Muaz717/sso/app/internal/services/auth"
	"github.com/Muaz717/sso/app/internal/storage/postgres"
	"log/slog"

	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort string,
	grpcHost string,
	db config.DBConfig,
	tokenTTL time.Duration,
) *App {
	storage, err := postgres.New(context.Background(), db)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, grpcPort, grpcHost, authService)

	return &App{
		GRPCSrv: grpcApp,
	}
}
