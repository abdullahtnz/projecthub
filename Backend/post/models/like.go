package models

import (
	"time"
)

type Like struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}
