package entity

import "github.com/platonso/hrmate-api/internal/domain"

func ToUserRecord(u domain.User) UserRecord {
	record := UserRecord{
		ID:             u.ID,
		Role:           string(u.Role),
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Position:       u.Position,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
		IsActive:       u.IsActive,
	}
	return record
}

func ToDomainUser(ur UserRecord) domain.User {
	user := domain.User{
		ID:             ur.ID,
		Role:           domain.Role(ur.Role),
		FirstName:      ur.FirstName,
		LastName:       ur.LastName,
		Position:       ur.Position,
		Email:          ur.Email,
		HashedPassword: ur.HashedPassword,
		IsActive:       ur.IsActive,
	}
	return user
}

func ToDomainUsers(records []UserRecord) []domain.User {
	users := make([]domain.User, len(records))
	for i := range records {
		users[i] = ToDomainUser(records[i])
	}
	return users
}
