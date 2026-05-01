package user

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
)

type Repository interface {
	Update(ctx context.Context, user *domain.User) error
	FindByRole(ctx context.Context, roles ...domain.Role) ([]domain.User, error)
	FindByID(ctx context.Context, userId uuid.UUID) (*domain.User, error)
	IsActive(ctx context.Context, userID uuid.UUID) (bool, error)
}
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID, requesterRole domain.Role, requesterID uuid.UUID) (*domain.User, error) {

	var hasAccess bool
	switch requesterRole {
	case domain.RoleAdmin:
		hasAccess = true
	case domain.RoleHR:
		hasAccess = true
	case domain.RoleEmployee:
		hasAccess = userID == requesterID
	}

	if !hasAccess {
		return nil, errs.ErrUserNotFound
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUserNotFound
		}
		log.Printf("failed to get user %s: %v", userID, err)
		return nil, errs.ErrInternalServer
	}

	return user, nil
}

func (s *Service) ChangeActiveStatus(ctx context.Context, userID uuid.UUID, isActive bool) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUserNotFound
		}
		log.Printf("failed to find user %s: %v", userID, err)
		return nil, errs.ErrInternalServer
	}

	var changed bool
	if isActive {
		changed = user.Activate()
	} else {
		changed = user.Deactivate()
	}

	if changed {
		if err := s.repo.Update(ctx, user); err != nil {
			log.Printf("failed to update user %s: %v", userID, err)
			return nil, errs.ErrInternalServer
		}
	}

	return user, nil
}

func (s *Service) GetUsersByRole(ctx context.Context, requesterRole domain.Role) ([]domain.User, error) {
	var rolesToQuery []domain.Role

	switch requesterRole {
	case domain.RoleAdmin:
		rolesToQuery = []domain.Role{domain.RoleHR, domain.RoleEmployee}
	case domain.RoleHR:
		rolesToQuery = []domain.Role{domain.RoleEmployee}
	default:
		return nil, errs.ErrForbidden
	}

	users, err := s.repo.FindByRole(ctx, rolesToQuery...)
	if err != nil {
		log.Printf("failed to find users by roles %v: %v", rolesToQuery, err)
		return nil, errs.ErrInternalServer
	}

	if users == nil {
		users = []domain.User{}
	}

	return users, nil
}

func (s *Service) IsActive(ctx context.Context, userID uuid.UUID) (bool, error) {
	active, err := s.repo.IsActive(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return false, errs.ErrUserNotFound
		}
		log.Printf("failed to check active status for user %s: %v", userID, err)
		return false, errs.ErrInternalServer
	}
	return active, nil
}
