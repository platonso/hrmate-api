package postgres

import (
	"context"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/hrmate-api/internal/repository/postgres/document"
	"github.com/platonso/hrmate-api/internal/repository/postgres/form"
	"github.com/platonso/hrmate-api/internal/repository/postgres/user"
)

type Repository struct {
	Users     *user.Repository
	Forms     *form.Repository
	Documents *document.Repository
	pool      *pgxpool.Pool
}

func NewRepository(ctx context.Context, connStr string) (*Repository, *manager.Manager, error) {
	pool, err := newPool(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	txMgr := manager.Must(trmpgx.NewDefaultFactory(pool))

	repo := &Repository{
		Users:     user.NewRepository(pool),
		Forms:     form.NewRepository(pool),
		Documents: document.NewRepository(pool),
		pool:      pool,
	}

	return repo, txMgr, nil
}

func newPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
