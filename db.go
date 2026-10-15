package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// openDatabase открывает SQLite-базу в каталоге data и включает WAL-режим.
func openDatabase() (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Join("data", "uploads"), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", "file:data/site.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			publish_date TEXT NOT NULL,
			author TEXT NOT NULL,
			tags TEXT NOT NULL DEFAULT '[]',
			content_md TEXT NOT NULL,
			cover_path TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
