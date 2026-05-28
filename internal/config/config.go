package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port        string
	MySQLDSN    string
	IngestToken string
	MediaDir    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        defaultString(strings.TrimSpace(os.Getenv("PORT")), "8080"),
		MySQLDSN:    strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		IngestToken: strings.TrimSpace(os.Getenv("INGEST_TOKEN")),
		MediaDir:    defaultString(strings.TrimSpace(os.Getenv("MEDIA_DIR")), "data/media"),
	}
	if cfg.MySQLDSN == "" {
		return Config{}, errors.New("MYSQL_DSN is required")
	}
	if cfg.IngestToken == "" {
		return Config{}, errors.New("INGEST_TOKEN is required")
	}
	return cfg, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
