package app

import (
	"context"

	"log/slog"

	"github.com/Muaz717/gym_app/app/internal/app/http"
	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/config"
	"github.com/Muaz717/gym_app/app/internal/cron"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"

	"github.com/Muaz717/gym_app/app/internal/services/auth"
	"github.com/Muaz717/gym_app/app/internal/services/person"
	"github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"github.com/Muaz717/gym_app/app/internal/services/single_visit"
	"github.com/Muaz717/gym_app/app/internal/services/statistics"
	"github.com/Muaz717/gym_app/app/internal/services/sub_freeze"
	"github.com/Muaz717/gym_app/app/internal/services/subscription"

	"github.com/Muaz717/gym_app/app/internal/storage/postgres"
	"github.com/Muaz717/gym_app/app/internal/storage/redis"
)

type App struct {
	HTTPSrv *httpApp.HttpApp
	Cron    *cron.CronJobs
}

func New(ctx context.Context, log *slog.Logger, cfg config.Config) *App {
	// --- Init Postgres ---
	storage, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		log.Error("failed to init storage", sl.Error(err))
		panic(err)
	}

	// --- Init Redis ---
	cache, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		log.Error("failed to init cache", sl.Error(err))
		panic(err)
	}

	// --- Init SSO Client ---
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

	// --- Init Services ---
	personSrv := personService.New(log, storage, cache, cache)
	subscriptionSrv := subscriptionService.New(log, storage)
	personSubSrv := personSubService.New(log, storage, cache, storage, cache)
	authSrv := authService.New(log, ssoClient, cfg.AppID)
	statSrv := statistics.New(log, storage, cache)
	freezeSrv := subFreezeService.New(log, storage, cache)
	singleVisitSrv := singleVisitService.New(log, storage, cache)

	// --- Init Cron ---
	cronJobs := cron.New(personSubSrv)

	// --- Init HTTP App ---
	httpSrv := httpApp.New(
		log,
		cfg,
		ssoClient,
		authSrv,
		personSrv,
		subscriptionSrv,
		personSubSrv,
		statSrv,
		freezeSrv,
		singleVisitSrv,
	)

	return &App{
		HTTPSrv: httpSrv,
		Cron:    cronJobs,
	}
}
