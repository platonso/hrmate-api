package domain

import (
	"time"

	"github.com/google/uuid"
	errs "github.com/platonso/hrmate-api/internal/errors"
)

type FormStatus string

const (
	StatusPending  FormStatus = "pending"
	StatusApproved FormStatus = "approved"
	StatusRejected FormStatus = "rejected"
)

type Form struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ExecutorID  uuid.UUID
	Title       string
	Description string

	StartDate *time.Time
	EndDate   *time.Time

	CreatedAt  time.Time
	ReviewedAt *time.Time
	Status     FormStatus
	Comment    *string
}

func NewForm(userID, executorID uuid.UUID, title, description string, startDate, endDate *time.Time) Form {
	return Form{
		ID:          uuid.New(),
		UserID:      userID,
		ExecutorID:  executorID,
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now(),
		ReviewedAt:  nil,
		Status:      StatusPending,
	}
}

func (f *Form) ApproveForm(comment string) (bool, error) {
	if f.Status == StatusApproved {
		return false, errs.ErrFormAlreadyApproved
	}

	if f.Status == StatusRejected {
		return false, errs.ErrFormAlreadyRejected
	}

	if f.Status != StatusPending {
		return false, errs.ErrInvalidRequest
	}

	approveTime := time.Now()
	f.ReviewedAt = &approveTime
	f.Status = StatusApproved
	if comment != "" {
		f.Comment = &comment
	}

	return true, nil
}

func (f *Form) RejectForm(comment string) (bool, error) {
	if f.Status == StatusRejected {
		return false, errs.ErrFormAlreadyRejected
	}

	if f.Status == StatusApproved {
		return false, errs.ErrFormAlreadyApproved
	}

	if f.Status != StatusPending {
		return false, errs.ErrInvalidRequest
	}

	rejectTime := time.Now()
	f.ReviewedAt = &rejectTime
	f.Status = StatusRejected
	if comment != "" {
		f.Comment = &comment
	}

	return true, nil
}
