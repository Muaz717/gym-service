package httpApp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/config"
	authHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/auth"
	personHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person"
	personSubHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person_sub"
	singleVisitHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/single_visit"
	statHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/statistics"
	subFreezeHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/sub_freeze"
	subscriptionHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/subscription"
	authMiddleware "github.com/Muaz717/gym_app/app/internal/http/middleware/auth"
	loggerMiddleware "github.com/Muaz717/gym_app/app/internal/http/middleware/logger"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	userRole  = "user"
	adminRole = "admin"
)

type HttpApp struct {
	HTTPServer *http.Server
	engine     *gin.Engine
	log        *slog.Logger
	cfg        config.Config
}

func New(
	log *slog.Logger,
	cfg config.Config,
	ssoClient *grpc.SSOClient,
	authService authHandler.AuthService,
	personService personHandler.PersonService,
	subscriptionService subscriptionHandler.SubscriptionService,
	personSubService personSubHandler.PersonSubService,
	statService statHandler.StatService,
	subFreezeService subFreezeHandler.SubFreezeService,
	singleVisitService singleVisitHandler.SingleVisitService,
) *HttpApp {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	setupMiddleware(engine, log, cfg)

	if cfg.Env == "dev" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	userMiddleware := authMiddleware.AuthMiddleware(log, ssoClient, cfg.AppID, userRole)
	adminMiddleware := authMiddleware.AuthMiddleware(log, ssoClient, cfg.AppID, adminRole)

	api := engine.Group("/api/v1")

	authHandle := authHandler.New(log, authService)
	personHandle := personHandler.New(log, personService)
	subscriptionHandle := subscriptionHandler.New(log, subscriptionService)
	personSubHandle := personSubHandler.New(log, personSubService)
	statHandle := statHandler.New(log, statService)
	freezeHandle := subFreezeHandler.New(log, subFreezeService)
	singleVisitHandle := singleVisitHandler.New(log, singleVisitService)

	// --- Auth routes ---
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandle.RegisterNewUser)
		auth.POST("/login", authHandle.Login)
		auth.GET("/me", authHandle.Me)
	}

	api.Use(userMiddleware)
	{
		// --- User routes ---
		registerPersonRoutes(api, personHandle, adminMiddleware)
		// --- Subscription routes ---
		registerSubscriptionRoutes(api, subscriptionHandle, adminMiddleware)
		// --- Person Subscription routes ---
		registerPersonSubRoutes(api, personSubHandle, adminMiddleware)
		// --- Freeze routes ---
		registerFreezeRoutes(api, freezeHandle, adminMiddleware)
		// --- Single Visit routes ---
		registerSingleVisitRoutes(api, singleVisitHandle, adminMiddleware)
		// --- Statistics routes ---
		registerStatRoutes(api, statHandle)
	}

	srv := &http.Server{
		Addr:         net.JoinHostPort(cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &HttpApp{
		HTTPServer: srv,
		engine:     engine,
		log:        log,
		cfg:        cfg,
	}
}

func (a *HttpApp) Run() error {
	const op = "httpApp.Run"

	log := a.log.With(slog.String("op", op))
	log.Info("HTTP server is starting", slog.String("addr", a.cfg.HTTPServer.Port))

	if err := a.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to run http server", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *HttpApp) Stop(ctx context.Context) error {
	const op = "httpApp.Stop"

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	a.log.With(slog.String("op", op)).
		Info("stopping HTTP server", slog.String("addr", a.HTTPServer.Addr))

	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to gracefully shutdown server", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func setupMiddleware(engine *gin.Engine, log *slog.Logger, cfg config.Config) {
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:80", "http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	engine.Use(cors.New(corsConfig))
	engine.Use(gin.Recovery())
	engine.Use(loggerMiddleware.New(log))
}
