package dto

import "github.com/platonso/hrmate-api/internal/domain"

func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Role:      string(user.Role),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Position:  user.Position,
		Email:     user.Email,
		IsActive:  user.IsActive,
	}
}

func ToUserResponses(users []domain.User) []UserResponse {
	if len(users) == 0 {
		return []UserResponse{}
	}

	responses := make([]UserResponse, len(users))
	for i := range users {
		responses[i] = ToUserResponse(&users[i])
	}
	return responses
}
