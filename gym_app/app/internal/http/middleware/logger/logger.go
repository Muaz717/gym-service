package loggerMiddleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func New(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		log = log.With(
			slog.String("component", "middleware/logger"),
		)

		c.Next()

		duration := time.Since(start)

		log.Info("HTTP Request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.String("client_ip", c.ClientIP()),
			slog.Duration("duration", duration),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
