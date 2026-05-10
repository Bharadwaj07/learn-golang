package domain

import (
	"context"
	"time"
)

// User is the core model
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // "-" means never include in JSON
	CreatedAt time.Time `json:"created_at"`
}

// UserRepository — the interface your service depends on
// repository layer implements this, mock can implement it for tests
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
}

// UserService — the interface your handler depends on
type UserService interface {
	Register(ctx context.Context, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, id uint) (*User, error)
}
