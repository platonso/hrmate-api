package user

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate/internal/domain"
	errs "github.com/platonso/hrmate/internal/errors"
	"github.com/platonso/hrmate/internal/handler/middleware"
	"github.com/platonso/hrmate/internal/handler/response"
	"github.com/platonso/hrmate/internal/handler/user/dto"
)

type Service interface {
	GetUsersByRole(ctx context.Context, requesterRole domain.Role) ([]domain.User, error)
	ChangeActiveStatus(ctx context.Context, userID uuid.UUID, newStatus bool) (*domain.User, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	users, err := h.svc.GetUsersByRole(r.Context(), requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to get users by role")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.ToUserResponses(users))
}

func (h *Handler) HandleActivate(w http.ResponseWriter, r *http.Request) {
	h.handleChangeActiveStatus(w, r, true)
}
func (h *Handler) HandleDeactivate(w http.ResponseWriter, r *http.Request) {
	h.handleChangeActiveStatus(w, r, false)
}

func (h *Handler) handleChangeActiveStatus(
	w http.ResponseWriter,
	r *http.Request,
	newStatus bool,
) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid user id format")
		return
	}
	user, err := h.svc.ChangeActiveStatus(r.Context(), userID, newStatus)
	if err != nil {
		response.WriteError(w, err, "failed to change user's active status")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.ToUserResponse(user))
}
