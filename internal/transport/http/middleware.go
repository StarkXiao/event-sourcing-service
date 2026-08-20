package httpapi

import (
	"github.com/google/uuid"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", uuid.NewString())
		next.ServeHTTP(w, r)
	})
}
