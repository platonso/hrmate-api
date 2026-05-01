package errors

import "errors"

var (
	// Auth errors
	ErrForbidden          = errors.New("FORBIDDEN")
	ErrUnauthorized       = errors.New("UNAUTHORIZED")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")

	// Request errors
	ErrInvalidRequest = errors.New("INVALID_REQUEST")
	ErrInternalServer = errors.New("INTERNAL_ERROR")

	// Form errors
	ErrFormNotFound        = errors.New("FORM_NOT_FOUND")
	ErrFormAlreadyRejected = errors.New("FORM_ALREADY_REJECTED")
	ErrFormAlreadyApproved = errors.New("FORM_ALREADY_APPROVED")

	// Assignment errors
	ErrNoAvailableExecutors = errors.New("NO_AVAILABLE_EXECUTORS")

	// User errors
	ErrUserNotFound      = errors.New("USER_NOT_FOUND")
	ErrUserNotActive     = errors.New("USER_NOT_ACTIVE")
	ErrUserAlreadyExists = errors.New("USER_ALREADY_EXISTS")

	// Document errors
	ErrDocumentNotFound = errors.New("DOCUMENT_NOT_FOUND")
)
