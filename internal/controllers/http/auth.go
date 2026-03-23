package http

import (
	"encoding/json"
	"errors"
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

	var form dto.AuthForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if err := form.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	auth, err := h.uc.SignIn(r.Context(), &form)
	if err != nil {
		if errors.Is(err, uc.ErrUsernameAlreadyExists) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(auth); err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
}

func (h *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) {

	var form dto.AuthForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if err := form.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	auth, err := h.uc.LogIn(r.Context(), &form)
	if err != nil {
		if errors.Is(err, uc.ErrNotExistedUser) || errors.Is(err, uc.ErrIncorrectPassword) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	header := fmt.Sprintf("Bearer %s", auth.AccessToken)
	w.Header().Set("Authorization", header)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(auth); err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
}
