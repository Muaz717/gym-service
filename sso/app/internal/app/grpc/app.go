package grpcapp

import (
	"fmt"
	authgrpc "github.com/Muaz717/sso/app/internal/grpc/auth"
	"github.com/Muaz717/sso/app/internal/lib/logger/sl"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log         *slog.Logger
	gRPCServer  *grpc.Server
	authService authgrpc.AuthSrv
	port        string
	host        string
}

func New(log *slog.Logger, port string, host string, authService authgrpc.AuthSrv) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Reg(gRPCServer, authService)

	return &App{
		log:         log,
		gRPCServer:  gRPCServer,
		authService: authService,
		port:        port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("port", a.port),
	)

	l, err := net.Listen("tcp", net.JoinHostPort(a.host, a.port))
	if err != nil {
		log.Error("failed to listen", sl.Error(err))
		return fmt.Errorf("failed to listen: %s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("address", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		log.Error("failed to serve", sl.Error(err))
		return fmt.Errorf("failed to serve: %s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.String("port", a.port))

	a.gRPCServer.GracefulStop()
}
