package httpApp

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/config"
	authHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/auth"
	personHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person"
	personSubHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person_sub"
	statHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/statistics"
	subscriptionHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/subscription"
	authMiddleware "github.com/Muaz717/gym_app/app/internal/http/middleware/auth"
	loggerMiddleware "github.com/Muaz717/gym_app/app/internal/http/middleware/logger"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	userRole  = "user"
	adminRole = "admin"
)

type HttpApp struct {
	HTTPServer *http.Server
	engine     *gin.Engine
	ctx        context.Context
	log        *slog.Logger
	cfg        config.Config
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg config.Config,
	ssoClient *grpc.SSOClient,
	authService authHandler.AuthService,
	personService personHandler.PersonService,
	subscriptionService subscriptionHandler.SubscriptionService,
	personSubService personSubHandler.PersonSubService,
	statService statHandler.StatService,
) *HttpApp {

	personHandle := personHandler.New(ctx, log, personService)
	subscriptionHandle := subscriptionHandler.New(ctx, log, subscriptionService)
	personSubHandle := personSubHandler.New(ctx, log, personSubService)
	authHandle := authHandler.New(ctx, log, authService)
	statHandle := statHandler.New(ctx, log, statService)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	setupMiddleware(engine, log, cfg)

	userMiddleware := authMiddleware.AuthMiddleware(log, ssoClient, cfg.AppID, userRole)
	adminMiddleware := authMiddleware.AuthMiddleware(log, ssoClient, cfg.AppID, adminRole)

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandle.RegisterNewUser)
		auth.POST("/login", authHandle.Login)
		auth.GET("/me", authHandle.Me)
	}

	api.Use(userMiddleware)
	{
		people := api.Group("/people")
		{
			people.GET("", personHandle.FindAllPeople)
			people.GET("/find", personHandle.FindPersonByName)
			people.GET("/find/:id", personHandle.FindPersonById)

			adminPeople := people.Group("")
			adminPeople.Use(adminMiddleware)
			adminPeople.POST("/add", personHandle.AddPerson)
			adminPeople.PUT("update/:id", personHandle.UpdatePerson)
			adminPeople.DELETE("delete/:id", personHandle.DeletePerson)
		}

		subscription := api.Group("/subscription")
		{
			subscription.GET("", subscriptionHandle.FindAllSubscriptions)

			adminSubscription := subscription.Group("")
			adminSubscription.Use(adminMiddleware)
			adminSubscription.POST("/add", subscriptionHandle.AddSubscription)
			adminSubscription.PUT("update/:id", subscriptionHandle.UpdateSubscription)
			adminSubscription.DELETE("delete/:id", subscriptionHandle.DeleteSubscription)
		}

		personSub := api.Group("/person_sub")
		{
			personSub.GET("find/:number", personSubHandle.FindPersonSubByNumber)
			personSub.GET("", personSubHandle.FindAllPersonSubs)
			personSub.GET("/find", personSubHandle.FindPersonSubByPersonName)
			personSub.GET("/find/id/:id", personSubHandle.FindPersonSubByPersonId)

			adminPersonSub := personSub.Group("")
			adminPersonSub.Use(adminMiddleware)
			adminPersonSub.POST("/add", personSubHandle.AddPersonSub)
			adminPersonSub.DELETE("delete/:number", personSubHandle.DeletePersonSub)
		}

		// --- STATISTICS ROUTES ---
		statistics := api.Group("/statistics")
		{
			statistics.GET("/total_clients", statHandle.TotalClients)
			statistics.GET("/new_clients", statHandle.NewClients)
			statistics.GET("/total_income", statHandle.TotalIncome)
			statistics.GET("/income", statHandle.Income)
			statistics.GET("/total_sold_subscriptions", statHandle.TotalSoldSubscriptions)
			statistics.GET("/sold_subscriptions", statHandle.SoldSubscriptions)
		}
		// --- END STATISTICS ROUTES ---
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
		ctx:        ctx,
		log:        log,
		cfg:        cfg,
	}
}

func (a *HttpApp) Run() error {
	const op = "httpApp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("addr", a.cfg.HTTPServer.Port),
	)

	log.Info("HTTP server is starting", slog.String("addr", a.cfg.HTTPServer.Port))

	if err := a.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to run http server", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *HttpApp) Stop() error {
	const op = "httpApp.Stop"

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
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
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:80",
			"http://localhost",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	engine.Use(cors.New(corsConfig))

	engine.Use(gin.Recovery())
	engine.Use(loggerMiddleware.New(log))
}
