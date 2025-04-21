package main

import (
	"todolist/todo-backend/database"
	"todolist/todo-backend/handler"
	"todolist/todo-backend/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db := database.Connect()

	e := echo.New()
	e.Use(middleware.CORS())

	// Menyajikan semua file statis dari folder "frontend"
	e.Static("/", "frontend")
	e.GET("/", func(c echo.Context) error {
		return c.File("frontend/landing_page.html")
	})

	// Public endpoints
	e.POST("/login", handler.LoginHandler(db))
	e.POST("/register", handler.RegisterHandler(db))

	// Group endpoint yang butuh login
	api := e.Group("/api")
	api.Use(utils.JWTMiddleware)

	//update user
	api.PUT("/update-username", handler.UpdateUsernameHandler(db))
	api.PUT("/update-password", handler.UpdatePasswordHandler(db))
	api.POST("/delete-account", handler.DeleteAccountHandler(db))

	// Notes endpoint
	api.GET("/notes", handler.GetNotes(db))
	api.POST("/notes", handler.CreateNote(db))
	api.PUT("/notes/:id", handler.UpdateNote(db))
	api.DELETE("/notes/:id", handler.DeleteNote(db))

	// Tasks endpoint
	api.GET("/tasks", handler.GetTasks(db))
	api.GET("/tasks/:id", handler.GetTaskByID(db))
	api.GET("/tasks/date/:date", handler.GetTasksByDate(db))
	api.POST("/tasks", handler.CreateTask(db))
	api.PUT("/tasks/:id", handler.UpdateTask(db))
	api.DELETE("/tasks/:id", handler.DeleteTask(db))

	e.Logger.Fatal(e.Start(":8080"))
}
