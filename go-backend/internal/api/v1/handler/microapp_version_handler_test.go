package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-backend/internal/api/v1/dto"
	"go-backend/internal/models"

	"github.com/go-chi/chi/v5"
)

func TestMicroAppVersionHandler_UpsertVersion(t *testing.T) {
	db := setupTestDB(t)
	setupTestData(t, db)
	handler := NewMicroAppVersionHandler(db)

	tests := []struct {
		name           string
		appID          string
		body           dto.CreateMicroAppVersionRequest
		userInfo       *TestUserInfo
		expectedStatus int
	}{
		{
			name:  "create new version",
			appID: "test-app",
			body: dto.CreateMicroAppVersionRequest{
				Version:     "2.0.0",
				Build:       2,
				DownloadURL: "https://example.com/app-v2.zip",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "update existing version",
			appID: "test-app",
			body: dto.CreateMicroAppVersionRequest{
				Version:     "1.0.0",
				Build:       1,
				DownloadURL: "https://example.com/app-updated.zip",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "user without access",
			appID: "test-app",
			body: dto.CreateMicroAppVersionRequest{
				Version:     "2.0.0",
				Build:       2,
				DownloadURL: "https://example.com/app.zip",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"user"},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "missing required fields",
			appID: "test-app",
			body: dto.CreateMicroAppVersionRequest{
				Version:     "", // Missing
				Build:       1,
				DownloadURL: "https://example.com/app.zip",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "app not found",
			appID: "non-existent",
			body: dto.CreateMicroAppVersionRequest{
				Version:     "1.0.0",
				Build:       1,
				DownloadURL: "https://example.com/app.zip",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "no user info",
			appID:          "test-app",
			body:           dto.CreateMicroAppVersionRequest{Version: "1.0.0", Build: 1, DownloadURL: "https://example.com/app.zip"},
			userInfo:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := createRequestWithAuth(http.MethodPost, "/api/v1/micro-apps/"+tt.appID+"/versions", bodyBytes, tt.userInfo)
			req.Header.Set("Content-Type", "application/json")

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("appID", tt.appID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.UpsertVersion(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			// Verify version was created/updated
			if tt.expectedStatus == http.StatusCreated {
				var version models.MicroAppVersion
				if err := db.Where("micro_app_id = ? AND version = ? AND build = ?", tt.appID, tt.body.Version, tt.body.Build).First(&version).Error; err != nil {
					t.Errorf("version should be created: %v", err)
				}
			}
		})
	}
}
