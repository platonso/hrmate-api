package auth

import (
	"context"
	"net/http"

	errs "github.com/platonso/hrmate/internal/errors"
	"github.com/platonso/hrmate/internal/handler/auth/dto"
	"github.com/platonso/hrmate/internal/handler/request"
	"github.com/platonso/hrmate/internal/handler/response"
	"github.com/platonso/hrmate/internal/service/auth/model"
)

type Service interface {
	Register(ctx context.Context, registerInput *model.RegisterInput) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid request format")
		return
	}

	token, err := h.svc.Register(r.Context(), dto.ToRegisterInput(&req))
	if err != nil {
		response.WriteError(w, err, "failed to register")
		return
	}

	response.WriteResponse(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid request format")
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.WriteError(w, err, "failed to login")
		return
	}

	response.WriteResponse(w, http.StatusOK, dto.AuthResponse{Token: token})
}
