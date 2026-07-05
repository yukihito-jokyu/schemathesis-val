package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
)

type MemoryStore struct {
	mu          sync.RWMutex
	users       map[int]model.User
	items       map[string]model.Item
	orders      map[string]model.Order
	nextUserID  int
	nextItemIdx int
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		users:       make(map[int]model.User),
		items:       make(map[string]model.Item),
		orders:      make(map[string]model.Order),
		nextUserID:  1,
		nextItemIdx: 1,
	}
	store.seed()
	return store
}

func (s *MemoryStore) seed() {
	s.items["item_1"] = model.Item{
		ID:       "item_1",
		Name:     "Go Programming Book",
		Category: model.CategoryBook,
		Price:    3500.0,
	}
	s.items["item_2"] = model.Item{
		ID:       "item_2",
		Name:     "Organic Apple",
		Category: model.CategoryFood,
		Price:    200.0,
	}
	s.items["item_3"] = model.Item{
		ID:       "item_3",
		Name:     "Screwdriver",
		Category: model.CategoryTool,
		Price:    1200.0,
	}
	s.nextItemIdx = 4
}

// User CRUD
func (s *MemoryStore) ListUsers(limit int, role model.UserRole) []model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.User
	for _, u := range s.users {
		if role != "" && u.Role != role {
			continue
		}
		result = append(result, u)
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (s *MemoryStore) EmailExists(email string, excludeID int) bool {
	for _, u := range s.users {
		if u.ID != excludeID && u.Email == email {
			return true
		}
	}
	return false
}

func (s *MemoryStore) CreateUser(name string, email string, age int, role model.UserRole) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.EmailExists(email, 0) {
		return model.User{}, errors.New("email already exists")
	}

	user := model.User{
		ID:        s.nextUserID,
		Name:      name,
		Email:     email,
		Age:       age,
		Role:      role,
		CreatedAt: time.Now(),
	}
	s.users[user.ID] = user
	s.nextUserID++

	return user, nil
}

func (s *MemoryStore) GetUser(id int) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

func (s *MemoryStore) UpdateUser(id int, req model.UpdateUserRequest) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return model.User{}, errors.New("user not found")
	}

	if req.Email != nil && s.EmailExists(*req.Email, id) {
		return model.User{}, errors.New("email already exists")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Age != nil {
		user.Age = *req.Age
	}
	if req.Role != nil {
		user.Role = *req.Role
	}

	s.users[id] = user
	return user, nil
}

func (s *MemoryStore) DeleteUser(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.users[id]
	if !ok {
		return false
	}
	delete(s.users, id)
	return true
}

// Item CRUD
func (s *MemoryStore) ListItems(category model.ItemCategory, minPrice, maxPrice *float64) []model.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.Item
	for _, item := range s.items {
		if category != "" && item.Category != category {
			continue
		}
		if minPrice != nil && item.Price < *minPrice {
			continue
		}
		if maxPrice != nil && item.Price > *maxPrice {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (s *MemoryStore) CreateItem(name string, category model.ItemCategory, price float64) model.Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := model.Item{
		ID:       fmt.Sprintf("item_%d", s.nextItemIdx),
		Name:     name,
		Category: category,
		Price:    price,
	}
	s.items[item.ID] = item
	s.nextItemIdx++
	return item
}

func (s *MemoryStore) GetItem(id string) (model.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

// Order CRUD
func (s *MemoryStore) CreateOrder(userID int, itemIDs []string, quantity int) (model.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify user exists
	if _, ok := s.users[userID]; !ok {
		return model.Order{}, errors.New("user not found")
	}

	// Verify all items exist
	for _, itemID := range itemIDs {
		if _, ok := s.items[itemID]; !ok {
			return model.Order{}, fmt.Errorf("item not found: %s", itemID)
		}
	}

	order := model.Order{
		ID:       uuid.New().String(),
		UserID:   userID,
		ItemIDs:  itemIDs,
		Quantity: quantity,
		Status:   model.StatusCreated,
	}
	s.orders[order.ID] = order
	return order, nil
}

func (s *MemoryStore) GetOrder(id string) (model.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[id]
	return order, ok
}
