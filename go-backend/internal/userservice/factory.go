package userservice

import (
	"fmt"
	"log/slog"
	"slices"
)

type ServiceType string

const (
	ServiceTypeDB ServiceType = "db"
	ServiceTypeHR ServiceType = "hr"
)

var ValidServiceTypes = []ServiceType{
	ServiceTypeDB,
	ServiceTypeHR,
}

type ServiceFactory struct{}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// NewUserService creates a UserService implementation based on the service type.
func (f *ServiceFactory) NewUserService(serviceType ServiceType) (UserService, error) {
	slog.Info("Creating user service", "type", serviceType)

	switch serviceType {
	case ServiceTypeDB:
		return NewDBUserService(), nil

	case ServiceTypeHR:
		slog.Warn("hr service type requested but not yet implemented")
		return nil, fmt.Errorf("hr user service not implemented yet")

	default:
		return nil, fmt.Errorf("unknown service type: %s (available: db, api, mock)", serviceType)
	}
}

// ValidateServiceType checks if the provided service type is valid.
func ValidateServiceType(serviceType string) error {
	st := ServiceType(serviceType)
	if slices.Contains(ValidServiceTypes, st) {
		return nil
	}
	return fmt.Errorf("invalid service type: %s (available: db, api, mock)", serviceType)
}
