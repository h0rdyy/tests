package main

import (
	"database/sql"
	"strings"
)

// Staff хранит профиль сотрудника и его профессиональные разделы.
type Staff struct {
	ID           int
	FullName     string
	Position     string
	ShortInfo    string
	Practice     string
	Achievements string
	Education    string
	PhotoURL     string
	Status       string
}

// AdminService хранит редактируемую услугу админки.
type AdminService struct {
	ID          int
	Direction   string
	Title       string
	Description string
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
			achievements TEXT,
			education TEXT,
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
			status TEXT NOT NULL DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`)
	return err
}

// listPublishedStaff возвращает опубликованных сотрудников.
func listPublishedStaff(db *sql.DB) []Staff {
	rows, err := db.Query(`SELECT id,full_name,position,short_info,practice,achievements,education,photo_url,status FROM staff WHERE status='published' ORDER BY id DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Staff
	for rows.Next() {
		var item Staff
		if rows.Scan(&item.ID, &item.FullName, &item.Position, &item.ShortInfo, &item.Practice, &item.Achievements, &item.Education, &item.PhotoURL, &item.Status) == nil {
			result = append(result, item)
		}
	}
	return result
}

// staffByID возвращает сотрудника по идентификатору.
func staffByID(db *sql.DB, id string) (Staff, error) {
	var item Staff
	err := db.QueryRow(`SELECT id,full_name,position,short_info,practice,achievements,education,photo_url,status FROM staff WHERE id=?`, id).
		Scan(&item.ID, &item.FullName, &item.Position, &item.ShortInfo, &item.Practice, &item.Achievements, &item.Education, &item.PhotoURL, &item.Status)
	return item, err
}

// listPublishedServices возвращает опубликованные услуги из БД для публичного сайта.
func listPublishedServices(db *sql.DB) []AdminService {
	rows, err := db.Query(`SELECT id,direction,title,description,status FROM services WHERE status='published' ORDER BY direction,id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []AdminService
	for rows.Next() {
		var item AdminService
		if rows.Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Status) == nil {
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
	rows, err := db.Query(`SELECT id,direction,title,description,status FROM services ORDER BY direction,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AdminService
	for rows.Next() {
		var item AdminService
		if rows.Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Status) == nil {
			result = append(result, item)
		}
	}
	return result, rows.Err()
}

// serviceByID возвращает услугу админки по идентификатору.
func serviceByID(db *sql.DB, id string) (AdminService, error) {
	var item AdminService
	err := db.QueryRow(`SELECT id,direction,title,description,status FROM services WHERE id=?`, id).
		Scan(&item.ID, &item.Direction, &item.Title, &item.Description, &item.Status)
	return item, err
}

// normalizeStatus ограничивает состояние записи двумя допустимыми значениями.
func normalizeStatus(value string) string {
	if strings.EqualFold(value, "published") {
		return "published"
	}
	return "draft"
}
