package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	JWTSecret       string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	UserServiceAddr string
	SwaggerEnabled  bool
	Google          GoogleConfig
	DB              DBConfig
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type DBConfig struct{ Host, Port, User, Password, Name, SSLMode string }

func (db DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode)
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		Port:            env("PORT", "8080"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTTL:       envDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:      envDuration("REFRESH_TTL", 720*time.Hour),
		UserServiceAddr: env("USER_SERVICE_ADDR", "localhost:50052"),
		SwaggerEnabled:  env("ENABLE_SWAGGER", "false") == "true",
		Google: GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  env("GOOGLE_REDIRECT_URL", "http://localhost:8084/api/v1/auth/google/callback"),
		},
		DB: DBConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5435"),
			User:     env("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     env("DB_NAME", "auth"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET required")
	}
	return cfg, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
