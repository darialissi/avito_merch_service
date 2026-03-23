package http

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"

	"github.com/darialissi/avito_merch_service/internal/contextkeys"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
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

func (h *ShopHandler) SendCoin(w http.ResponseWriter, r *http.Request) {

	username, ok := r.Context().Value(contextkeys.UserKey).(string)
	if !ok || username == "" {
		writeJSONError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req dto.TransactionData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if err := req.Validate(username); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.uc.SendCoin(r.Context(), username, &req); err != nil {
		if errors.Is(err, uc.ErrUserNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, uc.ErrNotEnoughCoins) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ShopHandler) BuyItem(w http.ResponseWriter, r *http.Request) {

	username, ok := r.Context().Value(contextkeys.UserKey).(string)
	if !ok || username == "" {
		http.Error(w, ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	item := chi.URLParam(r, "item")
	if item == "" {
		writeJSONError(w, http.StatusBadRequest, ErrRequiredField+" 'item'")
		return
	}

	var req dto.BuyItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	data := &dto.BuyItemData{
		ItemName: strings.ToLower(item),
		Quantity: req.Quantity,
	}
	if err := data.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.uc.BuyItem(r.Context(), username, data); err != nil {
		if errors.Is(err, uc.ErrItemNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, uc.ErrNotEnoughCoins) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ShopHandler) Info(w http.ResponseWriter, r *http.Request) {

	username, ok := r.Context().Value(contextkeys.UserKey).(string)
	if !ok || username == "" {
		writeJSONError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	info, err := h.uc.Info(r.Context(), username)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
}
