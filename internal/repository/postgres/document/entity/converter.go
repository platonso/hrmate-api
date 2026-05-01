package entity

import "github.com/platonso/hrmate-api/internal/domain"

func ToDomainDocument(rec DocumentRecord) domain.Document {
	return domain.Document{
		ID:           rec.ID,
		FormID:       rec.FormID,
		ObjectKey:    rec.ObjectKey,
		OriginalName: rec.OriginalName,
		UploadedAt:   rec.UploadedAt,
		Type:         domain.DocumentType(rec.Type),
	}
}

func ToDomainDocuments(records []DocumentRecord) []domain.Document {
	docs := make([]domain.Document, len(records))
	for i, r := range records {
		docs[i] = ToDomainDocument(r)
	}
	return docs
}

func ToDocumentRecord(d domain.Document) DocumentRecord {
	return DocumentRecord{
		ID:           d.ID,
		FormID:       d.FormID,
		ObjectKey:    d.ObjectKey,
		OriginalName: d.OriginalName,
		UploadedAt:   d.UploadedAt,
		Type:         string(d.Type),
	}
}
