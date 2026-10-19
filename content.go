package main

import (
	"database/sql"
	"strings"
)

// Staff хранит профиль сотрудника; Practice содержит резюме в Markdown.
type Staff struct {
	ID        int
	FullName  string
	Position  string
	ShortInfo string
	Practice  string
	PhotoURL  string
	Status    string
}

// AdminService хранит редактируемую услугу админки.
// Slug задаёт адрес страницы вида /services/<slug>; пустой — адрес по id.
type AdminService struct {
	ID          int
	Direction   string
	Title       string
	Description string
	Slug        string
	Status      string
}

var directions = []string{
	"Защита персональных данных",
	"Услуги по обеспечению информационной безопасности",
	"Информационное право",
	"Разработка образовательных программ",
	"Обучение и развитие персонала",
}

// initContentTables создаёт таблицы сотрудников и услуг без изменения articles.
func initContentTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS staff (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			position TEXT,
			short_info TEXT,
			practice TEXT,
			photo_url TEXT,
			status TEXT NOT NULL DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			direction TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			slug TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`)
	if err != nil {
		return err
	}
	// Миграции для таблиц, созданных в старых версиях схемы.
	for _, column := range []struct{ table, name, typ string }{
		{"services", "slug", "TEXT NOT NULL DEFAULT ''"},
	} {
		if !hasColumn(db, column.table, column.name) {
			if _, err := db.Exec(`ALTER TABLE ` + column.table + ` ADD COLUMN ` + column.name + ` ` + column.typ); err != nil {
				return err
			}
		}
	}
	return nil
}

// hasColumn проверяет наличие столбца в таблице SQLite.
func hasColumn(db *sql.DB, table, column string) bool {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil && strings.EqualFold(name, column) {
			return true
		}
	}
	return false
}

const staffColumns = `id,full_name,position,short_info,practice,photo_url,status`
const serviceColumns = `id,direction,title,description,slug,status`

// listPublishedStaff возвращает опубликованных сотрудников.
func listPublishedStaff(db *sql.DB) []Staff {
	rows, err := db.Query(`SELECT ` + staffColumns + ` FROM staff WHERE status='published' ORDER BY id DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Staff
	for rows.Next() {
		var item Staff
		if rows.Scan(&item.ID, &item.FullName, &item.Position, &item.ShortInfo, &item.Practice, &item.PhotoURL, &item.Status) == nil {
			result = append(result, item)
		}
	}
	return result
}

// staffByID возвращает сотрудника по идентификатору.
func staffByID(db *sql.DB, id string) (Staff, error) {
	var item Staff
	err := db.QueryRow(`SELECT `+staffColumns+` FROM staff WHERE id=?`, id).
		Scan(&item.ID, &item.FullName, &item.Position, &item.ShortInfo, &item.Practice, &item.PhotoURL, &item.Status)
	return item, err
}

// listPublishedServices возвращает опубликованные услуги из БД для публичного сайта.
func listPublishedServices(db *sql.DB) []AdminService {
	rows, err := db.Query(`SELECT ` + serviceColumns + ` FROM services WHERE status='published' ORDER BY direction,id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []AdminService
	for rows.Next() {
		var item AdminService
		if rows.Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Slug, &item.Status) == nil {
			result = append(result, item)
		}
	}
	return result
}

// publishedServicesByDirection группирует опубликованные услуги по направлениям.
func publishedServicesByDirection(db *sql.DB) map[string][]AdminService {
	grouped := make(map[string][]AdminService)
	for _, s := range listPublishedServices(db) {
		grouped[s.Direction] = append(grouped[s.Direction], s)
	}
	return grouped
}

// listAllServices возвращает все услуги админки (черновики и опубликованные).
func listAllServices(db *sql.DB) ([]AdminService, error) {
	rows, err := db.Query(`SELECT ` + serviceColumns + ` FROM services ORDER BY direction,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AdminService
	for rows.Next() {
		var item AdminService
		if rows.Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Slug, &item.Status) == nil {
			result = append(result, item)
		}
	}
	return result, rows.Err()
}

// serviceByID возвращает услугу по числовому id или строковому slug.
func serviceByID(db *sql.DB, key string) (AdminService, error) {
	var item AdminService
	err := db.QueryRow(`SELECT `+serviceColumns+` FROM services WHERE id=?`, key).
		Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Slug, &item.Status)
	if err == nil {
		return item, nil
	}
	if !isNumeric(key) {
		err = db.QueryRow(`SELECT `+serviceColumns+` FROM services WHERE slug<>'' AND slug=?`, key).
			Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Slug, &item.Status)
	}
	return item, err
}

// isNumeric сообщает, состоит ли строка только из цифр.
func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// normalizeStatus ограничивает состояние записи двумя допустимыми значениями.
func normalizeStatus(value string) string {
	if strings.EqualFold(value, "published") {
		return "published"
	}
	return "draft"
}
