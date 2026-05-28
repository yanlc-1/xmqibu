package config

import "testing"

func TestLoadIncludesMediaDir(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/db")
	t.Setenv("INGEST_TOKEN", "secret")
	t.Setenv("MEDIA_DIR", "/tmp/frontier-media")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MediaDir != "/tmp/frontier-media" {
		t.Fatalf("MediaDir = %q, want %q", cfg.MediaDir, "/tmp/frontier-media")
	}
}

func TestLoadDefaultsMediaDir(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/db")
	t.Setenv("INGEST_TOKEN", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MediaDir != "data/media" {
		t.Fatalf("MediaDir = %q, want %q", cfg.MediaDir, "data/media")
	}
}
