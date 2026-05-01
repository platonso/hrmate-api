package response

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	errs "github.com/platonso/hrmate-api/internal/errors"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code" validate:"required" enums:"FORBIDDEN,UNAUTHORIZED,INVALID_CREDENTIALS,INVALID_REQUEST,INTERNAL_ERROR,FORM_NOT_FOUND,FORM_ALREADY_REJECTED,FORM_ALREADY_APPROVED,NO_AVAILABLE_EXECUTORS,USER_NOT_FOUND,USER_NOT_ACTIVE,USER_ALREADY_EXISTS,DOCUMENT_NOT_FOUND" example:"ERROR_CODE"`
		Message string `json:"message" validate:"required"`
	} `json:"error" validate:"required"`
}

func WriteError(w http.ResponseWriter, err error, msg string) {
	if err == nil {
		err = errs.ErrInternalServer
		msg = "unknown error"
	}

	var statusCode int

	switch {
	case errors.Is(err, errs.ErrInvalidCredentials),
		errors.Is(err, errs.ErrUnauthorized):
		statusCode = http.StatusUnauthorized

	case errors.Is(err, errs.ErrUserNotActive),
		errors.Is(err, errs.ErrForbidden):
		statusCode = http.StatusForbidden

	case errors.Is(err, errs.ErrUserAlreadyExists),
		errors.Is(err, errs.ErrFormAlreadyApproved),
		errors.Is(err, errs.ErrFormAlreadyRejected),
		errors.Is(err, errs.ErrNoAvailableExecutors):
		statusCode = http.StatusConflict

	case errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrFormNotFound),
		errors.Is(err, errs.ErrDocumentNotFound):
		statusCode = http.StatusNotFound

	case errors.Is(err, errs.ErrInvalidRequest):
		statusCode = http.StatusBadRequest

	default:
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var errResponse ErrorResponse
	errResponse.Error.Code = err.Error()
	errResponse.Error.Message = msg

	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
