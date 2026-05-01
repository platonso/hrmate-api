package auth

import (
	"context"
	"net/http"

	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/handler/auth/dto"
	"github.com/platonso/hrmate-api/internal/handler/request"
	"github.com/platonso/hrmate-api/internal/handler/response"
	"github.com/platonso/hrmate-api/internal/service/auth/model"
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

// @Summary Register a new user
// @Description Register a new user and return an auth token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration details"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /register [post]
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

	response.WriteJSON(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

// @Summary User login
// @Description Authenticate a user and return an auth token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /login [post]
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

	response.WriteJSON(w, http.StatusOK, dto.AuthResponse{Token: token})
}
