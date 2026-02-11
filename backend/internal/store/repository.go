package store

import (
	"context"

	"github.com/google/uuid"

	dbgen "sfa/backend/internal/db/sqlc"
)

// Repository exposes sqlc-generated queries plus tenant-scoped transaction helper.
type Repository interface {
	Querier() dbgen.Querier
	WithTenantTx(ctx context.Context, tenantID uuid.UUID, fn func(*dbgen.Queries) error) error
}

type SQLCRepository struct {
	store *Store
}

func NewRepository(store *Store) SQLCRepository {
	return SQLCRepository{store: store}
}

func (r SQLCRepository) Querier() dbgen.Querier {
	return r.store.Queries
}

func (r SQLCRepository) WithTenantTx(ctx context.Context, tenantID uuid.UUID, fn func(*dbgen.Queries) error) error {
	return r.store.WithTenantTx(ctx, tenantID, fn)
}
