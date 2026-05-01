package model

import "github.com/platonso/hrmate-api/internal/domain"

type RegisterInput struct {
	FirstName string
	LastName  string
	Position  string
	Email     string
	Password  string
	Role      domain.Role
}
