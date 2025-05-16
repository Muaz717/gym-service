package jwt

import (
	"errors"
	"github.com/Muaz717/sso/app/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"

	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type Claims struct {
	UserId int64    `json:"uid"`
	AppID  int      `json:"app_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	//Exp  int64  `json:"exp"`
	jwt.RegisteredClaims
}

func ParseToken(tokenStr string, app models.App) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(app.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
