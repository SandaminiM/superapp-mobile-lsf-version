package userservice

import "go-backend/internal/models"

// UserService defines the interface for user-related operations.
type UserService interface {
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpsertUser(user *models.User) error
	UpsertUsers(users []models.User) error
	DeleteUser(email string) error
}
