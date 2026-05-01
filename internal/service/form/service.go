package form

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/service/assignment"
	"github.com/platonso/hrmate-api/internal/service/form/model"
)

type Repository interface {
	Create(ctx context.Context, form *domain.Form) error
	FindByID(ctx context.Context, formId uuid.UUID) (*domain.Form, error)
	FindByFilter(ctx context.Context, filter *Filter) ([]domain.Form, error)
	Update(ctx context.Context, form *domain.Form) error
	Delete(ctx context.Context, formID uuid.UUID) error
}

type UserRepository interface {
	FindByID(ctx context.Context, userId uuid.UUID) (*domain.User, error)
	FindActiveHRsWithWorkload(ctx context.Context) ([]assignment.HRWorkload, error)
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document) error
	FindByFormID(ctx context.Context, formID uuid.UUID) ([]domain.Document, error)
	FindByID(ctx context.Context, docID uuid.UUID) (*domain.Document, error)
}

type Storage interface {
	UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectKey string) error
}

type Service struct {
	txMgr    *manager.Manager
	formRepo Repository
	userRepo UserRepository
	docRepo  DocumentRepository
	storage  Storage
}

func NewService(txMgr *manager.Manager, formRepo Repository, userRepo UserRepository, docRepo DocumentRepository, storage Storage) *Service {
	return &Service{
		txMgr:    txMgr,
		formRepo: formRepo,
		userRepo: userRepo,
		docRepo:  docRepo,
		storage:  storage,
	}
}

func (s *Service) Create(ctx context.Context, formInput *model.FormCreateInput, userID uuid.UUID) (*model.FormWithDocs, error) {
	var resultForm model.FormWithDocs

	if err := s.txMgr.Do(ctx, func(txCtx context.Context) error {
		hrs, err := s.userRepo.FindActiveHRsWithWorkload(txCtx)
		if err != nil {
			log.Printf("failed to find active HRs: %v", err)
			return errs.ErrInternalServer
		}

		executorID, err := assignment.SelectOptimalHR(hrs)
		if err != nil {
			if errors.Is(err, errs.ErrNoAvailableExecutors) {
				return errs.ErrNoAvailableExecutors
			}
			log.Printf("failed to select optimal HR: %v", err)
			return errs.ErrInternalServer
		}

		form := domain.NewForm(
			userID,
			executorID,
			formInput.Title,
			formInput.Description,
			formInput.StartDate,
			formInput.EndDate,
		)

		if err := s.formRepo.Create(txCtx, &form); err != nil {
			log.Printf("failed to create form for user %s: %v", userID, err)
			return errs.ErrInternalServer
		}

		uploadedDocs, err := s.uploadDocuments(txCtx, form.ID, formInput.Documents, domain.DocumentTypeAttachment)
		if err != nil {
			log.Printf("failed to upload documents for form: %s: %v", userID, err)
			return errs.ErrInternalServer
		}

		resultForm = model.FormWithDocs{Form: form, Documents: uploadedDocs}
		return nil
	}); err != nil {
		return nil, err
	}
	return &resultForm, nil
}

func (s *Service) GetForm(ctx context.Context, formID, requesterID uuid.UUID, requesterRole domain.Role) (*model.FormWithDocs, error) {
	form, err := s.formRepo.FindByID(ctx, formID)
	if err != nil {
		if errors.Is(err, errs.ErrFormNotFound) {
			return nil, errs.ErrFormNotFound
		}
		log.Printf("failed to find form: %v", err)
		return nil, errs.ErrInternalServer
	}

	var hasAccess bool
	switch requesterRole {
	case domain.RoleAdmin:
		hasAccess = true
	case domain.RoleHR:
		hasAccess = form.ExecutorID == requesterID
	case domain.RoleEmployee:
		hasAccess = form.UserID == requesterID
	}

	if !hasAccess {
		return nil, errs.ErrFormNotFound
	}

	docs, err := s.docRepo.FindByFormID(ctx, formID)
	if err != nil {
		log.Printf("failed to find documents for form %s: %v", formID, err)
		return nil, errs.ErrInternalServer
	}

	return &model.FormWithDocs{
		Form:      *form,
		Documents: docs,
	}, nil
}

