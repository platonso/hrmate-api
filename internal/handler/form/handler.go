package form

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/handler/form/dto"
	"github.com/platonso/hrmate-api/internal/handler/middleware"
	"github.com/platonso/hrmate-api/internal/handler/response"
	formservice "github.com/platonso/hrmate-api/internal/service/form"
	"github.com/platonso/hrmate-api/internal/service/form/model"
)

type Service interface {
	Create(ctx context.Context, formDTO *model.FormCreateInput, userID uuid.UUID) (*model.FormWithDocs, error)
	GetForm(ctx context.Context, formID, requesterID uuid.UUID, requesterRole domain.Role) (*model.FormWithDocs, error)
	DownloadDocument(ctx context.Context, docID, requesterID uuid.UUID, requesterRole domain.Role) (*domain.Document, io.ReadCloser, error)
	GetForms(ctx context.Context, filter *formservice.Filter, requesterID uuid.UUID, requesterRole domain.Role) ([]domain.Form, error)
	Approve(ctx context.Context, formID uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role, comment string, docsInput []model.DocumentInput) (*domain.Form, error)
	Reject(ctx context.Context, formID uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role, comment string, docsInput []model.DocumentInput) (*domain.Form, error)
	DeleteWithDocs(ctx context.Context, formID uuid.UUID, requesterRole domain.Role) error
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// @Summary Create form
// @Description Create a new form with optional documents
// @Tags Forms
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Form Title"
// @Param description formData string false "Form Description"
// @Param startDate formData string false "Start Date in RFC3339 format"
// @Param endDate formData string false "End Date in RFC3339 format"
// @Param documents formData file false "Documents to attach"
// @Success 201 {object} dto.FormWithDocsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms [post]
func (h *Handler) HandleCreateForm(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		response.WriteError(w, errs.ErrInvalidRequest, "title is required")
		return
	}

	description := r.FormValue("description")

	var startDate, endDate *time.Time
	if startStr := r.FormValue("startDate"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startDate = &t
		}
	}
	if endStr := r.FormValue("endDate"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endDate = &t
		}
	}

	formCreateInput := model.FormCreateInput{
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	docs, err := extractDocuments(r)
	if err != nil {
		log.Printf("failed to process file: %v", err)
		response.WriteError(w, errs.ErrInternalServer, "failed to process file")
		return
	}

	formCreateInput.Documents = docs

	formWithDocs, err := h.svc.Create(r.Context(), &formCreateInput, requesterID)
	if err != nil {
		response.WriteError(w, err, "failed to create form")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.ToFormWithDocsResponse(formWithDocs))
}

// @Summary Get a full form
// @Description Get a specific form and its documents
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID"
// @Success 200 {object} dto.FormWithDocsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/forms/{id} [get]
// @Router /hr/forms/{id} [get]
// @Router /forms/{id} [get]
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

	formWithDocs, err := h.svc.GetForm(r.Context(), formID, requesterID, requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to get form")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToFormWithDocsResponse(formWithDocs))
}

// @Summary Get forms
// @Description Get forms with optional filters
// @Tags Forms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by user ID"
// @Param status query string false "Filter by status" Enums(pending, approved, rejected)
// @Success 200 {array} dto.FormResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/forms [get]
// @Router /hr/forms [get]
// @Router /forms [get]
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

	response.WriteJSON(w, http.StatusOK, dto.ToFormsResponse(forms))
}

// @Summary Download document
// @Description Download a document attached to a form
// @Tags Documents
// @Accept json
// @Produce application/pdf
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Success 200 {file} file
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /documents/{id}/download [get]
func (h *Handler) HandleDownloadDocument(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid document id format")
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

	doc, stream, err := h.svc.DownloadDocument(r.Context(), docID, requesterID, requesterRole)
	if err != nil {
		response.WriteError(w, err, "failed to download document")
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("Error closing stream for document %s: %v", docID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.OriginalName))
	if _, err := io.Copy(w, stream); err != nil {
		log.Printf("Error streaming file to client: %v\n", err)
	}
}

