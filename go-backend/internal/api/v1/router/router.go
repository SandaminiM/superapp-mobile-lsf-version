package router

import (
	"net/http"

	"go-backend/internal/api/v1/handler"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// NewV1Router returns the main http.Handler configured with chi routes.
func NewV1Router(db *gorm.DB) http.Handler {
	r := chi.NewRouter()

	r.Mount("/micro-apps", microAppRoutes(db))
	r.Mount("/users", userRoutes(db))

	return r
}

// microAppRoutes sets up a sub-router for all endpoints prefixed with /micro-apps.
func microAppRoutes(db *gorm.DB) http.Handler {
	r := chi.NewRouter()

	// Initialize Microapp Handlers
	microappHandler := handler.NewMicroAppHandler(db)
	microappVersionHandler := handler.NewMicroAppVersionHandler(db)

	// GET /micro-apps
	r.Get("/", microappHandler.GetAll)

	// GET /micro-apps/{appID}
	r.Get("/{appID}", microappHandler.GetByID)

	// POST /micro-apps
	r.Post("/", microappHandler.Upsert)

	// PUT /micro-apps/deactivate/{appID}
	r.Put("/deactivate/{appID}", microappHandler.Deactivate)

	// POST /micro-apps/{appID}/versions
	r.Post("/{appID}/versions", microappVersionHandler.UpsertVersion)

	return r
}

// userRoutes sets up a sub-router for all endpoints prefixed with /users.
func userRoutes(db *gorm.DB) http.Handler {
	r := chi.NewRouter()

	// Initialize User Config Handler
	userConfigHandler := handler.NewUserConfigHandler(db)

	// GET /users/app-configs
	r.Get("/app-configs", userConfigHandler.GetAppConfigs)

	// POST /users/app-configs
	r.Post("/app-configs", userConfigHandler.UpsertAppConfig)

	return r
}