func (s *Service) GetForms(ctx context.Context, filter *Filter, requesterID uuid.UUID, requesterRole domain.Role) ([]domain.Form, error) {
	// 1. Access control
	switch requesterRole {
	case domain.RoleEmployee:
		if filter.UserID != nil && *filter.UserID != requesterID {
			return nil, errs.ErrForbidden
		}
		filter.UserID = &requesterID
		filter.SortOrder = "DESC"

	case domain.RoleHR:
		filter.ExecutorID = &requesterID
		filter.SortOrder = "ASC"

	case domain.RoleAdmin:
		filter.SortOrder = "ASC"

	default:
		return nil, errs.ErrForbidden
	}

	// 2. Validate form status
	if err := filter.ValidateStatus(); err != nil {
		return nil, err
	}

	// 3. Get forms
	forms, err := s.formRepo.FindByFilter(ctx, filter)
	if err != nil {
		log.Printf("failed to find forms: %v", err)
		return nil, errs.ErrInternalServer
	}

	return forms, nil
}

func (s *Service) DownloadDocument(
	ctx context.Context,
	docID, requesterID uuid.UUID,
	requesterRole domain.Role,
) (*domain.Document, io.ReadCloser, error) {
	doc, err := s.docRepo.FindByID(ctx, docID)
	if err != nil {
		if errors.Is(err, errs.ErrDocumentNotFound) {
			return nil, nil, errs.ErrDocumentNotFound
		}
		log.Printf("failed to find document: %v", err)
		return nil, nil, errs.ErrInternalServer
	}

	// Access control: load form to check access
	form, err := s.formRepo.FindByID(ctx, doc.FormID)
	if err != nil {
		if errors.Is(err, errs.ErrFormNotFound) {
			return nil, nil, errs.ErrDocumentNotFound
		}
		log.Printf("failed to find form: %v", err)
		return nil, nil, errs.ErrInternalServer
	}

	var hasAccess bool
	switch requesterRole {
	case domain.RoleAdmin:
		hasAccess = true
	case domain.RoleHR:
		hasAccess = form.ExecutorID == requesterID
	case domain.RoleEmployee:
		hasAccess = form.UserID == requesterID
	}

	if !hasAccess {
		return nil, nil, errs.ErrDocumentNotFound
	}

	stream, err := s.storage.DownloadFile(ctx, doc.ObjectKey)
	if err != nil {
		log.Printf("failed to download file from storage: %v", err)
		return nil, nil, errs.ErrInternalServer
	}

	return doc, stream, nil
}

func (s *Service) Approve(ctx context.Context, formID, requesterID uuid.UUID, requesterRole domain.Role, comment string, docsInput []model.DocumentInput) (*domain.Form, error) {
	var resultForm *domain.Form
	if err := s.txMgr.Do(ctx, func(txCtx context.Context) error {
		form, err := s.formRepo.FindByID(txCtx, formID)
		if err != nil {
			if errors.Is(err, errs.ErrFormNotFound) {
				return errs.ErrFormNotFound
			}
			log.Printf("Failed to find form: %v", err)
			return errs.ErrInternalServer
		}

		var hasAccess bool
		switch requesterRole {
		case domain.RoleAdmin:
			hasAccess = true
		case domain.RoleHR:
			hasAccess = form.ExecutorID == requesterID
		case domain.RoleEmployee:
			hasAccess = form.UserID == requesterID
		}

		if !hasAccess {
			return errs.ErrForbidden
		}

		changed, err := form.ApproveForm(comment)
		if err != nil {
			return err
		}

		if changed {
			if err := s.formRepo.Update(txCtx, form); err != nil {
				log.Printf("Failed to update form: %v", err)
				return errs.ErrInternalServer
			}

			if _, err := s.uploadDocuments(txCtx, formID, docsInput, domain.DocumentTypeResult); err != nil {
				log.Printf("Failed to upload result documents: %v", err)
				return errs.ErrInternalServer
			}
		}

		resultForm = form
		return nil
	}); err != nil {
		return nil, err
	}

	return resultForm, nil
}

