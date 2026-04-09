package form

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate/internal/domain"
	errs "github.com/platonso/hrmate/internal/errors"
	"github.com/platonso/hrmate/internal/handler/form/dto"
	"github.com/platonso/hrmate/internal/handler/middleware"
	"github.com/platonso/hrmate/internal/handler/request"
	"github.com/platonso/hrmate/internal/handler/response"
	formservice "github.com/platonso/hrmate/internal/service/form"
	"github.com/platonso/hrmate/internal/service/form/model"
)

type Service interface {
	Create(ctx context.Context, formDTO *model.FormCreateInput, userID uuid.UUID) (*domain.Form, error)
	GetForm(ctx context.Context, formID uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role) (*domain.Form, error)
	GetForms(ctx context.Context, filter *formservice.Filter, requesterID uuid.UUID, requesterRole domain.Role) ([]domain.Form, error)
	GetFormsWithUsers(ctx context.Context, filter *formservice.Filter, requesterID uuid.UUID, requesterRole domain.Role) ([]model.FormsWithUser, error)
	Approve(ctx context.Context, formID uuid.UUID, comment string) (*domain.Form, error)
	Reject(ctx context.Context, formID uuid.UUID, comment string) (*domain.Form, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) HandleCreateForm(w http.ResponseWriter, r *http.Request) {
	var req dto.FormCreateRequest

	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid request format")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	formCreateInput := dto.ToFormCreateInput(req)

	form, err := h.svc.Create(r.Context(), &formCreateInput, requesterID)
	if err != nil {
		response.WriteError(w, err, "failed to create form")
		return
	}

	response.WriteResponse(w, http.StatusCreated, dto.ToFormResponse(form))
}

func (h *Handler) HandleGetForm(w http.ResponseWriter, r *http.Request) {
	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid form id format")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	form, err := h.svc.GetForm(r.Context(), formID, requesterID, requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to get form")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.ToFormResponse(form))
}

func (h *Handler) HandleGetForms(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	filter, err := parseFilter(r)
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid filter parameters")
		return
	}

	forms, err := h.svc.GetForms(r.Context(), filter, requesterID, requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to get forms")
		return
	}

	response.WriteResponse(w, http.StatusOK, forms)
}

func (h *Handler) HandleGetFormsWithUsers(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	filter, err := parseFilter(r)
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid filter parameters")
		return
	}

	formsWithUsers, err := h.svc.GetFormsWithUsers(r.Context(), filter, requesterID, requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to get forms")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.ToFormsWithUserResponses(formsWithUsers))
}

func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	handleFormAction(w, r, h.svc.Approve)
}
func (h *Handler) HandleReject(w http.ResponseWriter, r *http.Request) {
	handleFormAction(w, r, h.svc.Reject)
}

func parseFilter(r *http.Request) (*formservice.Filter, error) {
	filter := &formservice.Filter{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid user id: %w", err)
		}
		filter.UserID = &userID
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.FormStatus(statusStr)
		filter.FormStatus = &status
		if err := filter.ValidateStatus(); err != nil {
			return nil, fmt.Errorf("invalid form status: %s", statusStr)
		}
	}

	return filter, nil
}

func handleFormAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(ctx context.Context, formID uuid.UUID, comment string) (*domain.Form, error),
) {
	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid form id format")
		return
	}
	var req dto.FormCommentRequest

	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid request format")
		return
	}

	form, err := action(r.Context(), formID, req.Comment)
	if err != nil {
		response.WriteError(w, err, "failed to process form action")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.ToFormResponse(form))
}
