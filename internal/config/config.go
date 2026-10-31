package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort        = "8080"
	defaultJWTTTL      = 60
	minJWTSecretLength = 32
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration

	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (*Config, error) {
	// Load .env if present. Existing environment variables take precedence.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	port := getEnv("PORT", defaultPort)

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("invalid PORT: expected a value between 1 and 65535")
	}

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	jwtSecret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	if len(jwtSecret) < minJWTSecretLength {
		return nil, fmt.Errorf(
			"JWT_SECRET must be at least %d bytes",
			minJWTSecretLength,
		)
	}

	ttlMinutes := getEnv(
		"JWT_TTL_MINUTES",
		strconv.Itoa(defaultJWTTTL),
	)

	minutes, err := strconv.Atoi(ttlMinutes)
	if err != nil || minutes <= 0 {
		return nil, fmt.Errorf(
			"invalid JWT_TTL_MINUTES: expected a positive integer, got %q",
			ttlMinutes,
		)
	}

	if minutes > int((1<<63-1)/int64(time.Minute)) {
		return nil, fmt.Errorf("JWT_TTL_MINUTES is too large")
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		JWTTTL:      time.Duration(minutes) * time.Minute,

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf(
			"missing required environment variable: %s",
			key,
		)
	}

	return value, nil
}
