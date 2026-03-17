package usecases

import (
	"github.com/darialissi/avito_merch_service/internal/repositories/shop"
)

type ShopUsecase struct {
	repo *shop.ShopRepository
}

func NewShopUsecase(repo *shop.ShopRepository) *ShopUsecase {
	return &ShopUsecase{
		repo: repo,
	}
}