// @Summary Approve form
// @Description Approve a form with an optional comment and documents
// @Tags Forms
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID"
// @Param comment formData string false "Approval comment"
// @Param documents formData file false "Documents to attach"
// @Success 200 {object} dto.FormResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /hr/forms/{id}/approve [patch]
func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
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

	docs, err := extractDocuments(r)
	if err != nil {
		log.Printf("failed to process file: %v", err)
		response.WriteError(w, errs.ErrInternalServer, "failed to process file")
		return
	}

	form, err := executeFormAction(r, docs, requesterID, requesterRole, h.svc.Approve)
	if err != nil {
		response.WriteError(w, err, "failed to approve form")
		return
	}

	formWithDocs, err := h.svc.GetForm(r.Context(), form.ID, form.ExecutorID, domain.RoleHR)
	if err != nil {
		response.WriteError(w, err, "failed to get form")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToFormWithDocsResponse(formWithDocs))

}

// @Summary Reject form
// @Description Reject a form with an optional comment and documents
// @Tags Forms
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Form ID"
// @Param comment formData string false "Rejection comment"
// @Param documents formData file false "Documents to attach"
// @Success 200 {object} dto.FormResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /hr/forms/{id}/reject [patch]
func (h *Handler) HandleReject(w http.ResponseWriter, r *http.Request) {
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

	docs, err := extractDocuments(r)
	if err != nil {
		log.Printf("failed to process file: %v", err)
		response.WriteError(w, errs.ErrInternalServer, "failed to process file")
		return
	}

	form, err := executeFormAction(r, docs, requesterID, requesterRole, h.svc.Reject)

	if err != nil {
		response.WriteError(w, err, "failed to approve form")
		return
	}

	formWithDocs, err := h.svc.GetForm(r.Context(), form.ID, form.ExecutorID, domain.RoleHR)
	if err != nil {
		response.WriteError(w, err, "failed to get form")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToFormWithDocsResponse(formWithDocs))
}

// @Summary DeleteWithDocs form
// @Description Permanently deletes a form and all its associated documents
// @Tags Forms
// @Security BearerAuth
// @Param id path string true "Form ID"
// @Success 204 "Form and all associated documents successfully deleted"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/forms/{id} [delete]
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, errs.ErrInvalidRequest, "invalid document id format")
		return
	}

	requesterRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		response.WriteError(w, errs.ErrUnauthorized, "authentication required")
		return
	}

	if err := h.svc.DeleteWithDocs(r.Context(), formID, requesterRole); err != nil {
		response.WriteError(w, err, "failed to delete form")
		return
	}

	response.WriteNoContent(w)
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

func executeFormAction(r *http.Request, docsInput []model.DocumentInput, requesterID uuid.UUID, requesterRole domain.Role,
	action func(ctx context.Context, formID uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role, comment string, docsInput []model.DocumentInput) (*domain.Form, error),
) (*domain.Form, error) {
	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return nil, errs.ErrInvalidRequest
	}

	comment := r.FormValue("comment")

	form, err := action(r.Context(), formID, requesterID, requesterRole, comment, docsInput)
	if err != nil {
		return nil, err
	}

	return form, nil
}

func extractDocuments(r *http.Request) ([]model.DocumentInput, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	var documents []model.DocumentInput

	files := r.MultipartForm.File["documents"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s: %w", fileHeader.Filename, err)
		}

		content, err := io.ReadAll(file)
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("failed to read file: %s: %w", fileHeader.Filename, err)
		}

		if err := file.Close(); err != nil {
			log.Printf("Failed to close file %s: %v", fileHeader.Filename, err)
		}

		documents = append(documents, model.DocumentInput{
			Name:        fileHeader.Filename,
			Size:        int64(len(content)),
			ContentType: fileHeader.Header.Get("Content-Type"),
			Content:     content,
		})
	}

	return documents, nil
}
