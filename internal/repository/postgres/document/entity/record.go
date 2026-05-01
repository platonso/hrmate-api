package entity

import (
	"time"

	"github.com/google/uuid"
)

type DocumentRecord struct {
	ID           uuid.UUID `db:"id"`
	FormID       uuid.UUID `db:"form_id"`
	ObjectKey    string    `db:"object_key"`
	OriginalName string    `db:"original_name"`
	UploadedAt   time.Time `db:"uploaded_at"`
	Type         string    `json:"type"`
}
