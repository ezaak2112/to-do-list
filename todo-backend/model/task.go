package model

import "time"

type Task struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
}
