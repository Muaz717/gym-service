package authMiddleware

import (
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/clients/sso/grpc"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/gin-gonic/gin"

	"log/slog"
	"net/http"
)

const userContextKey = "user"

func AuthMiddleware(log *slog.Logger, ssoClient *grpc.SSOClient, appId int32, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "middleware.AuthMiddleware"

		log = log.With(
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
			log.Error("failed to check token", slog.String("op", op), sl.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token validation failed"})
			return
		}

		log.Info("token check result",
			slog.Bool("is_valid", resp.IsValid),
			slog.Int64("user_id", resp.GetUserId()),
			slog.Any("roles", resp.Roles),
			slog.String("token", token),
			slog.Int("app_id", int(appId)),
		)

		if !resp.IsValid {
			log.Warn("invalid token", slog.String("op", op), slog.String("token", token))
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
			log.Warn(fmt.Sprintf("%s role required", requiredRole), slog.String("op", op), slog.Int64("user_id", resp.GetUserId()))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("%s role required", requiredRole)})
			return
		}

		c.Set(userContextKey, resp)
		c.Next()
	}
}

//func GetUserFromContext(c *gin.Context)
