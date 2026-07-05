package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yukihito-jokyu/schemathesis-val/internal/handler"
	customMiddleware "github.com/yukihito-jokyu/schemathesis-val/internal/middleware"
	"github.com/yukihito-jokyu/schemathesis-val/internal/store"
)

func allowedMethods(path string) string {
	if path == "/health" {
		return "GET"
	}
	if path == "/users" {
		return "GET, POST, OPTIONS"
	}
	if strings.HasPrefix(path, "/users/") {
		return "GET, PUT, DELETE, OPTIONS"
	}
	if path == "/items" {
		return "GET, POST, OPTIONS"
	}
	if strings.HasPrefix(path, "/items/") {
		return "GET, OPTIONS"
	}
	if path == "/orders" {
		return "POST, OPTIONS"
	}
	if strings.HasPrefix(path, "/orders/") {
		return "GET, OPTIONS"
	}
	if path == "/me" {
		return "GET, OPTIONS"
	}
	if path == "/bugs/panic-on-zero" {
		return "POST, OPTIONS"
	}
	if strings.HasPrefix(path, "/bugs/") {
		return "GET, OPTIONS"
	}
	return "GET, POST, PUT, DELETE, OPTIONS"
}

func main() {
	log.Println("Starting Schemathesis Go Verification API...")

	store := store.NewMemoryStore()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(customMiddleware.RecoverJSON)

	// Custom Method Not Allowed handler to set "Allow" header (required by RFC 9110 and verified by Schemathesis)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allowedMethods(r.URL.Path))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"code":"method_not_allowed","message":"method not allowed"}`))
	})

	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(store)
	itemHandler := handler.NewItemHandler(store)
	orderHandler := handler.NewOrderHandler(store)
	bugHandler := handler.NewBugHandler()

	r.Get("/health", healthHandler.GetHealth)

	// Users
	r.Get("/users", userHandler.ListUsers)
	r.Post("/users", userHandler.CreateUser)
	r.Get("/users/{userId}", userHandler.GetUser)
	r.Put("/users/{userId}", userHandler.UpdateUser)
	r.Delete("/users/{userId}", userHandler.DeleteUser)

	// Items
	r.Get("/items", itemHandler.ListItems)
	r.Post("/items", itemHandler.CreateItem)
	r.Get("/items/{itemId}", itemHandler.GetItem)

	// Orders
	r.Post("/orders", orderHandler.CreateOrder)
	r.Get("/orders/{orderId}", orderHandler.GetOrder)

	// Auth
	r.With(customMiddleware.BearerAuth("test-token")).Get("/me", userHandler.GetMe)

	// Bugs
	r.Get("/bugs/schema-mismatch", bugHandler.SchemaMismatch)
	r.Get("/bugs/status-mismatch", bugHandler.StatusMismatch)
	r.Post("/bugs/panic-on-zero", bugHandler.PanicOnZero)
	r.Get("/bugs/invalid-email", bugHandler.InvalidEmail)

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
