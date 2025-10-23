package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string

	JWTSecret       string
	JWTAccessTTLMin int // in minutes (not used strictly but placeholder)

	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
}

func LoadConfigFromEnv() (*Config, error) {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		return nil, errors.New("DATABASE_URL required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET required")
	}
	cfg := &Config{
		DatabaseURL:     db,
		JWTSecret:       jwtSecret,
		JWTAccessTTLMin: 15,
		S3Endpoint:      os.Getenv("S3_ENDPOINT"),
		S3Region:        os.Getenv("S3_REGION"),
		S3AccessKey:     os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:     os.Getenv("S3_SECRET_KEY"),
		S3Bucket:        os.Getenv("S3_BUCKET"),
	}
	// S3_USE_SSL default true
	if os.Getenv("S3_USE_SSL") == "false" {
		cfg.S3UseSSL = false
	} else {
		cfg.S3UseSSL = true
	}
	return cfg, nil
}
