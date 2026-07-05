package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
)

type BugHandler struct{}

func NewBugHandler() *BugHandler {
	return &BugHandler{}
}

func (h *BugHandler) SchemaMismatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	mismatchJSON := `{
		"id": "not-integer",
		"name": 123,
		"email": "invalid",
		"role": "super-admin"
	}`
	_, _ = w.Write([]byte(mismatchJSON))
}

func (h *BugHandler) StatusMismatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTeapot)
	teapotJSON := `{
		"code": "teapot",
		"message": "unexpected status"
	}`
	_, _ = w.Write([]byte(teapotJSON))
}

type PanicOnZeroRequest struct {
	Value *int `json:"value"`
}

func (h *BugHandler) PanicOnZero(w http.ResponseWriter, r *http.Request) {
	// First, decode as raw map to check for null values on required fields
	var raw map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&raw); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"invalid JSON"}`))
		return
	}

	// Check if value is present and not null (it's required)
	val, exists := raw["value"]
	if !exists || val == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"value is required and must not be null"}`))
		return
	}

	// Convert value to int
	valFloat, ok := val.(float64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"value must be an integer"}`))
		return
	}

	intVal := int(valFloat)

	// Validate bounds minimum: -10, maximum: 10 as specified in schema
	if intVal < -10 || intVal > 10 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"value must be between -10 and 10"}`))
		return
	}

	if intVal == 0 {
		panic("zero is not allowed")
	}

	resp := map[string]int{"result": intVal * 2}
	response.JSON(w, http.StatusOK, resp)
}

func (h *BugHandler) InvalidEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	invalidEmailJSON := `{
		"id": 1,
		"name": "Broken User",
		"email": "not-an-email",
		"age": 20,
		"role": "member",
		"createdAt": "2026-07-05T10:00:00Z"
	}`
	_, _ = w.Write([]byte(invalidEmailJSON))
}
