package auth

import (
	"context"
	"errors"
	"github.com/Muaz717/sso/app/internal/lib/jwt"
	"github.com/Muaz717/sso/app/internal/lib/validation"
	"github.com/Muaz717/sso/app/internal/services/auth"
	"github.com/Muaz717/sso/app/internal/storage"
	ssov1 "github.com/Muaz717/sso/app/pkg/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthSrv interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appId int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	Logout(ctx context.Context, token string, appID int32) error
	CheckToken(ctx context.Context, token string, appID int32) (*jwt.Claims, error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth AuthSrv
}

func Reg(gRPC *grpc.Server, auth AuthSrv) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverApi) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	if err := validation.ValidateLoginInput(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverApi) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {

	if err := validation.ValidateRegisterInput(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverApi) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {

	if err := validation.ValidateIsAdminInput(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverApi) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {

	if err := validation.ValidateLogoutInput(req); err != nil {
		return nil, err
	}

	if err := s.auth.Logout(ctx, req.GetToken(), req.GetAppId()); err != nil {
		//if errors.Is(err, auth.ErrInvalidToken) {
		//	return nil, status.Error(codes.InvalidArgument, err.Error())
		//}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LogoutResponse{Success: true}, nil
}

func (s *serverApi) CheckToken(ctx context.Context, req *ssov1.CheckTokenRequest) (*ssov1.CheckTokenResponse, error) {

	if err := validation.ValidateCheckTokenInput(req); err != nil {
		return nil, err
	}

	claims, err := s.auth.CheckToken(ctx, req.GetToken(), req.GetAppId())
	if err != nil {
		//if errors.Is(err, auth.ErrInvalidToken) {
		//	return nil, status.Error(codes.InvalidArgument, err.Error())
		//}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.CheckTokenResponse{
		UserId:  claims.UserId,
		IsValid: true,
		Roles:   claims.Roles,
		Email:   claims.Email,
	}, nil

}
