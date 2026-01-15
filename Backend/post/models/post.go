package models

import "time"

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Images    []string  `json:"images,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
