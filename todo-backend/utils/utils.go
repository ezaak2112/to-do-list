package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your_secret_key") // Ganti sesuai kebutuhan

// JWTClaims untuk mengambil data dari token
type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

func GenerateJWT(userID uint) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// Fungsi untuk memverifikasi JWT token di header Authorization
func ValidateJWT(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token tidak ditemukan")
	}

	// Token dikirim dengan format: Bearer <token>
	tokenString := strings.Split(authHeader, " ")[1]

	// Parse token dan verifikasi
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pastikan token menggunakan algoritma yang benar
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("token signed with invalid method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid")
	}

	// Validasi token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Ambil user_id dari claims dan set ke context untuk digunakan di handler
		c.Set("user_id", claims["user_id"])
		return nil
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
