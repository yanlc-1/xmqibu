package store

import (
	"context"
)

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS articles (
		id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		summary TEXT NOT NULL,
		cover_image VARCHAR(1024) NOT NULL DEFAULT '',
		body LONGTEXT NOT NULL,
		author VARCHAR(255) NOT NULL,
		source_links_json JSON NULL,
		published_at DATETIME NOT NULL,
		featured BOOLEAN NOT NULL DEFAULT FALSE,
		status VARCHAR(32) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS briefs (
		id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		summary TEXT NOT NULL,
		external_url VARCHAR(1024) NOT NULL DEFAULT '',
		importance INT NOT NULL DEFAULT 0,
		source_links_json JSON NULL,
		published_at DATETIME NOT NULL,
		status VARCHAR(32) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS tags (
		id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS content_tags (
		content_type VARCHAR(32) NOT NULL,
		content_id BIGINT NOT NULL,
		tag_id BIGINT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (content_type, content_id, tag_id),
		INDEX idx_content_tags_tag_id (tag_id),
		CONSTRAINT fk_content_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	)`,
	`CREATE TABLE IF NOT EXISTS ingest_records (
		id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		content_type VARCHAR(32) NOT NULL,
		external_id VARCHAR(255) NOT NULL,
		content_slug VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uniq_ingest_record (content_type, external_id)
	)`,
}

func (s *MySQLStore) EnsureSchema(ctx context.Context) error {
	for _, statement := range schemaStatements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
