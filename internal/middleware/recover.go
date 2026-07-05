package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
)

func RecoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				response.Error(
					w,
					http.StatusInternalServerError,
					model.CodeInternalError,
					"internal server error",
					[]string{fmt.Sprintf("%v", err)},
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
