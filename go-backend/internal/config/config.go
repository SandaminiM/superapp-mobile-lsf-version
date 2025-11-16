package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	ServerPort string

	JWKSURL     string
	JWTIssuer   string
	JWTAudience string

	SuperAppIssuer  string
	SuperAppKeyID   string
	JWKSFilePath    string
	PrivateKeyPath  string
	TokenTTLSeconds string
}

func Load() *Config {
	cfg := &Config{
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "testdb"),
		ServerPort: getEnv("SERVER_PORT", "9090"),

		JWKSURL:     getEnv("JWKS_URL", "fallback <idp-metadata-url>/jwks"),
		JWTIssuer:   getEnv("JWT_ISSUER", "fallback <idp-issuer-url>"),
		JWTAudience: getEnv("JWT_AUDIENCE", "fallback <target-audience-in-token>"),

		SuperAppIssuer:  getEnv("SUPERAPP_ISSUER", "superapp-backend"),
		SuperAppKeyID:   getEnv("SUPERAPP_KEY_ID", "superapp-backend-publicKey_1"),
		JWKSFilePath:    getEnv("JWKS_FILE_PATH", "./internal/tokenexchange/jwks.json"),
		PrivateKeyPath:  getEnv("PRIVATE_KEY_PATH", "./internal/tokenexchange/private_key.pem"),
		TokenTTLSeconds: getEnv("TOKEN_TTL_SECONDS", "300"),
	}

	slog.Info("Configuration loaded", "server_port", cfg.ServerPort, "db_host", cfg.DBHost)
	slog.Info("Token exchange config", "private_key_path", cfg.PrivateKeyPath, "jwks_path", cfg.JWKSFilePath)
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
