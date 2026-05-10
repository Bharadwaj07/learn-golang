package repository

import (
	"auth-api/internal/domain"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository returns a UserRepository backed by postgres
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// check for unique constraint violation
		if isUniqueViolation(result.Error) {
			return fmt.Errorf("create user: %w", domain.ErrConflict)
		}
		return fmt.Errorf("create user: %w", result.Error)
	}
	return nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("GetByEmail: %w", domain.ErrNotFound)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("GetByEmail: %w", result.Error)
	}
	return &user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("GetByID: %w", domain.ErrNotFound)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("GetByID: %w", result.Error)
	}
	return &user, nil
}

// isUniqueViolation checks for postgres unique constraint error code 23505
func isUniqueViolation(err error) bool {
	return err != nil && contains(err.Error(), "23505")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
