package shop

import (
	"github.com/Masterminds/squirrel"
	"github.com/darialissi/avito_merch_service/lib/postgres"
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
