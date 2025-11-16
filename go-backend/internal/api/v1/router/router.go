package router

import (
	"net/http"

	"go-backend/internal/api/v1/handler"
	"go-backend/internal/config"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// NewV1Router returns the main http.Handler configured with chi routes.
func NewV1Router(db *gorm.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Mount("/micro-apps", microAppRoutes(db))
	r.Mount("/tokens", tokenRoutes(cfg))

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

// tokenRoutes sets up a sub-router for token operations.
func tokenRoutes(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	tokenHandler := handler.NewTokenHandler(cfg)

	// POST /tokens
	r.Post("/", tokenHandler.IssueToken)

	return r
}

// WellKnownRoutes sets up a sub-router for well-known endpoints.
func WellKnownRoutes(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	tokenHandler := handler.NewTokenHandler(cfg)

	// GET /.well-known/jwks
	r.Get("/jwks", tokenHandler.GetJWKS)

	return r
}
