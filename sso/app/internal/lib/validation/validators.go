package validation

import (
	"strings"

	ssov1 "github.com/Muaz717/sso/app/pkg/sso"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewValidationError(fields map[string]string) error {
	var violations []*errdetails.BadRequest_FieldViolation

	for field, desc := range fields {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: desc,
		})
	}

	st := status.New(codes.InvalidArgument, "Ошибки валидации")
	stWithDetails, err := st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
	if err != nil {
		// fallback, если не удалось добавить детали
		return status.Error(codes.InvalidArgument, "Ошибка валидации")
	}

	return stWithDetails.Err()
}

func ValidateRegisterInput(req *ssov1.RegisterRequest) error {
	errors := make(map[string]string)

	if !strings.Contains(req.GetEmail(), "@") {
		errors["email"] = "Неверный формат email"
	}

	if len([]rune(req.GetPassword())) < 6 {
		errors["password"] = "Пароль должен содержать минимум 6 символов"
	}

	if len(errors) > 0 {
		return NewValidationError(errors)
	}
	return nil
}

func ValidateLoginInput(req *ssov1.LoginRequest) error {
	errors := make(map[string]string)

	if !strings.Contains(req.GetEmail(), "@") {
		errors["email"] = "Неверный формат email"
	}

	if len([]rune(req.GetPassword())) == 0 {
		errors["password"] = "Пароль не должен быть пустым"
	}

	if req.GetAppId() == 0 {
		errors["app_id"] = "App ID обязателен"
	}

	if len(errors) > 0 {
		return NewValidationError(errors)
	}
	return nil
}

func ValidateIsAdminInput(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == 0 {
		return NewValidationError(map[string]string{
			"user_id": "User ID обязателен",
		})
	}
	return nil
}

func ValidateLogoutInput(req *ssov1.LogoutRequest) error {
	errors := make(map[string]string)

	if strings.TrimSpace(req.GetToken()) == "" {
		errors["token"] = "Token обязателен"
	}

	if req.GetAppId() == 0 {
		errors["app_id"] = "App ID обязателен"
	}

	if len(errors) > 0 {
		return NewValidationError(errors)
	}
	return nil
}

func ValidateCheckTokenInput(req *ssov1.CheckTokenRequest) error {
	errors := make(map[string]string)

	if strings.TrimSpace(req.GetToken()) == "" {
		errors["token"] = "Token обязателен"
	}

	if req.GetAppId() == 0 {
		errors["app_id"] = "App ID обязателен"
	}

	if len(errors) > 0 {
		return NewValidationError(errors)
	}
	return nil
}
