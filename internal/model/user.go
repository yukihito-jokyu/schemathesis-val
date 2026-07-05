package model

import "time"

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
	RoleGuest  UserRole = "guest"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateUserRequest struct {
	Name  *string   `json:"name"`
	Email *string   `json:"email"`
	Age   *int      `json:"age"`
	Role  *UserRole `json:"role"`
}

type UpdateUserRequest struct {
	Name  *string   `json:"name"`
	Email *string   `json:"email"`
	Age   *int      `json:"age"`
	Role  *UserRole `json:"role"`
}
