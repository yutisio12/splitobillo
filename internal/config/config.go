package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	Port            string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	OCRServiceURL   string
	JWTSecret       string
	SessionTTL      time.Duration
	UploadDir       string
	MaxUploadBytes  int64
	RateLimitPerMin int
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
}

func Load() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "splitbill"),
		DBPassword:      getEnv("DB_PASSWORD", "splitbill"),
		DBName:          getEnv("DB_NAME", "splitbill"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		OCRServiceURL:   getEnv("OCR_SERVICE_URL", "http://localhost:8090"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		SessionTTL:      time.Duration(getEnvInt("SESSION_TTL_MINUTES", 1440)) * time.Minute,
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadBytes:  int64(getEnvInt("MAX_UPLOAD_BYTES", 10*1024*1024)),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MINUTE", 300),
		AllowedOrigins:  splitCSV(getEnv("ALLOWED_ORIGINS", "*")),
		ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
	}
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
