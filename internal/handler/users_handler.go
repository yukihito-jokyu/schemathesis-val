package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
	"github.com/yukihito-jokyu/schemathesis-val/internal/store"
)

type UserHandler struct {
	store *store.MemoryStore
}

func NewUserHandler(s *store.MemoryStore) *UserHandler {
	return &UserHandler{store: s}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	for k := range r.URL.Query() {
		if k != "limit" && k != "role" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid query parameter", []string{fmt.Sprintf("unknown query parameter: %s", k)})
			return
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if _, ok := r.URL.Query()["limit"]; ok {
		if limitStr == "" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid limit parameter", []string{"limit must be an integer between 1 and 100"})
			return
		}
		val, err := strconv.Atoi(limitStr)
		if err != nil || val < 1 || val > 100 {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid limit parameter", []string{"limit must be an integer between 1 and 100"})
			return
		}
		limit = val
	}

	roleStr := r.URL.Query().Get("role")
	var role model.UserRole
	if _, ok := r.URL.Query()["role"]; ok {
		if roleStr == "" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid role parameter", []string{"role must be admin, member, or guest"})
			return
		}
		role = model.UserRole(roleStr)
		if role != model.RoleAdmin && role != model.RoleMember && role != model.RoleGuest {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid role parameter", []string{"role must be admin, member, or guest"})
			return
		}
	}

	users := h.store.ListUsers(limit, role)
	if users == nil {
		users = []model.User{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": users})
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if _, ok := response.DecodeAndValidate(w, r, &req, true); !ok {
		return
	}

	var details []string
	if req.Name == nil {
		details = append(details, "name is required")
	} else if utf8.RuneCountInString(*req.Name) < 1 || utf8.RuneCountInString(*req.Name) > 50 {
		details = append(details, "name length must be between 1 and 50")
	}

	if req.Email == nil {
		details = append(details, "email is required")
	} else if !strings.Contains(*req.Email, "@") {
		details = append(details, "email must be a valid email format")
	}

	if req.Age == nil {
		details = append(details, "age is required")
	} else if *req.Age < 0 || *req.Age > 120 {
		details = append(details, "age must be between 0 and 120")
	}

	if req.Role == nil {
		details = append(details, "role is required")
	} else {
		role := *req.Role
		if role != model.RoleAdmin && role != model.RoleMember && role != model.RoleGuest {
			details = append(details, "role must be admin, member, or guest")
		}
	}

	if len(details) > 0 {
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", details)
		return
	}

	user, err := h.store.CreateUser(*req.Name, *req.Email, *req.Age, *req.Role)
	if err != nil {
		response.Error(w, http.StatusConflict, model.CodeConflict, "user creation failed", []string{err.Error()})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
	response.JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{"userId must be a positive integer"})
		return
	}

	user, ok := h.store.GetUser(id)
	if !ok {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{fmt.Sprintf("user with ID %d does not exist", id)})
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{"userId must be a positive integer"})
		return
	}

	var req model.UpdateUserRequest
	if _, ok := response.DecodeAndValidate(w, r, &req, false); !ok {
		return
	}

	if req.Name == nil && req.Email == nil && req.Age == nil && req.Role == nil {
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", []string{"at least one field (name, email, age, role) must be provided"})
		return
	}

	var details []string
	if req.Name != nil && (utf8.RuneCountInString(*req.Name) < 1 || utf8.RuneCountInString(*req.Name) > 50) {
		details = append(details, "name length must be between 1 and 50")
	}
	if req.Email != nil && !strings.Contains(*req.Email, "@") {
		details = append(details, "email must be a valid email format")
	}
	if req.Age != nil && (*req.Age < 0 || *req.Age > 120) {
		details = append(details, "age must be between 0 and 120")
	}
	if req.Role != nil {
		role := *req.Role
		if role != model.RoleAdmin && role != model.RoleMember && role != model.RoleGuest {
			details = append(details, "role must be admin, member, or guest")
		}
	}

	if len(details) > 0 {
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", details)
		return
	}

	user, err := h.store.UpdateUser(id, req)
	if err != nil {
		if err.Error() == "user not found" {
			response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{err.Error()})
		} else {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "user update failed", []string{err.Error()})
		}
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{"userId must be a positive integer"})
		return
	}

	ok := h.store.DeleteUser(id)
	if !ok {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "user not found", []string{fmt.Sprintf("user with ID %d does not exist", id)})
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	dummyMe := model.User{
		ID:        999,
		Name:      "Authorized User",
		Email:     "auth-user@example.com",
		Age:       30,
		Role:      model.RoleMember,
		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
	}
	response.JSON(w, http.StatusOK, dummyMe)
}
