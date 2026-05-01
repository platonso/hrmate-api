package document

import (
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/repository/postgres/document/entity"
)

type Repository struct {
	db        *pgxpool.Pool
	ctxGetter *trmpgx.CtxGetter
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:        db,
		ctxGetter: trmpgx.DefaultCtxGetter,
	}
}

func (r *Repository) Create(ctx context.Context, doc *domain.Document) error {
	rec := entity.ToDocumentRecord(*doc)
	query := `
		INSERT INTO documents (id, form_id, object_key, original_name, uploaded_at, type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(
		ctx,
		query,
		rec.ID,
		rec.FormID,
		rec.ObjectKey,
		rec.OriginalName,
		rec.UploadedAt,
		rec.Type,
	)
	return err
}

func (r *Repository) FindByFormID(ctx context.Context, formID uuid.UUID) ([]domain.Document, error) {
	query := `
		SELECT id, form_id, object_key, original_name, uploaded_at, type
		FROM documents
		WHERE form_id = $1
		ORDER BY uploaded_at
	`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, formID)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.DocumentRecord])
	if err != nil {
		return nil, fmt.Errorf("collect documents: %w", err)
	}

	return entity.ToDomainDocuments(records), nil
}

func (r *Repository) FindByID(ctx context.Context, docID uuid.UUID) (*domain.Document, error) {
	query := `
		SELECT id, form_id, object_key, original_name, uploaded_at, type
		FROM documents
		WHERE id = $1
	`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)
	var rec entity.DocumentRecord
	err := conn.QueryRow(ctx, query, docID).Scan(
		&rec.ID,
		&rec.FormID,
		&rec.ObjectKey,
		&rec.OriginalName,
		&rec.UploadedAt,
		&rec.Type,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrDocumentNotFound
		}
		return nil, err
	}

	doc := entity.ToDomainDocument(rec)
	return &doc, nil
}
