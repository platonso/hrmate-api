package model

import (
	"time"

	"github.com/platonso/hrmate-api/internal/domain"
)

type DocumentInput struct {
	Name        string
	Size        int64
	ContentType string
	Content     []byte
}

type FormCreateInput struct {
	Title       string
	Description string
	StartDate   *time.Time
	EndDate     *time.Time
	Documents   []DocumentInput
}

type FormWithDocs struct {
	Form      domain.Form
	Documents []domain.Document
}
