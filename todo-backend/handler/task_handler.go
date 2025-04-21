package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todolist/todo-backend/model"

	"github.com/labstack/echo/v4"
)

// GetTasks mengambil semua tugas milik user yang login
func GetTasks(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		val := c.Get("user_id")
		userID, ok := val.(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Unauthorized"})
		}

		// Ambil filter dari query string
		title := c.QueryParam("title")
		deadline := c.QueryParam("deadline")
		status := c.QueryParam("status")

		query := "SELECT id, title, description, deadline, status FROM tasks WHERE user_id = ?"
		args := []interface{}{userID}

		if title != "" {
			query += " AND title LIKE ?"
			args = append(args, "%"+title+"%")
		}
		if deadline != "" {
			query += " AND DATE(deadline) = ?"
			args = append(args, deadline)
		}
		if status != "" {
			query += " AND status = ?"
			args = append(args, status)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal mengambil data tugas"})
		}
		defer rows.Close()

		var tasks []model.Task
		for rows.Next() {
			var t model.Task
			err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Deadline, &t.Status)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Error membaca data"})
			}
			tasks = append(tasks, t)
		}

		return c.JSON(http.StatusOK, tasks)
	}
}

func GetTaskByID(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		val := c.Get("user_id")
		userID, ok := val.(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Unauthorized"})
		}

		taskID := c.Param("id")

		var t model.Task
		err := db.QueryRow("SELECT id, title, description, deadline, status FROM tasks WHERE id = ? AND user_id = ?", taskID, userID).
			Scan(&t.ID, &t.Title, &t.Description, &t.Deadline, &t.Status)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Tugas tidak ditemukan"})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal mengambil data tugas"})
		}

		return c.JSON(http.StatusOK, t)
	}
}

func GetTasksByDate(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		val := c.Get("user_id")
		fmt.Println("DEBUG raw user_id dari context:", val)

		userID, ok := val.(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Unauthorized"})
		}

		date := c.Param("date") // format: YYYY-MM-DD
		fmt.Println("DEBUG final user_id:", userID)
		fmt.Println("DEBUG param date:", date)

		rows, err := db.Query(`
			SELECT id, title, description, deadline, status 
			FROM tasks 
			WHERE user_id = ? AND DATE(deadline) = ?`, userID, date)
		if err != nil {
			fmt.Println("DEBUG query error:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal mengambil data tugas"})
		}
		defer rows.Close()

		var tasks []model.Task
		for rows.Next() {
			var t model.Task
			err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Deadline, &t.Status)
			if err != nil {
				fmt.Println("DEBUG row scan error:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal membaca data tugas"})
			}
			tasks = append(tasks, t)
		}

		fmt.Printf("DEBUG jumlah tugas ditemukan: %d\n", len(tasks))

		if len(tasks) == 0 {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Tugas tidak ditemukan"})
		}

		return c.JSON(http.StatusOK, tasks)
	}
}

// CreateTask membuat tugas baru milik user yang login
func CreateTask(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		val := c.Get("user_id")
		userID, ok := val.(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Unauthorized"})
		}

		// Struct khusus untuk input, agar deadline bisa diterima sebagai string
		type TaskInput struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Deadline    string `json:"deadline"` // format "yyyy-mm-dd"
			Status      string `json:"status"`
		}

		var input TaskInput
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Input tidak valid"})
		}

		// Parse string deadline jadi time.Time
		parsedDeadline, err := time.Parse("2006-01-02", input.Deadline)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Format deadline harus YYYY-MM-DD"})
		}

		// Masukkan ke struct asli yang pakai time.Time
		task := model.Task{
			Title:       input.Title,
			Description: input.Description,
			Deadline:    parsedDeadline,
			Status:      input.Status,
			UserID:      userID,
		}

		query := "INSERT INTO tasks (title, description, deadline, status, user_id) VALUES (?, ?, ?, ?, ?)"
		_, err = db.Exec(query, task.Title, task.Description, task.Deadline, task.Status, task.UserID)
		if err != nil {
			c.Logger().Error("Insert error: ", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal menyimpan tugas"})
		}

		return c.JSON(http.StatusCreated, echo.Map{"message": "Tugas berhasil dibuat"})
	}
}

// UpdateTask mengedit tugas milik user yang login
func UpdateTask(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)
		id := c.Param("id")

		// Struct khusus input agar deadline bisa string
		type TaskInput struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Deadline    string `json:"deadline"` // yyyy-mm-dd
			Status      string `json:"status"`
		}

		var input TaskInput
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Input tidak valid"})
		}

		if input.Status != "Belum Selesai" && input.Status != "Selesai" {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Status harus 'Belum Selesai' atau 'Selesai'"})
		}

		// Parse deadline
		parsedDeadline, err := time.Parse("2006-01-02", input.Deadline)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Format deadline harus YYYY-MM-DD"})
		}

		// Eksekusi query update
		result, err := db.Exec(
			"UPDATE tasks SET title = ?, description = ?, deadline = ?, status = ? WHERE id = ? AND user_id = ?",
			input.Title, input.Description, parsedDeadline, input.Status, id, userID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal memperbarui tugas"})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Tugas tidak ditemukan atau bukan milikmu"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Tugas berhasil diperbarui"})
	}
}

// DeleteTask menghapus tugas milik user yang login
func DeleteTask(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uint)
		id := c.Param("id")

		taskID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "ID tidak valid"})
		}

		// Query untuk menghapus tugas berdasarkan ID
		result, err := db.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Gagal menghapus tugas"})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Tugas tidak ditemukan atau bukan milikmu"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Tugas berhasil dihapus"})
	}
}
