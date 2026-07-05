package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
	"github.com/yukihito-jokyu/schemathesis-val/internal/response"
	"github.com/yukihito-jokyu/schemathesis-val/internal/store"
)

type OrderHandler struct {
	store *store.MemoryStore
}

func NewOrderHandler(s *store.MemoryStore) *OrderHandler {
	return &OrderHandler{store: s}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if _, ok := response.DecodeAndValidate(w, r, &req, false); !ok {
		return
	}

	var details []string
	if req.UserID == nil {
		details = append(details, "userId is required")
	} else if *req.UserID < 1 {
		details = append(details, "userId must be greater than or equal to 1")
	}

	if req.ItemIDs == nil {
		details = append(details, "itemIds is required")
	} else {
		items := *req.ItemIDs
		if len(items) < 1 || len(items) > 5 {
			details = append(details, "itemIds length must be between 1 and 5")
		}
		itemPattern := regexp.MustCompile("^item_[0-9]+$")
		for idx, itemID := range items {
			if !itemPattern.MatchString(itemID) {
				details = append(details, fmt.Sprintf("itemIds[%d] must match pattern '^item_[0-9]+$'", idx))
			}
		}
	}

	if req.Quantity == nil {
		details = append(details, "quantity is required")
	} else if *req.Quantity < 1 || *req.Quantity > 10 {
		details = append(details, "quantity must be between 1 and 10")
	}

	if len(details) > 0 {
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", details)
		return
	}

	order, err := h.store.CreateOrder(*req.UserID, *req.ItemIDs, *req.Quantity)
	if err != nil {
		if err.Error() == "user not found" || strings.HasPrefix(err.Error(), "item not found") {
			response.Error(w, http.StatusNotFound, model.CodeNotFound, "resource not found", []string{err.Error()})
			return
		}
		response.Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "order creation failed", []string{err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "orderId")

	if _, err := uuid.Parse(id); err != nil {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "order not found", []string{"orderId must be a valid UUID"})
		return
	}

	order, ok := h.store.GetOrder(id)
	if !ok {
		response.Error(w, http.StatusNotFound, model.CodeNotFound, "order not found", []string{fmt.Sprintf("order with ID %s does not exist", id)})
		return
	}

	response.JSON(w, http.StatusOK, order)
}
