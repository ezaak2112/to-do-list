package handler

import (
	"database/sql"
	"log"
	"net/http"
	"todolist/todo-backend/model"
	"todolist/todo-backend/utils"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUsernameRequest struct {
	NewUsername string `json:"new_username"`
}

type UpdatePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func LoginHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Permintaan tidak valid"})
		}

		var user model.User
		err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", req.Username).
			Scan(&user.ID, &user.Password)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Username tidak ditemukan"})
		}

		if !utils.CheckPasswordHash(req.Password, user.Password) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Password salah"})
		}

		// Generate JWT token
		token, err := utils.GenerateJWT(uint(user.ID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal membuat token"})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Login berhasil",
			"token":   token,
		})
	}
}

func RegisterHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Permintaan tidak valid"})
		}

		// Validasi username dan password
		if req.Username == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Username dan Password harus diisi"})
		}

		// Cek apakah username sudah terdaftar
		var existingUser model.User
		err := db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingUser.ID)
		if err == nil {
			return c.JSON(http.StatusConflict, echo.Map{"message": "Username sudah terdaftar"})
		}

		// Hash password sebelum disimpan
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Terjadi kesalahan saat meng-hash password"})
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", req.Username, hashedPassword)
		if err != nil {
			log.Println("Error insert user:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal mendaftar pengguna"})
		}

		return c.JSON(http.StatusCreated, echo.Map{
			"message": "Pendaftaran berhasil",
		})
	}
}

// Fungsi untuk update username
func UpdateUsernameHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ambil user_id dari token JWT
		userIDFromToken, ok := c.Get("user_id").(uint)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
		}

		// Ambil username baru dari form data
		var req struct {
			NewUsername string `json:"new_username"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		// Cek jika username sudah ada di database
		existingUser, err := getUserByUsername(db, req.NewUsername)
		if err == nil && existingUser.Username == req.NewUsername {
			return echo.NewHTTPError(http.StatusConflict, "Username already taken")
		}

		// Update username di database
		err = updateUsernameInDB(db, userIDFromToken, req.NewUsername)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update username")
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Username updated successfully"})
	}
}

// Fungsi untuk update password
func UpdatePasswordHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDFromToken, ok := c.Get("user_id").(uint)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
		}

		// Ambil data password
		var req struct {
			OldPassword     string `json:"old_password"`
			NewPassword     string `json:"new_password"`
			ConfirmPassword string `json:"confirm_password"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		// Validasi password baru
		if req.NewPassword != req.ConfirmPassword {
			return echo.NewHTTPError(http.StatusBadRequest, "Password confirmation does not match")
		}

		// Ambil user dari database
		user, err := getUserByID(db, userIDFromToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}

		// Cek apakah password lama benar
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Incorrect old password")
		}

		// Hash password baru
		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
		}

		// Update password di database
		err = updatePasswordInDB(db, userIDFromToken, string(newPasswordHash))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update password")
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
	}
}

// Fungsi untuk menghapus akun
func DeleteAccountHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDFromToken, ok := c.Get("user_id").(uint)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
		}

		// Hapus user dari database
		err := deleteUserInDB(db, userIDFromToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete account")
		}

		// Logout / hapus session (opsional, tergantung implementasi logout)
		return c.JSON(http.StatusOK, map[string]string{"message": "Account deleted successfully"})
	}
}

// Helper function untuk mengambil user berdasarkan ID
func getUserByID(db *sql.DB, userID uint) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, password FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Helper function untuk mengambil user berdasarkan username
func getUserByUsername(db *sql.DB, username string) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Helper function untuk update username di database
func updateUsernameInDB(db *sql.DB, userID uint, newUsername string) error {
	_, err := db.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userID)
	return err
}

// Helper function untuk update password di database
func updatePasswordInDB(db *sql.DB, userID uint, newPassword string) error {
	_, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", newPassword, userID)
	return err
}

// Helper function untuk menghapus user dari database
func deleteUserInDB(db *sql.DB, userID uint) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	return err
}
