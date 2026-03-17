package http

import (
	"net/http"

	uc "github.com/darialissi/avito_merch_service/internal/usecases"
)

type ShopHandler struct {
	uc *uc.ShopUsecase
}

func NewShopHandler(uc *uc.ShopUsecase) *ShopHandler {
	return &ShopHandler{
		uc: uc,
	}
}

func (h *ShopHandler) Info(w http.ResponseWriter, r *http.Request) {}

func (h *ShopHandler) SendCoin(w http.ResponseWriter, r *http.Request) {}

func (h *ShopHandler) BuyItem(w http.ResponseWriter, r *http.Request) {}
