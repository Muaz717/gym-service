package grpc

import (
	"context"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	ssov1 "github.com/Muaz717/gym_app/app/pkg/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"

	"time"
)

type SSOClient struct {
	api  ssov1.AuthClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func NewSSOClient(
	log *slog.Logger,
	host string,
	port string,
	timeout time.Duration,
	retriesCount int,
) (*SSOClient, error) {
	const op = "sso.grpc.NewClient"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		//grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	addr := net.JoinHostPort(host, port)

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcretry.UnaryClientInterceptor(retryOpts...),
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
		),
	)

	if err != nil {
		log.Error("failed to create grpc client", slog.String("op", op), sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &SSOClient{
		api:  ssov1.NewAuthClient(cc),
		conn: cc,
		log:  log,
	}, nil
}

func (c *SSOClient) Close() error {
	return c.conn.Close()
}

func (c *SSOClient) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
	const op = "sso.grpc.RegisterNewUser"

	log := c.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		log.Error("failed to register new user", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if resp.GetUserId() == 0 {
		log.Error("empty user id received")
		return 0, fmt.Errorf("%s: empty user id received", op)
	}
	log.Info("user registered successfully")
	return resp.GetUserId(), nil
}

func (c *SSOClient) Login(ctx context.Context, appID int32, email, password string) (string, error) {
	const op = "sso.grpc.Login"

	log := c.log.With(
		slog.String("op", op),
		slog.String("login", email),
	)

	log.Info("logging in")

	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
		AppId:    appID,
		Email:    email,
		Password: password,
	})

	if err != nil {
		log.Error("failed to login", sl.Error(err))
		return "", err
	}

	if resp.GetToken() == "" {
		log.Error("empty token received")
		return "", fmt.Errorf("%s: empty token received", op)
	}
	log.Info("login successful")
	return resp.GetToken(), nil
}

func (c *SSOClient) CheckToken(ctx context.Context, appID int32, token string) (*ssov1.CheckTokenResponse, error) {
	const op = "sso.grpc.CheckToken"

	log := c.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("checking token")

	resp, err := c.api.CheckToken(ctx, &ssov1.CheckTokenRequest{
		AppId: appID,
		Token: token,
	})
	if err != nil {
		log.Error("failed to check token", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
