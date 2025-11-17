package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"go-backend/internal/api/v1/dto"
	"go-backend/internal/auth"
	"go-backend/internal/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type MicroAppVersionHandler struct {
	db *gorm.DB
}

func NewMicroAppVersionHandler(db *gorm.DB) *MicroAppVersionHandler {
	return &MicroAppVersionHandler{db: db}
}

// getMicroAppIDsByGroups fetches micro app IDs accessible by the given user groups
func (h *MicroAppVersionHandler) getMicroAppIDsByGroups(groups []string) ([]string, error) {
	if len(groups) == 0 {
		slog.Warn("No groups found for the user")
		return []string{}, nil
	}

	var appIDs []string
	if err := h.db.Model(&models.MicroAppRole{}).
		Select("DISTINCT micro_app_id").
		Where("active = ? AND role IN ?", 1, groups).
		Pluck("micro_app_id", &appIDs).Error; err != nil {
		return nil, err
	}

	if len(appIDs) == 0 {
		slog.Warn("No micro apps found for the given groups", "groups", groups)
		return []string{}, nil
	}

	return appIDs, nil
}

// UpsertVersion handles creating or updating a version for a micro app
func (h *MicroAppVersionHandler) UpsertVersion(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by auth middleware)
	userInfo, ok := auth.GetUserInfo(r.Context())
	if !ok {
		http.Error(w, "user info not found in context", http.StatusUnauthorized)
		return
	}
	userEmail := userInfo.Email

	appID := chi.URLParam(r, "appID")
	if appID == "" {
		http.Error(w, "missing micro_app_id", http.StatusBadRequest)
		return
	}

	// Check authorization: user must have access to this app
	authorizedAppIDs, err := h.getMicroAppIDsByGroups(userInfo.Groups)
	if err != nil {
		slog.Error("Failed to get authorized app IDs", "error", err, "groups", userInfo.Groups)
		http.Error(w, "failed to verify authorization", http.StatusInternalServerError)
		return
	}

	if !slices.Contains(authorizedAppIDs, appID) {
		slog.Warn("User not authorized to create version for micro app", "appID", appID, "email", userInfo.Email, "groups", userInfo.Groups)
		http.Error(w, "micro app not found", http.StatusNotFound)
		return
	}

	if !validateContentType(w, r) {
		return
	}

	limitRequestBody(w, r, 0) // 1MB default limit
	var req dto.CreateMicroAppVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if isMaxBytesError(err) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid request body", http.StatusBadRequest)
		}
		return
	}

	// Validate request
	if !validateStruct(w, &req) {
		return
	}

	// Use transaction to prevent race condition: check app exists and create version atomically
	var version models.MicroAppVersion
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Check if micro app exists within the transaction
		// This prevents the race condition where app could be deleted between check and create
		var microApp models.MicroApp
		if err := tx.Where("micro_app_id = ?", appID).First(&microApp).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("micro app not found: %w", err)
			}
			return fmt.Errorf("failed to fetch micro app: %w", err)
		}

		// Create or update version within the same transaction
		// If app gets deleted during this transaction, foreign key constraint will prevent orphaned record
		result := tx.Where("micro_app_id = ? AND version = ? AND build = ?", appID, req.Version, req.Build).
			Assign(models.MicroAppVersion{
				ReleaseNotes: req.ReleaseNotes,
				IconURL:      req.IconURL,
				DownloadURL:  req.DownloadURL,
				Active:       1,
				UpdatedBy:    &userEmail,
			}).
			Attrs(models.MicroAppVersion{
				MicroAppID: appID,
				Version:    req.Version,
				Build:      req.Build,
				CreatedBy:  userEmail,
			}).FirstOrCreate(&version)

		if result.Error != nil {
			return fmt.Errorf("failed to upsert version: %w", result.Error)
		}

		return nil
	})

	// Handle transaction errors
	if err != nil {
		// Check for foreign key constraint violation (app was deleted during transaction)
		if isForeignKeyError(err) {
			http.Error(w, "micro app not found or was deleted", http.StatusNotFound)
			return
		}

		// Check for "not found" error from app check
		if strings.Contains(err.Error(), "micro app not found") {
			http.Error(w, "micro app not found", http.StatusNotFound)
			return
		}

		// Other database errors
		slog.Error("Failed to upsert version", "error", err, "appID", appID, "version", req.Version, "build", req.Build)
		http.Error(w, "failed to upsert version", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusCreated, dto.MicroAppVersionResponse{
		ID:           version.ID,
		MicroAppID:   version.MicroAppID,
		Version:      version.Version,
		Build:        version.Build,
		ReleaseNotes: version.ReleaseNotes,
		IconURL:      version.IconURL,
		DownloadURL:  version.DownloadURL,
		Active:       version.Active,
	}); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
