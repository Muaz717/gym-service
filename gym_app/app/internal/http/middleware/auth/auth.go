package authMiddleware

import (
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	ssov1 "github.com/Muaz717/gym_app/app/pkg/sso"
	"github.com/gin-gonic/gin"

	"log/slog"
	"net/http"
)

const userContextKey = "user"

func AuthMiddleware(log *slog.Logger, ssoClient *grpc.SSOClient, appId int32, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "middleware.AuthMiddleware"

		reqLog := log.With(
			slog.String("op", op),
		)

		//authHeader := c.GetHeader("Authorization")
		//if authHeader == "" {
		//	log.Error("authorization header missing", slog.String("op", op))
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
		//	return
		//}
		//
		//token := strings.TrimPrefix(authHeader, "Bearer ")
		//if token == authHeader {
		//	log.Error("invalid authorization format", slog.String("op", op))
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		//	return
		//}

		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := ssoClient.CheckToken(c.Request.Context(), appId, token)
		if err != nil {
			reqLog.Error("failed to check token", slog.String("op", op), sl.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token validation failed"})
			return
		}

		reqLog.Info("token check result",

			slog.Any("roles", resp.Roles),
			slog.String("token", token),
			slog.Int("app_id", int(appId)),
		)

		if !resp.IsValid {
			reqLog.Warn("invalid token", slog.String("op", op), slog.String("token", token))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		hasRequiredRole := false
		for _, role := range resp.Roles {
			if role == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			reqLog.Warn(fmt.Sprintf("%s role required", requiredRole), slog.String("op", op), slog.Int64("user_id", resp.GetUserId()))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("%s role required", requiredRole)})
			return
		}

		c.Set(userContextKey, resp)
		c.Next()
	}
}

// GetUserFromContext достаёт пользователя (SSO CheckTokenResponse) из контекста gin
func GetUserFromContext(c *gin.Context) (*ssov1.CheckTokenResponse, bool) {
	val, exists := c.Get(userContextKey)
	if !exists {
		return nil, false
	}
	user, ok := val.(*ssov1.CheckTokenResponse)
	return user, ok
}
