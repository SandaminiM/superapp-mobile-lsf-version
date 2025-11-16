package tokenexchange

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"go-backend/internal/config"
)

var (
	privateKey     *rsa.PrivateKey
	privateKeyOnce sync.Once
	tokenTTL       time.Duration
	tokenTTLOnce   sync.Once
)

// GetHashedMicroAppID hashes the microAppId using SHA-256 and returns base64 encoded string
func GetHashedMicroAppID(microAppID string) string {
	hash := sha256.Sum256([]byte(microAppID))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GetJWKS reads and returns the JSON Web Key Set from the JWKS file
func GetJWKS(cfg *config.Config) (*JSONWebKeySet, error) {
	data, err := os.ReadFile(cfg.JWKSFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS file: %w", err)
	}

	var jwks JSONWebKeySet
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	return &jwks, nil
}

// loadPrivateKey loads the RSA private key from PEM file (singleton)
func loadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	var err error
	privateKeyOnce.Do(func() {
		slog.Info("Loading RSA private key for JWT signing", "path", privateKeyPath)

		keyData, readErr := os.ReadFile(privateKeyPath)
		if readErr != nil {
			err = fmt.Errorf("failed to read private key file: %w", readErr)
			return
		}

		block, _ := pem.Decode(keyData)
		if block == nil {
			err = fmt.Errorf("failed to decode PEM block from private key")
			return
		}

		key, parseErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse RSA private key: %w", parseErr)
			return
		}

		privateKey = key
		slog.Info("RSA private key loaded successfully")
	})

	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// getTokenTTL returns the token TTL duration (singleton)
func getTokenTTL(ttlSeconds string) time.Duration {
	tokenTTLOnce.Do(func() {
		// Default to 5 minutes (300 seconds)
		ttlSecondsInt := 300
		if ttlSeconds != "" {
			if val, err := strconv.Atoi(ttlSeconds); err == nil {
				ttlSecondsInt = val
			}
		}
		tokenTTL = time.Duration(ttlSecondsInt) * time.Second
		slog.Info("Token TTL configured", "seconds", ttlSecondsInt)
	})
	return tokenTTL
}
