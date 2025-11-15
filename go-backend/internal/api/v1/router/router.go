package router

import (
	"log/slog"
	"net/http"

	"go-backend/internal/api/v1/handler"
	"go-backend/internal/config"
	"go-backend/internal/userservice"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// NewV1Router returns the main http.Handler configured with chi routes.
func NewV1Router(db *gorm.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	serviceFactory := userservice.NewServiceFactory()
	userService, err := serviceFactory.NewUserService(userservice.ServiceType(cfg.UserServiceType))
	if err != nil {
		slog.Error("Failed to create user service", "error", err)
		panic(err)
	}

	r.Mount("/micro-apps", microAppRoutes(db))
	r.Mount("/users", userRoutes(db, userService))
	r.Mount("/user-info", userInfoRoutes(userService))

	return r
}

// microAppRoutes sets up a sub-router for all endpoints prefixed with /micro-apps.
func microAppRoutes(db *gorm.DB) http.Handler {
	r := chi.NewRouter()

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

// userInfoRoutes sets up a sub-router for /user-info endpoint.
func userInfoRoutes(userService userservice.UserService) http.Handler {
	r := chi.NewRouter()

	userHandler := handler.NewUserHandler(userService)

	// GET /user-info - Get current logged-in user's info
	r.Get("/", userHandler.GetUserInfo)

	return r
}

// userRoutes sets up a sub-router for all endpoints prefixed with /users.
func userRoutes(db *gorm.DB, userService userservice.UserService) http.Handler {
	r := chi.NewRouter()

	userHandler := handler.NewUserHandler(userService)
	userConfigHandler := handler.NewUserConfigHandler(db)

	// GET /users
	r.Get("/", userHandler.GetAll)

	// POST /users
	r.Post("/", userHandler.Upsert)

	// DELETE /users/{email}
	r.Delete("/{email}", userHandler.Delete)

	// GET /users/app-configs
	r.Get("/app-configs", userConfigHandler.GetAppConfigs)

	// POST /users/app-configs
	r.Post("/app-configs", userConfigHandler.UpsertAppConfig)

	return r
}
