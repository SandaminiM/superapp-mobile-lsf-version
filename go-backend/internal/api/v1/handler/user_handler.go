package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go-backend/internal/api/v1/dto"
	"go-backend/internal/auth"
	"go-backend/internal/models"
	"go-backend/internal/userservice"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService userservice.UserService
}

func NewUserHandler(userService userservice.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserInfo retrieves the currently logged-in user's information.
func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by auth middleware)
	userInfo, ok := auth.GetUserInfo(r.Context())
	if !ok {
		http.Error(w, "user info not found in context", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUserByEmail(userInfo.Email)
	if err != nil {
		slog.Error("Failed to fetch user info", "error", err, "email", userInfo.Email)
		http.Error(w, "failed to fetch user information", http.StatusInternalServerError)
		return
	}

	if user == nil {
		slog.Warn("User not found", "email", userInfo.Email)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	response := dto.UserResponse{
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		UserThumbnail: user.UserThumbnail,
		Location:      user.Location,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// GetAllUsers retrieves all users from the system.
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		slog.Error("Failed to fetch all users", "error", err)
		http.Error(w, "failed to fetch users", http.StatusInternalServerError)
		return
	}

	var response []dto.UserResponse
	for _, user := range users {
		response = append(response, dto.UserResponse{
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			UserThumbnail: user.UserThumbnail,
			Location:      user.Location,
		})
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// UpsertUser creates a new user(s) or updates an existing one.
func (h *UserHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r) {
		return
	}

	limitRequestBody(w, r, 1<<20) // 1MB default limit

	requests, isBulk, err := parseUpsertPayload(r.Body)
	if err != nil {
		slog.Error("Invalid request body for upsert", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	users, valid := convertAndValidateRequests(w, requests)
	if !valid {
		slog.Error("Validation failed for upsert")
		http.Error(w, "validation error", http.StatusBadRequest)
		return
	}

	// Bulk user upsert
	if isBulk {
		if err := h.userService.UpsertBulkUsers(users); err != nil {
			slog.Error("Failed to upsert bulk users", "error", err, "count", len(users))
			http.Error(w, "failed to upsert bulk users", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"message": "Users created/updated successfully"})
		return
	}

	// Single user upsert
	if err := h.userService.UpsertUser(users[0]); err != nil {
		slog.Error("Failed to upsert user", "error", err, "email", users[0].Email)
		http.Error(w, "failed to upsert user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "User created/updated successfully",
	})
}

// DeleteUser removes a user by their email address.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		http.Error(w, "missing email parameter", http.StatusBadRequest)
		return
	}

	err := h.userService.DeleteUser(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			slog.Error("Failed to delete user", "error", err, "email", email)
			http.Error(w, "failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"}); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// Helper functions

// Parses the request body for upsert user(s) operation.
func parseUpsertPayload(body any) ([]dto.UpsertUserRequest, bool, error) {
	var rawBody json.RawMessage
	if err := json.NewDecoder(body.(interface{ Read([]byte) (int, error) })).Decode(&rawBody); err != nil {
		return nil, false, err
	}

	// Check if it's an array
	if len(rawBody) > 0 && rawBody[0] == '[' {
		var requests []dto.UpsertUserRequest
		if err := json.Unmarshal(rawBody, &requests); err != nil {
			return nil, true, err
		}
		return requests, true, nil
	}

	var request dto.UpsertUserRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return nil, false, err
	}
	return []dto.UpsertUserRequest{request}, false, nil
}

// Converts DTOs to models and validates them.
func convertAndValidateRequests(w http.ResponseWriter, reqs []dto.UpsertUserRequest) ([]*models.User, bool) {
	users := make([]*models.User, 0, len(reqs))

	for _, r := range reqs {
		if !validateStruct(w, &r) {
			return nil, false
		}

		users = append(users, &models.User{
			Email:         r.Email,
			FirstName:     r.FirstName,
			LastName:      r.LastName,
			UserThumbnail: r.UserThumbnail,
			Location:      r.Location,
		})
	}

	return users, true
}
