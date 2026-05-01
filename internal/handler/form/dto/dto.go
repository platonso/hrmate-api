package dto

import (
	"time"

	"github.com/google/uuid"
)

type FormResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	Status    string    `json:"status"`
}

type FormWithDocsResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`

	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`

	CreatedAt  time.Time          `json:"createdAt"`
	Status     string             `json:"status"`
	Documents  []DocumentResponse `json:"attachDocs"`
	Resolution *Resolution        `json:"resolution,omitempty"`
}

type DocumentResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Resolution struct {
	Comment      *string            `json:"comment"`
	ResolvedAt   time.Time          `json:"resolvedAt"`
	ResponseDocs []DocumentResponse `json:"responseDocs"`
}
