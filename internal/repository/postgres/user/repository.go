package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/hrmate-api/internal/domain"
	errs "github.com/platonso/hrmate-api/internal/errors"
	"github.com/platonso/hrmate-api/internal/repository/postgres/user/entity"
	"github.com/platonso/hrmate-api/internal/service/assignment"
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

func (r *Repository) CheckSchema(ctx context.Context) error {
	var exists bool
	query := `
        SELECT EXISTS (
        	SELECT FROM information_schema.tables 
        	WHERE table_name = 'users'
        )
    `
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	if err := conn.QueryRow(ctx, query).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check schema: %w", err)
	}

	if !exists {
		return fmt.Errorf("required table 'users' does not exist")
	}

	return nil
}

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	rec := entity.ToUserRecord(*user)
	query := `
		INSERT INTO users (id, user_role, first_name, last_name, position, email, hashed_password, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(
		ctx,
		query,
		rec.ID,
		rec.Role,
		rec.FirstName,
		rec.LastName,
		rec.Position,
		rec.Email,
		rec.HashedPassword,
		rec.IsActive,
	)
	return err
}

func (r *Repository) FindByID(ctx context.Context, userId uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, user_role, first_name, last_name, position, email, hashed_password, is_active
		FROM users
		WHERE id = $1		
`
	rec, err := r.findUser(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	user := entity.ToDomainUser(rec)
	return &user, nil
}

func (r *Repository) FindByIDs(ctx context.Context, userIDs []uuid.UUID) ([]domain.User, error) {
	if len(userIDs) == 0 {
		return []domain.User{}, nil
	}

	query := `
		SELECT id, user_role, first_name, last_name, position, email, hashed_password, is_active
		FROM users
		WHERE id = ANY($1)
	`

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, query, userIDs)
	if err != nil {
		log.Printf("query failed: %v", err)
		return nil, errs.ErrInternalServer
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.UserRecord])
	if err != nil {
		log.Printf("collect rows: %v", err)
		return nil, errs.ErrInternalServer
	}

	if len(records) == 0 {
		return []domain.User{}, nil
	}

	return entity.ToDomainUsers(records), nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, user_role, first_name, last_name, position, email, hashed_password, is_active
		FROM users
		WHERE email = $1		
`
	rec, err := r.findUser(ctx, query, email)
	if err != nil {
		return nil, err
	}
	user := entity.ToDomainUser(rec)
	return &user, nil
}

func (r *Repository) Update(ctx context.Context, user *domain.User) error {
	rec := entity.ToUserRecord(*user)
	query := `
        UPDATE users SET
            user_role = $1,
            first_name = $2,
            last_name = $3,
            position = $4,
            email = $5,
            hashed_password = $6,
            is_active = $7
        WHERE id = $8
    `

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	tag, err := conn.Exec(ctx, query,
		rec.Role,
		rec.FirstName,
		rec.LastName,
		rec.Position,
		rec.Email,
		rec.HashedPassword,
		rec.IsActive,
		rec.ID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}

func (r *Repository) FindByRole(ctx context.Context, roles ...domain.Role) ([]domain.User, error) {
	if len(roles) == 0 {
		return []domain.User{}, nil
	}

	query := `
		SELECT id, user_role, first_name, last_name, position, email, hashed_password, is_active
		FROM users
		WHERE user_role = ANY($1)
`
	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, query, roles)
	if err != nil {
		log.Printf("failed to query users by roles %v: %v", roles, err)
		return nil, errs.ErrInternalServer
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.UserRecord])
	if err != nil {
		log.Printf("collect rows: %v", err)
		return nil, errs.ErrInternalServer
	}

	if len(records) == 0 {
		return []domain.User{}, nil
	}

	users := entity.ToDomainUsers(records)
	return users, nil
}

func (r *Repository) IsActive(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT is_active FROM users WHERE id = $1`

	var active bool

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	err := conn.QueryRow(ctx, query, userID).Scan(&active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, errs.ErrUserNotFound
		}
		return false, err
	}

	return active, nil
}

func (r *Repository) findUser(ctx context.Context, query string, args ...any) (entity.UserRecord, error) {
	var rec entity.UserRecord

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	err := conn.QueryRow(ctx, query, args...).Scan(
		&rec.ID,
		&rec.Role,
		&rec.FirstName,
		&rec.LastName,
		&rec.Position,
		&rec.Email,
		&rec.HashedPassword,
		&rec.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.UserRecord{}, errs.ErrUserNotFound
		}
		return entity.UserRecord{}, err
	}

	return rec, nil
}

func (r *Repository) FindActiveHRsWithWorkload(ctx context.Context) ([]assignment.HRWorkload, error) {
	query := `
				SELECT 
					u.id,
					COALESCE(COUNT(f.id) FILTER (WHERE f.status = 'pending'), 0) AS pending_forms_count
				FROM users u
				LEFT JOIN forms f ON f.executor_id = u.id
				WHERE u.user_role = 'hr' AND u.is_active = true
				GROUP BY u.id
				ORDER BY pending_forms_count , u.id
			`

	conn := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		log.Printf("failed to query active HRs with workload: %v", err)
		return nil, errs.ErrInternalServer
	}
	defer rows.Close()

	var results []assignment.HRWorkload
	for rows.Next() {
		var hw assignment.HRWorkload
		if err := rows.Scan(&hw.UserID, &hw.PendingFormsCount); err != nil {
			log.Printf("failed to scan HR workload: %v", err)
			return nil, errs.ErrInternalServer
		}
		results = append(results, hw)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v", err)
		return nil, errs.ErrInternalServer
	}

	return results, nil
}
