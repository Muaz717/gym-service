package app

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/config"
	"github.com/Muaz717/gym_app/app/internal/cron"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	personService "github.com/Muaz717/gym_app/app/internal/services/person"
	statService "github.com/Muaz717/gym_app/app/internal/services/statistics"
	subFreezeService "github.com/Muaz717/gym_app/app/internal/services/sub_freeze"
	subscriptionService "github.com/Muaz717/gym_app/app/internal/services/subscription"
	"github.com/Muaz717/gym_app/app/internal/storage/postgres"
	"github.com/Muaz717/gym_app/app/internal/storage/redis"

	httpApp "github.com/Muaz717/gym_app/app/internal/app/http"
	authService "github.com/Muaz717/gym_app/app/internal/services/auth"
	personSubService "github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"log/slog"
)

type App struct {
	HTTPSrv *httpApp.HttpApp
	Cron    *cron.CronJobs
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) *App {
	storage, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		log.Error("failed to init storage", sl.Error(err))
		panic(err)
	}

	cache, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		log.Error("failed to init cache", sl.Error(err))
		panic(err)
	}

	ssoClient, err := grpc.NewSSOClient(
		log,
		cfg.Clients.SSO.Host,
		cfg.Clients.SSO.Port,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", sl.Error(err))
		panic(err)
	}

	personSrv := personService.New(log, storage, cache, cache)
	subscriptionSrv := subscriptionService.New(log, storage)
	personSubSrv := personSubService.New(log, storage, cache, storage, cache)
	authSrv := authService.New(log, ssoClient, cfg.AppID)
	statSrv := statService.New(log, storage, cache)
	freezeSrv := subFreezeService.New(log, storage, cache)

	cr := cron.New(personSubSrv)

	httpApplication := httpApp.New(ctx, log, *cfg, ssoClient, authSrv, personSrv, subscriptionSrv, personSubSrv, statSrv, freezeSrv)

	return &App{
		HTTPSrv: httpApplication,
		Cron:    cr,
	}
}
