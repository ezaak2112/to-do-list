package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"todolist/todo-backend/model"

	"github.com/labstack/echo/v4"
)

// GetNotes mengambil semua catatan milik user yang login
func GetNotes(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)

		query := "SELECT id, title, content, date, created_at, updated_at FROM notes WHERE user_id = ?"
		rows, err := db.Query(query, userID)
		if err != nil {
			c.Logger().Error("Query error: ", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal mengambil data"})
		}
		defer rows.Close()

		var notes []model.Notes
		for rows.Next() {
			var note model.Notes
			err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Date, &note.CreatedAt, &note.UpdatedAt)
			if err != nil {
				c.Logger().Error("Scan error: ", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal membaca data"})
			}
			notes = append(notes, note)
		}

		return c.JSON(http.StatusOK, notes)
	}
}
func CreateNote(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)

		var note model.Notes
		if err := c.Bind(&note); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Input tidak valid"})
		}

		note.UserID = userID
		note.CreatedAt = time.Now()
		note.UpdatedAt = time.Now()

		query := "INSERT INTO notes (`title`, `content`, `date`, `user_id`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?)"
		_, err := db.Exec(query, note.Title, note.Content, note.Date, note.UserID, note.CreatedAt, note.UpdatedAt)
		if err != nil {
			// Log error agar bisa di-debug lebih lanjut
			c.Logger().Error("Error executing insert query: ", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal menyimpan catatan"})
		}

		return c.JSON(http.StatusCreated, echo.Map{"message": "Catatan berhasil dibuat"})
	}
}

// UpdateNote mengedit catatan user yang login
func UpdateNote(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)
		id := c.Param("id")

		var note model.Notes
		if err := c.Bind(&note); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Input tidak valid"})
		}

		note.UpdatedAt = time.Now()

		result, err := db.Exec(
			"UPDATE notes SET title = ?, content = ?, updated_at = ? WHERE id = ? AND user_id = ?",
			note.Title, note.Content, note.UpdatedAt, id, userID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal memperbarui catatan"})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Catatan tidak ditemukan atau bukan milikmu"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Catatan berhasil diperbarui"})
	}
}

// DeleteNote menghapus catatan user yang login
func DeleteNote(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)
		id := c.Param("id")

		noteID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "ID tidak valid"})
		}

		result, err := db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?", noteID, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal menghapus catatan"})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Catatan tidak ditemukan atau bukan milikmu"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Catatan berhasil dihapus"})
	}
}
