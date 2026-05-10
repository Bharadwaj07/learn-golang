package service

import (
	"auth-api/internal/domain"
	"auth-api/pkg/jwt"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      domain.UserRepository
	jwtSecret string
}

// NewUserService returns a UserService
func NewUserService(repo domain.UserRepository, jwtSecret string) domain.UserService {
	return &userService{repo: repo, jwtSecret: jwtSecret}
}

func (s *userService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	// 1. validate inputs
	if email == "" || password == "" {
		return nil, domain.NewBadRequest("email and password are required")
	}
	if len(password) < 8 {
		return nil, domain.NewBadRequest("password must be at least 8 characters")
	}

	// 2. hash the password — never store plain text
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternal(fmt.Errorf("hashing password: %w", err))
	}

	// 3. create user in db
	user := &domain.User{
		Email:    email,
		Password: string(hash),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, domain.NewConflict("email already registered")
		}
		slog.Error("register: failed to create user", "err", err)
		return nil, domain.NewInternal(err)
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// 1. find user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// don't reveal whether the email exists
			return "", domain.NewUnauthorized("invalid email or password")
		}
		slog.Error("login: failed to get user", "err", err)
		return "", domain.NewInternal(err)
	}

	// 2. compare password against stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", domain.NewUnauthorized("invalid email or password")
	}

	// 3. generate JWT
	token, err := jwt.Generate(user.ID, s.jwtSecret)
	if err != nil {
		slog.Error("login: failed to generate token", "err", err)
		return "", domain.NewInternal(err)
	}

	slog.Info("user logged in", "user_id", user.ID)
	return token, nil
}

func (s *userService) GetProfile(ctx context.Context, id uint) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewNotFound("user not found", err)
		}
		slog.Error("getProfile: failed to get user", "err", err)
		return nil, domain.NewInternal(err)
	}
	return user, nil
}