func (s *Service) Reject(ctx context.Context, formID uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role, comment string, docsInput []model.DocumentInput) (*domain.Form, error) {
	var resultForm *domain.Form
	if err := s.txMgr.Do(ctx, func(txCtx context.Context) error {
		form, err := s.formRepo.FindByID(txCtx, formID)
		if err != nil {
			if errors.Is(err, errs.ErrFormNotFound) {
				return errs.ErrFormNotFound
			}
			log.Printf("Failed to find form: %v", err)
			return errs.ErrInternalServer
		}

		var hasAccess bool
		switch requesterRole {
		case domain.RoleAdmin:
			hasAccess = true
		case domain.RoleHR:
			hasAccess = form.ExecutorID == requesterID
		case domain.RoleEmployee:
			hasAccess = form.UserID == requesterID
		}

		if !hasAccess {
			return errs.ErrForbidden
		}

		changed, err := form.RejectForm(comment)
		if err != nil {
			return err
		}

		if changed {
			if err := s.formRepo.Update(txCtx, form); err != nil {
				log.Printf("Failed to update form: %v", err)
				return errs.ErrInternalServer
			}

			if _, err := s.uploadDocuments(txCtx, formID, docsInput, domain.DocumentTypeResult); err != nil {
				log.Printf("Failed to upload result documents: %v", err)
				return errs.ErrInternalServer
			}
		}

		resultForm = form
		return nil
	}); err != nil {
		return nil, err
	}

	return resultForm, nil
}

func (s *Service) DeleteWithDocs(ctx context.Context, formID uuid.UUID, requesterRole domain.Role) error {
	switch requesterRole {
	case domain.RoleEmployee:
		return errs.ErrForbidden
	case domain.RoleHR:
		return errs.ErrForbidden
	case domain.RoleAdmin:
	default:
		return errs.ErrForbidden
	}

	// 1. Get a list of documents before deleting form
	docsIDs, err := s.docRepo.FindByFormID(ctx, formID)
	if err != nil {
		log.Printf("Failed to fetch documents for form %s before deletion: %v", formID, err)
		return errs.ErrInternalServer
	}

	// 2. Delete the application from the database
	if err := s.formRepo.Delete(ctx, formID); err != nil {
		if errors.Is(err, errs.ErrFormNotFound) {
			return errs.ErrFormNotFound
		}
		log.Printf("Failed to delete form: %v", err)
		return errs.ErrInternalServer
	}

	// 3. Deleting files from S3 storage
	for _, doc := range docsIDs {
		if err := s.storage.DeleteFile(context.Background(), doc.ObjectKey); err != nil {
			log.Printf("Failed to delete file %s from storage: %v", doc.ObjectKey, err)
		}
	}

	return nil
}

func (s *Service) uploadDocuments(ctx context.Context, formID uuid.UUID, docsInput []model.DocumentInput, docsType domain.DocumentType) ([]domain.Document, error) {
	var uploadedDocs []domain.Document
	for _, docInput := range docsInput {
		if docInput.Size > 10*1024*1024 {
			return nil, fmt.Errorf("file size is exceeded: %w", errs.ErrInvalidRequest)
		}

		if docInput.ContentType != "application/pdf" {
			return nil, fmt.Errorf("document %s: invalid content type: %w", docInput.Name, errs.ErrInvalidRequest)
		}

		var doc domain.Document
		switch docsType {
		case domain.DocumentTypeAttachment:
			doc = domain.NewAttachDocument(formID, docInput.Name)
		case domain.DocumentTypeResult:
			doc = domain.NewResultDocument(formID, docInput.Name)
		default:
			return nil, fmt.Errorf("invalid document type name: %w", errs.ErrInvalidRequest)
		}

		if err := s.storage.UploadFile(ctx, doc.ObjectKey, bytes.NewReader(docInput.Content), docInput.Size, docInput.ContentType); err != nil {
			log.Printf("failed to upload document %s: %v", doc.OriginalName, err)
			return nil, err
		}

		if err := s.docRepo.Create(ctx, &doc); err != nil {
			log.Printf("failed to save document record: %v", err)
			return nil, err
		}

		uploadedDocs = append(uploadedDocs, doc)
	}
	return uploadedDocs, nil
}
