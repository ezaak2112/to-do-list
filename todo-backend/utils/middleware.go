package utils

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// Secret key untuk JWT
var jwtSecretKey = []byte("your_secret_key")

// Middleware untuk memverifikasi JWT
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ambil token dari header Authorization
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(401, "Authorization token is missing")
		}

		// Pisahkan token dari 'Bearer'
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return echo.NewHTTPError(401, "Bearer token is missing")
		}

		// Parsing token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})
		if err != nil || !token.Valid {
			return echo.NewHTTPError(401, "Invalid or expired token")
		}

		// Ambil claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return echo.NewHTTPError(401, "Invalid token claims")
		}

		// Simpan user_id di context untuk digunakan oleh handler
		c.Set("user_id", claims.UserID)
		c.Logger().Info("User ID from token:", claims.UserID)

		return next(c)
	}
}
