package dto

import (
	"github.com/platonso/hrmate-api/internal/domain"
	"github.com/platonso/hrmate-api/internal/service/form/model"
)

func ToFormWithDocsResponse(m *model.FormWithDocs) FormWithDocsResponse {
	if m == nil {
		return FormWithDocsResponse{
			Documents: []DocumentResponse{},
		}
	}
	resp := FormWithDocsResponse{
		ID:          m.Form.ID,
		UserID:      m.Form.UserID,
		Title:       m.Form.Title,
		Description: m.Form.Description,
		StartDate:   m.Form.StartDate,
		EndDate:     m.Form.EndDate,
		CreatedAt:   m.Form.CreatedAt,
		Status:      string(m.Form.Status),
		Documents:   []DocumentResponse{},
	}

	if m.Form.Status != domain.StatusPending {
		resp.Resolution = &Resolution{
			Comment:      m.Form.Comment,
			ResolvedAt:   *m.Form.ReviewedAt,
			ResponseDocs: []DocumentResponse{},
		}
	}

	if len(m.Documents) > 0 {
		for _, d := range m.Documents {
			switch d.Type {
			case domain.DocumentTypeAttachment:
				resp.Documents = append(resp.Documents, toDocumentResponse(d))
			case domain.DocumentTypeResult:
				resp.Resolution.ResponseDocs = append(resp.Resolution.ResponseDocs, toDocumentResponse(d))
			}
		}
	}
	return resp
}

func toDocumentResponse(doc domain.Document) DocumentResponse {
	return DocumentResponse{
		ID:   doc.ID,
		Name: doc.OriginalName,
	}
}

func ToFormResponse(f *domain.Form) FormResponse {
	return FormResponse{
		ID:     f.ID,
		UserID: f.UserID,
		Title:  f.Title,
		//Description: form.Description,
		//StartDate: form.StartDate,
		//EndDate:   form.EndDate,
		CreatedAt: f.CreatedAt,
		Status:    string(f.Status),
	}
}

func ToFormsResponse(forms []domain.Form) []FormResponse {
	if len(forms) == 0 {
		return []FormResponse{}
	}
	responses := make([]FormResponse, len(forms))
	for i := range forms {
		responses[i] = ToFormResponse(&forms[i])
	}
	return responses
}
