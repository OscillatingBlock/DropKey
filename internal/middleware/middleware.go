package custom_middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)

		req := c.Request()
		res := c.Response()

		slog.Info(
			"http_request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", res.Status,
			"latency", time.Since(start).String(),
			"ip", c.RealIP(),
			"user_agent", req.UserAgent(),
			"request_id", req.Header.Get("X-Request-ID"),
			"error", err,
		)
		return err
	}
}

type UserInfo struct {
	UserID    string `json:"userID"`
	Publickey string `json:"public_key"`
}

type JwtCustomClaims struct {
	UserInfo
	jwt.RegisteredClaims
}

func JwtAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtSecret := os.Getenv("JWTSECRET")
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
		}
		tokenString := parts[1]

		claims := new(JwtCustomClaims)

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})
		if err != nil {
			slog.Error("JWT parsing error (Echo): %v", "error", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		if !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token expired")
		}

		c.Set("userInfo", claims.UserInfo)

		return next(c)
	}
}
