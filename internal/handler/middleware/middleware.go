package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/handler/response"
)

type userIDKeyType struct{}
type userRoleKeyType struct{}

var userIDKey = userIDKeyType{}
var userRoleKey = userRoleKeyType{}

type AuthService interface {
	GetJWTSecret() string
}

type UserService interface {
	IsActive(ctx context.Context, userID uuid.UUID) (bool, error)
}

type Auth struct {
	AuthSvc AuthService
	UserSvc UserService
}

func (m *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.WriteError(w, errs.ErrUnauthorized, "authorization header is required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.WriteError(w, errs.ErrUnauthorized, "bearer token is required")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.AuthSvc.GetJWTSecret()), nil
		})

		if err != nil || !token.Valid {
			response.WriteError(w, errs.ErrUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.WriteError(w, errs.ErrUnauthorized, "invalid token claims")
			return
		}

		userIDStr, ok1 := claims["id"].(string)
		userRoleStr, ok2 := claims["role"].(string)
		if !ok1 || !ok2 {
			response.WriteError(w, errs.ErrUnauthorized, "invalid token payload: missing id or role")
			return
		}

		userRole := domain.Role(userRoleStr)

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.WriteError(w, errs.ErrUnauthorized, "invalid user id format in token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, userRole)

		if next == nil {
			response.WriteError(w, errs.ErrInternalServer, "endpoint not implemented")
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Auth) RequireRoles(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				response.WriteError(w, errs.ErrUnauthorized, "missing user role in context")
				return
			}

			for _, role := range allowedRoles {
				if userRole == role {
					if next == nil {
						response.WriteError(w, errs.ErrInternalServer, "endpoint not implemented")
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			response.WriteError(w, errs.ErrForbidden, fmt.Sprintf("role '%s' is not allowed to access this resource", userRole))
		})
	}
}

func (m *Auth) RequireActiveStatus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			response.WriteError(w, errs.ErrUnauthorized, "authentication required")
			return
		}

		isActive, err := m.UserSvc.IsActive(r.Context(), userID)
		if err != nil {
			if errors.Is(err, errs.ErrUserNotFound) {
				response.WriteError(w, errs.ErrUnauthorized, "authentication required")
				return
			}
			log.Printf("failed to check isActive status: %v", err)
			response.WriteError(w, errs.ErrInternalServer, "failed to verify account status")

			return
		}

		if !isActive {
			response.WriteError(w, errs.ErrForbidden, "account is not active")
			return
		}

		if next == nil {
			response.WriteError(w, errs.ErrInternalServer, "endpoint not implemented")
			return
		}

		next.ServeHTTP(w, r)
	})
}
