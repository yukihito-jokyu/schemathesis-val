package handler

import (
	"net/http"

	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	response.JSON(w, http.StatusOK, resp)
}
