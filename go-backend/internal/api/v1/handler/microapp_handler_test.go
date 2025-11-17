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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.MicroApp{},
		&models.MicroAppVersion{},
		&models.MicroAppRole{},
		&models.MicroAppConfig{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// setupTestData creates test data in the database
func setupTestData(t *testing.T, db *gorm.DB) {
	userEmail := "test@example.com"
	appID := "test-app"

	// Create test app
	app := models.MicroApp{
		MicroAppID: appID,
		Name:       "Test App",
		Active:     1,
		CreatedBy:  userEmail,
	}
	if err := db.Create(&app).Error; err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	// Create test role
	role := models.MicroAppRole{
		MicroAppID: appID,
		Role:       "admin",
		Active:     1,
		CreatedBy:  userEmail,
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}

	// Create test version
	version := models.MicroAppVersion{
		MicroAppID:  appID,
		Version:     "1.0.0",
		Build:       1,
		DownloadURL: "https://example.com/app.zip",
		Active:      1,
		CreatedBy:   userEmail,
	}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("failed to create test version: %v", err)
	}
}

func TestMicroAppHandler_GetAll(t *testing.T) {
	db := setupTestDB(t)
	setupTestData(t, db)
	handler := NewMicroAppHandler(db)

	tests := []struct {
		name           string
		userInfo       *TestUserInfo
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "user with access",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "user without access",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"user"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "no user info",
			userInfo:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithAuth(http.MethodGet, "/api/v1/micro-apps", nil, tt.userInfo)
			w := httptest.NewRecorder()

			handler.GetAll(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []dto.MicroAppResponse
				if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if len(response) != tt.expectedCount {
					t.Errorf("expected %d apps, got %d", tt.expectedCount, len(response))
				}
			}
		})
	}
}

func TestMicroAppHandler_GetByID(t *testing.T) {
	db := setupTestDB(t)
	setupTestData(t, db)
	handler := NewMicroAppHandler(db)

	tests := []struct {
		name           string
		appID          string
		userInfo       *TestUserInfo
		expectedStatus int
	}{
		{
			name:  "user with access",
			appID: "test-app",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "user without access",
			appID: "test-app",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"user"},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "app not found",
			appID:          "non-existent",
			userInfo:       &TestUserInfo{Email: "test@example.com", Groups: []string{"admin"}},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing appID",
			appID:          "",
			userInfo:       &TestUserInfo{Email: "test@example.com", Groups: []string{"admin"}},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithAuth(http.MethodGet, "/api/v1/micro-apps/"+tt.appID, nil, tt.userInfo)
			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			if tt.appID != "" {
				rctx.URLParams.Add("appID", tt.appID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.GetByID(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestMicroAppHandler_Upsert(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMicroAppHandler(db)

	tests := []struct {
		name           string
		body           dto.CreateMicroAppRequest
		userInfo       *TestUserInfo
		expectedStatus int
	}{
		{
			name: "create new app",
			body: dto.CreateMicroAppRequest{
				AppID: "new-app",
				Name:  "New App",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			body: dto.CreateMicroAppRequest{
				AppID: "", // Missing
				Name:  "Test",
			},
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "no user info",
			body: dto.CreateMicroAppRequest{
				AppID: "test-app",
				Name:  "Test App",
			},
			userInfo:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := createRequestWithAuth(http.MethodPost, "/api/v1/micro-apps", bodyBytes, tt.userInfo)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.Upsert(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestMicroAppHandler_Deactivate(t *testing.T) {
	db := setupTestDB(t)
	setupTestData(t, db)
	handler := NewMicroAppHandler(db)

	tests := []struct {
		name           string
		appID          string
		userInfo       *TestUserInfo
		expectedStatus int
	}{
		{
			name:  "deactivate existing app",
			appID: "test-app",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "app not found",
			appID: "non-existent",
			userInfo: &TestUserInfo{
				Email:  "test@example.com",
				Groups: []string{"admin"},
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithAuth(http.MethodPut, "/api/v1/micro-apps/deactivate/"+tt.appID, nil, tt.userInfo)
			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("appID", tt.appID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.Deactivate(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			// Verify app is deactivated
			if tt.expectedStatus == http.StatusOK {
				var app models.MicroApp
				if err := db.Where("micro_app_id = ?", tt.appID).First(&app).Error; err == nil {
					if app.Active != 0 {
						t.Error("app should be deactivated")
					}
				}
			}
		})
	}
}
