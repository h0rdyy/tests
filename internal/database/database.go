package database

import (
	"database/sql"
	"os"
	"path/filepath"
)

const createLeadsTableSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS leads (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    phone      TEXT NOT NULL,
    created_at TEXT NOT NULL,
    source     TEXT NOT NULL
);`

func Initialize(db *sql.DB, root, configuredPath string) error {
	dbPath := configuredPath
	if dbPath == "" {
		dbPath = filepath.Join(root, "data", "cybersec.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}
	_, err := db.Exec(createLeadsTableSQL)
	return err
}
