package userservice

import (
	"log/slog"

	"go-backend/internal/config"
	"go-backend/internal/database"
	"go-backend/internal/models"

	"gorm.io/gorm"
)

type DBUserService struct {
	db *gorm.DB
}

// Verify that DBUserService implements UserService interface
var _ UserService = (*DBUserService)(nil)

func NewDBUserService() UserService {
	cfg := config.Load()
	db := database.Connect(cfg)
	return &DBUserService{db: db}
}

// GetUserByEmail retrieves a user by their email address.
func (s *DBUserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		slog.Error("Failed to fetch user by email", "error", result.Error, "email", email)
		return nil, result.Error
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func (s *DBUserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := s.db.Order("firstName, lastName").Find(&users)

	if result.Error != nil {
		slog.Error("Failed to fetch all users", "error", result.Error)
		return nil, result.Error
	}

	return users, nil
}

// UpsertUser creates a new user or updates an existing one.
func (s *DBUserService) UpsertUser(user *models.User) error {
	result := s.db.Where("email = ?", user.Email).
		Assign(models.User{
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			UserThumbnail: user.UserThumbnail,
			Location:      user.Location,
		}).
		Attrs(models.User{
			Email: user.Email,
		}).
		FirstOrCreate(user)

	if result.Error != nil {
		slog.Error("Failed to upsert user", "error", result.Error, "email", user.Email)
		return result.Error
	}

	return nil
}

// UpsertUsers creates or updates multiple users in the database within a transaction.
func (s *DBUserService) UpsertUsers(users []models.User) error {
	slog.Info("Upserting bulk users", "count", len(users))

	return s.db.Transaction(func(tx *gorm.DB) error {
		txService := &DBUserService{db: tx}

		for i, user := range users {
			if err := txService.UpsertUser(&user); err != nil {
				slog.Error("Failed to upsert user in bulk operation",
					"error", err, "index", i, "email", user.Email)
				return err
			}
		}

		slog.Info("Successfully upserted bulk users", "count", len(users))
		return nil
	})
}

// DeleteUser removes a user by their email address.
func (s *DBUserService) DeleteUser(email string) error {
	result := s.db.Where("email = ?", email).Delete(&models.User{})

	if result.Error != nil {
		slog.Error("Failed to delete user", "error", result.Error, "email", email)
		return result.Error
	}

	if result.RowsAffected == 0 {
		slog.Warn("No user found to delete", "email", email)
		return gorm.ErrRecordNotFound
	}

	return nil
}
