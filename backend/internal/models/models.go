package models

import (
	"time"

	"github.com/google/uuid"
)

// Role constants for users.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents an account. PasswordHash is never serialised to JSON.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Task represents a single task owned by a user.
type Task struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Activity is one entry in a task's change history.
type Activity struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"taskId"`
	UserID    uuid.UUID `json:"userId"`
	Action    string    `json:"action"`
	Detail    any       `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}
