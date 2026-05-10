package router

import (
	"auth-api/internal/handler"
	"auth-api/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Setup(userHandler *handler.UserHandler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	// global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// public routes
	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Use(middleware.RateLimiter(10, 20)) // 10 rps, burst of 20

		r.Get("/profile", userHandler.GetProfile)
	})

	return r
}
