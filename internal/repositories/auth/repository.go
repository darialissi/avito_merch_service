package auth

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/darialissi/avito_merch_service/internal/models"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/darialissi/avito_merch_service/lib/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

type AuthRepository struct {
	provider postgres.QueryEngineProvider
	sb       squirrel.StatementBuilderType
}

func NewAuthRepository(provider postgres.QueryEngineProvider) *AuthRepository {
	return &AuthRepository{
		provider: provider,
		sb:       squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *AuthRepository) SaveUser(ctx context.Context, data *dto.UserForm) (*models.User, error) {

	q := r.sb.
		Insert(usersTable).
		Columns(
			usersTableColumnUsername,
			usersTableColumnHashedPassword,
		).
		Values(data.Username, data.HashedPassword).
		Suffix("RETURNING " + strings.Join(usersTableColumns, ","))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrUniqueConflict
		}
		return nil, err
	}

	return &u, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	q := r.sb.
		Select(usersTableColumnID, usersTableColumnUsername, usersTableColumnHashedPassword, usersTableColumnCreatedAt).
		From(usersTable).
		Where(squirrel.Eq{usersTableColumnID: id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	q := r.sb.
		Select(usersTableColumnID, usersTableColumnUsername, usersTableColumnHashedPassword, usersTableColumnCreatedAt).
		From(usersTable).
		Where(squirrel.Eq{usersTableColumnUsername: username})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		return nil, err
	}

	return &u, nil
}
