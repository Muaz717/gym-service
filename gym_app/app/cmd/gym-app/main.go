// @title       Gym API
// @version     1.0
// @description     Backend for Gym application
// @termsOfService  "https://example.com/terms"
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @contact.name	Murad Golang Backend
// @contact.email	m.azizov03@mail.ru

// @host	localhost:8082
// @BasePath /api/v1
package main

import (
	"context"
	"github.com/Muaz717/gym_app/app/internal/app"
	"github.com/Muaz717/gym_app/app/internal/config"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/handlers/slogpretty"

	_ "github.com/Muaz717/gym_app/app/docs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application")

	application := app.New(ctx, log, cfg)

	application.Cron.Start(ctx)

	log.Info("cron jobs started")
	defer application.Cron.Stop()

	go application.HTTPSrv.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	application.HTTPSrv.Stop()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envLocal:
		log = setupPrettySlog()
	}

	return log
}

func setupPrettySlog() *slog.Logger {

	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
