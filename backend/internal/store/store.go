package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "sfa/backend/internal/db/sqlc"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *dbgen.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    pool,
		Queries: dbgen.New(pool),
	}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}

func (s *Store) WithTenantTx(ctx context.Context, tenantID uuid.UUID, fn func(*dbgen.Queries) error) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID.String()); err != nil {
		return fmt.Errorf("set tenant id: %w", err)
	}

	if err := fn(s.Queries.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
