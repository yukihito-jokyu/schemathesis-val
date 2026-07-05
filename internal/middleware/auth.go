package middleware

import (
	"net/http"
	"strings"

	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
)

func BearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing authorization header", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] != token {
				response.Error(w, http.StatusUnauthorized, model.CodeUnauthorized, "invalid or expired token", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
