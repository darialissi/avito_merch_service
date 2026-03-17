package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	uc "github.com/darialissi/avito_merch_service/internal/usecases"
)

type AuthHandler struct {
	uc *uc.AuthUsecase
}

func NewAuthHandler(uc *uc.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		uc: uc,
	}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var form dto.AuthForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := form.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	auth, err := h.uc.SignIn(r.Context(), &form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(auth); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var form dto.AuthForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := form.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	auth, err := h.uc.LogIn(r.Context(), &form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	header := fmt.Sprintf("Bearer %s", auth.AccessToken)
	w.Header().Set("Authorization", header)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(auth); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
