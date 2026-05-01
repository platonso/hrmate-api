package form

import (
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
)

type Filter struct {
	UserID     *uuid.UUID
	ExecutorID *uuid.UUID
	FormStatus *domain.FormStatus
	SortOrder  string
}

func (f *Filter) ValidateStatus() error {
	if f.FormStatus == nil {
		return nil
	}
	switch *f.FormStatus {
	case domain.StatusPending, domain.StatusApproved, domain.StatusRejected:
		return nil
	default:
		return errs.ErrInvalidRequest
	}
}
