package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
)

func GetUserRole(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(userRoleKey).(domain.Role)
	return role, ok
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
