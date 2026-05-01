package auth

import (
	"context"
	"errors"
	"log"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/service/auth/model"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByRole(ctx context.Context, roles ...domain.Role) ([]domain.User, error)
}
type Service struct {
	txMgr     *manager.Manager
	repo      Repository
	jwtSecret string
}

func NewService(txMgr *manager.Manager, repo Repository, jwtSecret string) *Service {
	return &Service{
		txMgr:     txMgr,
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) ImplementAdmin(ctx context.Context, email, password string) error {

	admin, err := s.repo.FindByRole(ctx, domain.RoleAdmin)
	if err != nil {
		log.Printf("failed to find admin: %v", err)
		return errs.ErrInternalServer
	}

	if len(admin) > 0 {
		log.Println("the existing admin is used")
		return nil
	}

	if email == "" || password == "" {
		return errs.ErrInvalidRequest
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash admin password: %v", err)
		return errs.ErrInternalServer
	}

	adminUser := domain.NewUser(
		domain.RoleAdmin,
		"Super",
		"user",
		"Administrator",
		email,
		string(hashedPassword),
	)

	adminUser.Activate()

	if err := s.repo.Create(ctx, &adminUser); err != nil {
		log.Printf("failed to create admin: %v", err)
		return errs.ErrInternalServer
	}

	log.Println("admin has been created successfully")
	return nil
}

func (s *Service) Register(ctx context.Context, registerInput *model.RegisterInput) (string, error) {
	var user domain.User

	if err := s.txMgr.Do(ctx, func(txCtx context.Context) error {
		existingUser, err := s.repo.FindByEmail(txCtx, registerInput.Email)
		if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
			log.Printf("failed to check user existence: %v", err)
			return errs.ErrInternalServer
		}

		if existingUser != nil {
			return errs.ErrUserAlreadyExists
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerInput.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("failed to hash password: %v", err)
			return errs.ErrInternalServer
		}

		user = domain.NewUser(
			registerInput.Role,
			registerInput.FirstName,
			registerInput.LastName,
			registerInput.Position,
			registerInput.Email,
			string(hashedPassword),
		)

		if err := s.repo.Create(txCtx, &user); err != nil {
			log.Printf("failed to create user: %v", err)
			return errs.ErrInternalServer
		}
		return nil
	}); err != nil {
		return "", err
	}

	token, err := generateJWT(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		log.Printf("failed to generate JWT: %v", err)
		return "", errs.ErrInternalServer
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return "", errs.ErrInvalidCredentials
		}
		log.Printf("failed to find user by email: %v", err)
		return "", errs.ErrInternalServer
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", errs.ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", errs.ErrUserNotActive
	}

	token, err := generateJWT(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		log.Printf("failed to generate JWT: %v", err)
		return "", errs.ErrInternalServer
	}

	return token, nil
}

func (s *Service) GetJWTSecret() string {
	return s.jwtSecret
}

func generateJWT(userID uuid.UUID, role domain.Role, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   userID,
		"role": role,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
