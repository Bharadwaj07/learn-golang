package middleware

import (
	"auth-api/internal/domain"
	"auth-api/internal/handler"
	"auth-api/pkg/jwt"
	"context"
	"net/http"
	"strings"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, domain.NewUnauthorized("missing authorization header"))
				return
			}

			// 2. header must be "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, domain.NewUnauthorized("invalid authorization format"))
				return
			}

			// 3. validate the token
			claims, err := jwt.Validate(parts[1], jwtSecret)
			if err != nil {
				writeError(w, domain.NewUnauthorized("invalid or expired token"))
				return
			}

			// 4. inject user_id into context for handlers to use
			ctx := context.WithValue(r.Context(), handler.UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeError is duplicated here to avoid circular imports
// in a larger project you'd move it to a shared pkg/httputil package
func writeError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if e, ok := err.(*domain.AppError); ok {
		appErr = e
	}
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Code)
		return
	}
	http.Error(w, "internal server error", 500)
}
