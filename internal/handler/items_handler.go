package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
	"github.com/yukihito-jokyu/schemathesis-val/internal/store"
)

type ItemHandler struct {
	store *store.MemoryStore
}

func NewItemHandler(s *store.MemoryStore) *ItemHandler {
	return &ItemHandler{store: s}
}

func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	for k := range r.URL.Query() {
		if k != "category" && k != "minPrice" && k != "maxPrice" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid query parameter", []string{fmt.Sprintf("unknown query parameter: %s", k)})
			return
		}
	}

	categoryStr := r.URL.Query().Get("category")
	if _, ok := r.URL.Query()["category"]; ok {
		if categoryStr == "" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid category parameter", []string{"category must be book, food, or tool"})
			return
		}
		category := model.ItemCategory(categoryStr)
		if category != model.CategoryBook && category != model.CategoryFood && category != model.CategoryTool {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid category parameter", []string{"category must be book, food, or tool"})
			return
		}
	}

	var category model.ItemCategory
	if categoryStr != "" {
		category = model.ItemCategory(categoryStr)
	}

	var minPrice *float64
	if _, ok := r.URL.Query()["minPrice"]; ok {
		minPriceStr := r.URL.Query().Get("minPrice")
		if minPriceStr == "" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid minPrice parameter", []string{"minPrice must be a number greater than or equal to 0"})
			return
		}
		val, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil || val < 0 {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid minPrice parameter", []string{"minPrice must be a number greater than or equal to 0"})
			return
		}
		minPrice = &val
	}

	var maxPrice *float64
	if _, ok := r.URL.Query()["maxPrice"]; ok {
		maxPriceStr := r.URL.Query().Get("maxPrice")
		if maxPriceStr == "" {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid maxPrice parameter", []string{"maxPrice must be a number less than or equal to 10000"})
			return
		}
		val, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil || val > 10000 {
			response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "invalid maxPrice parameter", []string{"maxPrice must be a number less than or equal to 10000"})
			return
		}
		maxPrice = &val
	}

	items := h.store.ListItems(category, minPrice, maxPrice)
	if items == nil {
		items = []model.Item{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req model.CreateItemRequest
	if _, ok := response.DecodeAndValidate(w, r, &req, false); !ok {
		return
	}

	var details []string
	if req.Name == nil {
		details = append(details, "name is required")
	} else if utf8.RuneCountInString(*req.Name) < 1 || utf8.RuneCountInString(*req.Name) > 80 {
		details = append(details, "name length must be between 1 and 80")
	}

	if req.Category == nil {
		details = append(details, "category is required")
	} else {
		cat := *req.Category
		if cat != model.CategoryBook && cat != model.CategoryFood && cat != model.CategoryTool {
			details = append(details, "category must be book, food, or tool")
		}
	}

	if req.Price == nil {
		details = append(details, "price is required")
	} else if *req.Price < 0 || *req.Price > 10000 {
		details = append(details, "price must be between 0 and 10000")
	}

	if len(details) > 0 {
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", details)
		return
	}

	item := h.store.CreateItem(*req.Name, *req.Category, *req.Price)
	response.JSON(w, http.StatusCreated, item)
}

func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "itemId")

	matched, _ := regexp.MatchString("^item_[0-9]+$", id)
	if !matched {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "item not found", []string{"itemId must match pattern '^item_[0-9]+$'"})
		return
	}

	item, ok := h.store.GetItem(id)
	if !ok {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "item not found", []string{fmt.Sprintf("item with ID %s does not exist", id)})
		return
	}

	response.JSON(w, http.StatusOK, item)
}
