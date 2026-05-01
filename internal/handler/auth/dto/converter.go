package dto

import (
	"github.com/platonso/hrmate-api/internal/domain"
	"github.com/platonso/hrmate-api/internal/service/auth/model"
)

func ToRegisterInput(req *RegisterRequest) *model.RegisterInput {
	return &model.RegisterInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Position:  req.Position,
		Email:     req.Email,
		Password:  req.Password,
		Role:      domain.Role(req.Role),
	}
}
