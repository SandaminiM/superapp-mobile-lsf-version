package tokenexchange

import (
	"fmt"
	"log/slog"
	"time"

	"go-backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// IssueJWT generates a JWT token for the given user and microapp
func IssueJWT(cfg *config.Config, userEmail, microAppID string) (string, error) {
	slog.Info("Issuing JWT token", "email", userEmail, "microAppId", microAppID)

	key, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}

	hashedAudience := GetHashedMicroAppID(microAppID)

	ttl := getTokenTTL(cfg.TokenTTLSeconds)
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := jwt.MapClaims{
		"iss":   cfg.SuperAppIssuer,  // Issuer
		"aud":   hashedAudience,      // Audience (hashed microAppId)
		"exp":   expiresAt.Unix(),    // Expiration time
		"iat":   now.Unix(),          // Issued at
		"jti":   uuid.New().String(), // JWT ID (unique identifier)
		"email": userEmail,           // Custom claim: user email
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	token.Header["kid"] = cfg.SuperAppKeyID

	tokenString, err := token.SignedString(key)
	if err != nil {
		slog.Error("Failed to sign JWT token", "error", err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	slog.Info("JWT token issued successfully", "email", userEmail, "expiresAt", expiresAt)
	return tokenString, nil
}
