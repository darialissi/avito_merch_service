package middleware

import (
	"fmt"
	"net/http"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Panic recovered:", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
