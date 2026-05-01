package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DocumentType string

const (
	DocumentTypeAttachment DocumentType = "attachment"
	DocumentTypeResult     DocumentType = "result"
)

type Document struct {
	ID           uuid.UUID
	FormID       uuid.UUID
	ObjectKey    string
	OriginalName string
	UploadedAt   time.Time
	Type         DocumentType
}

func NewAttachDocument(formID uuid.UUID, originalName string) Document {
	return Document{
		ID:           uuid.New(),
		FormID:       formID,
		ObjectKey:    fmt.Sprintf("%s/%s", formID, uuid.New()),
		OriginalName: originalName,
		UploadedAt:   time.Now(),
		Type:         DocumentTypeAttachment,
	}
}

func NewResultDocument(formID uuid.UUID, originalName string) Document {
	return Document{
		ID:           uuid.New(),
		FormID:       formID,
		ObjectKey:    fmt.Sprintf("%s/%s", formID, uuid.New()),
		OriginalName: originalName,
		UploadedAt:   time.Now(),
		Type:         DocumentTypeResult,
	}
}
