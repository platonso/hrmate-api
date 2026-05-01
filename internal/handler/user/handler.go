package user

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/handler/middleware"
	"github.com/platonso/hrmate-api/internal/handler/response"
	"github.com/platonso/hrmate-api/internal/handler/user/dto"
)

type Service interface {
	GetUserByID(ctx context.Context, userID uuid.UUID, requesterRole domain.Role, requesterID uuid.UUID) (*domain.User, error)
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

// @Summary Get current user profile
// @Description Returns the profile information of the authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me [get]
func (h *Handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), requesterID, requesterRole, requesterID)
	if err != nil {
		response.WriteError(w, err, "failed to get user")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToUserResponse(user))
}

// @Summary Get user
// @Description Get user based on the requester's role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/user [get]
// @Router /hr/user/{id} [get]
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid user id format")
		return
	}

	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), userID, requesterRole, requesterID)
	if err != nil {
		response.WriteError(w, err, "failed to get user")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToUserResponse(user))
}

// @Summary Get users
// @Description Get users based on the requester's role.
// @Description **Admin**: Returns all users with roles 'employee' and 'hr'
// @Description **HR**: Returns only users with role 'employee'
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users [get]
// @Router /hr/users [get]
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

	response.WriteJSON(w, http.StatusOK, dto.ToUserResponses(users))
}

// @Summary Activate user
// @Description Change a user's active status to true
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{id}/activate [patch]
func (h *Handler) HandleActivate(w http.ResponseWriter, r *http.Request) {
	h.handleChangeActiveStatus(w, r, true)
}

// @Summary Deactivate user
// @Description Change a user's active status to false
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{id}/deactivate [patch]
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

	response.WriteJSON(w, http.StatusOK, dto.ToUserResponse(user))
}
