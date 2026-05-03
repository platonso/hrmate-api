package form

import (
	"context"
	"errors"
	"fmt"
	"strings"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/repository/postgres/form/entity"
	formservice "github.com/platonso/hrmate-api/internal/service/form"
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

func (r *Repository) Create(ctx context.Context, form *domain.Form) error {
	rec := entity.ToFormRecord(*form)
	query := `
		INSERT INTO forms (id, user_id, executor_id, title, description, start_date, end_date, created_at, reviewed_at, status, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(
		ctx,
		query,
		rec.ID,
		rec.UserID,
		rec.ExecutorID,
		rec.Title,
		rec.Description,
		rec.StartDate,
		rec.EndDate,
		rec.CreatedAt,
		rec.ReviewedAt,
		rec.Status,
		rec.Comment,
	)
	return err
}

func (r *Repository) FindByID(ctx context.Context, formId uuid.UUID) (*domain.Form, error) {
	query := `
		SELECT id, user_id, executor_id, title, description, start_date, end_date, created_at, reviewed_at, status, comment
		FROM forms
		WHERE id = $1
`
	rec, err := r.findForm(ctx, query, formId)
	if err != nil {
		return nil, err
	}
	form := entity.ToDomainForm(rec)
	return &form, nil
}

func (r *Repository) FindByFilter(ctx context.Context, filter *formservice.Filter) ([]domain.Form, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, user_id, executor_id, title, description, start_date, end_date, created_at, reviewed_at, status, comment 
		FROM forms 
`)
	var conditions []string
	var args []any
	argPos := 1

	if filter != nil {
		if filter.UserID != nil {
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
			args = append(args, *filter.UserID)
			argPos++
		}

		if filter.ExecutorID != nil {
			conditions = append(conditions, fmt.Sprintf("executor_id = $%d", argPos))
			args = append(args, *filter.ExecutorID)
			argPos++
		}

		if filter.FormStatus != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
			args = append(args, string(*filter.FormStatus))
		}
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	sortOrder := "DESC"
	if filter != nil && filter.SortOrder != "" {
		order := strings.ToUpper(filter.SortOrder)
		if order == "ASC" || order == "DESC" {
			sortOrder = order
		}
	}

	queryBuilder.WriteString(" ORDER BY created_at ")
	queryBuilder.WriteString(sortOrder)

	records, err := r.findForms(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("find forms by filter: %w", err)
	}

	return entity.ToDomainForms(records), nil
}

func (r *Repository) Update(ctx context.Context, form *domain.Form) error {
	rec := entity.ToFormRecord(*form)
	query := `
	UPDATE forms 
	SET reviewed_at = $1, status = $2, comment = $3
	WHERE id = $4`

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	tag, err := conn.Exec(ctx, query, rec.ReviewedAt, rec.Status, rec.Comment, rec.ID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrFormNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, formID uuid.UUID) error {
	query := `
		DELETE FROM forms
		WHERE id = $1
	`

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	tag, err := conn.Exec(ctx, query, formID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrFormNotFound
	}

	return nil
}

func (r *Repository) findForm(ctx context.Context, query string, args ...any) (entity.FormRecord, error) {
	var rec entity.FormRecord
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, args...).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.ExecutorID,
		&rec.Title,
		&rec.Description,
		&rec.StartDate,
		&rec.EndDate,
		&rec.CreatedAt,
		&rec.ReviewedAt,
		&rec.Status,
		&rec.Comment,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.FormRecord{}, errs.ErrFormNotFound
		}
		return entity.FormRecord{}, err
	}
	return rec, nil
}

func (r *Repository) findForms(ctx context.Context, query string, args ...any) ([]entity.FormRecord, error) {
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query forms: %w", err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.FormRecord])
	if err != nil {
		return nil, fmt.Errorf("collect forms: %w", err)
	}

	return records, nil
}
