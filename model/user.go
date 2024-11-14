package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID            uuid.UUID  `json:"user_id"`
	Username          string     `json:"username"`
	DisplayName       string     `json:"display_name"`
	Email             string     `json:"email"`
	Password          string     `json:"password"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	LastOnline        *time.Time `json:"last_online"`
	ProfilePictureURL string     `json:"profile_picture_url"`
}
