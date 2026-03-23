package shop

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/darialissi/avito_merch_service/internal/models"
	"github.com/darialissi/avito_merch_service/lib/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

type ShopRepository struct {
	provider postgres.QueryEngineProvider
	sb       squirrel.StatementBuilderType
}

func NewShopRepository(provider postgres.QueryEngineProvider) *ShopRepository {
	return &ShopRepository{
		provider: provider,
		sb:       squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ShopRepository) GetUserCoinsByUsername(ctx context.Context, username string, forUpdate bool) (*models.User, error) {
	q := r.sb.
		Select(usersTableColumnID, usersTableColumnUsername, usersTableColumnCoins).
		From(usersTable).
		Where(squirrel.Eq{usersTableColumnUsername: username})

	if forUpdate {
		q = q.Suffix("FOR UPDATE")
	}

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

func (r *ShopRepository) GetSenderReceiverCoins(ctx context.Context, username1, username2 string, forUpdate bool) ([]models.User, error) {

	q := r.sb.
		Select(
			usersTableColumnID,
			usersTableColumnUsername,
			usersTableColumnCoins,
		).
		From(usersTable).
		Where(squirrel.Eq{
			usersTableColumnUsername: []string{username1, username2},
		})

	if forUpdate {
		q = q.Suffix("FOR UPDATE")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *ShopRepository) UpdateUserCoinsByUsername(ctx context.Context, username string, coins float64) (*models.User, error) {
	q := r.sb.
		Update(usersTable).
		Set(usersTableColumnCoins, coins).
		Where(squirrel.Eq{usersTableColumnUsername: username}).
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
		return nil, err
	}

	return &u, nil
}

func (r *ShopRepository) GetUserItemsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserItemExtended, error) {

	q := r.sb.
		Select(
			userItemsTablePrefix+"."+userItemsTableColumnUserID+" AS "+userItemsTableColumnUserID,
			userItemsTablePrefix+"."+userItemsTableColumnItemID+" AS "+userItemsTableColumnItemID,
			userItemsTablePrefix+"."+userItemsTableColumnQuantity+" AS "+userItemsTableColumnQuantity,
			itemsTablePrefix+"."+itemsTableColumnName+" AS "+itemsTableColumnName,
		).
		From(userItemsTable + " " + userItemsTablePrefix).
		Join(itemsTable + " " + itemsTablePrefix + " ON " + userItemsTablePrefix + "." + userItemsTableColumnItemID + " = " + itemsTablePrefix + "." + itemsTableColumnID).
		Where(squirrel.Eq{userItemsTablePrefix + "." + userItemsTableColumnUserID: userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.UserItemExtended])
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ShopRepository) GetReceivedTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	q := r.sb.
		Select(transactionsTableColumns...).
		From(transactionsTable).
		Where(squirrel.Eq{transactionsTableColumnToUserID: userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Transaction])
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *ShopRepository) GetSentTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	q := r.sb.
		Select(transactionsTableColumns...).
		From(transactionsTable).
		Where(squirrel.Eq{transactionsTableColumnFromUserID: userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Transaction])
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *ShopRepository) SaveTransaction(ctx context.Context, fromUserID uuid.UUID, toUserID uuid.UUID, amount float64) (*models.Transaction, error) {

	q := r.sb.
		Insert(transactionsTable).
		Columns(
			transactionsTableColumnFromUserID,
			transactionsTableColumnToUserID,
			transactionsTableColumnCoins,
		).
		Values(fromUserID, toUserID, amount).
		Suffix("RETURNING " + strings.Join(transactionsTableColumns, ","))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.Transaction])
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrUniqueConflict
		}
		return nil, err
	}

	return &t, nil
}

func (r *ShopRepository) GetItemByName(ctx context.Context, name string) (*models.Item, error) {
	q := r.sb.
		Select(itemsTableColumns...).
		From(itemsTable).
		Where(squirrel.Eq{itemsTableColumnName: name})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.Item])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *ShopRepository) GetUserItem(ctx context.Context, userID, itemID uuid.UUID, forUpdate bool) (*models.UserItem, error) {
	q := r.sb.
		Select(userItemsTableColumnUserID, userItemsTableColumnItemID, userItemsTableColumnQuantity).
		From(userItemsTable).
		Where(squirrel.And{
			squirrel.Eq{userItemsTableColumnUserID: userID},
			squirrel.Eq{userItemsTableColumnItemID: itemID},
		})

	if forUpdate {
		q = q.Suffix("FOR UPDATE")
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.UserItem])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *ShopRepository) UpdateUserItemQuantity(ctx context.Context, userID, itemID uuid.UUID, quantity int) (*models.UserItem, error) {
	q := r.sb.
		Update(userItemsTable).
		Set(userItemsTableColumnQuantity, quantity).
		Where(squirrel.And{
			squirrel.Eq{userItemsTableColumnUserID: userID},
			squirrel.Eq{userItemsTableColumnItemID: itemID},
		}).
		Suffix("RETURNING " + strings.Join(userItemsTableColumns, ","))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.UserItem])
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (r *ShopRepository) SaveUserItem(ctx context.Context, userID, itemID uuid.UUID, quantity int) (*models.UserItem, error) {

	q := r.sb.
		Insert(userItemsTable).
		Columns(
			userItemsTableColumnUserID,
			userItemsTableColumnItemID,
			userItemsTableColumnQuantity,
		).
		Values(userID, itemID, quantity).
		Suffix("RETURNING " + strings.Join(userItemsTableColumns, ","))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.provider.GetQueryEngine(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.UserItem])
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrUniqueConflict
		}
		return nil, err
	}

	return &u, nil
}
