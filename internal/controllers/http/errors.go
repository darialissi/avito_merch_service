package http

import (
	"encoding/json"
	"net/http"
)

var (
	ErrBadRequest          = "Bad Request"
	ErrUnauthorized        = "Unauthorized"
	ErrInternalServerError = "Internal Server Error"
	ErrRequiredField       = "Field is required"
)

type ErrorResponse struct {
	Errors string `json:"errors"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Errors: msg,
	})
}
