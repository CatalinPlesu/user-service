package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents the user model which maps to the "users" table in PostgreSQL
type User struct {
	bun.BaseModel `bun:"table:users"` // Tells Bun to use the "users" table

	UserID      uuid.UUID  `bun:"user_id,type:uuid,default:gen_random_uuid(),pk" json:"user_id"`
	Username    string     `bun:"username,unique,notnull" json:"username"`
	DisplayName string     `bun:"display_name,notnull" json:"display_name"`
	Email       string     `bun:"email,unique,notnull" json:"email"`
	Password    string     `bun:"password,notnull" json:"password"`
	CreatedAt   *time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
