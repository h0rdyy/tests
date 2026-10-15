package main

import (
	"database/sql"
	"net/http"
	"strings"
)

// adminContentHandler обслуживает формы сотрудников и услуг.
func adminContentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, logger interface{ Printf(string, ...interface{}) }) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	section, id := parts[1], ""
	if len(parts) > 2 && parts[2] != "new" {
		id = parts[2]
	}
	if len(parts) == 2 && section == "staff" && r.Method == http.MethodPost {
		saveStaff(w, r, db, id, logger)
		return
	}
	if len(parts) == 2 && section == "services" && r.Method == http.MethodPost {
		saveAdminService(w, r, db, id, logger)
		return
	}
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	if section == "staff" {
		if r.Method == http.MethodGet {
			renderStaffForm(w, db, id, logger)
			return
		}
		if r.Method == http.MethodPost && len(parts) == 3 {
			saveStaff(w, r, db, id, logger)
			return
		}
	}
	if section == "services" {
		if r.Method == http.MethodGet {
			renderServiceForm(w, db, id, logger)
			return
		}
		if r.Method == http.MethodPost && len(parts) == 3 {
			saveAdminService(w, r, db, id, logger)
			return
		}
	}
	http.NotFound(w, r)
}

// renderStaffForm отображает форму сотрудника.
func renderStaffForm(w http.ResponseWriter, db *sql.DB, id string, logger interface{ Printf(string, ...interface{}) }) {
	data := PageData{}
	if id != "" {
		item, err := staffByID(db, id)
		if err != nil {
			http.NotFound(w, nil)
			return
		}
		data.Staff = item
	}
	renderNamedTemplate(w, "admin/staff_form.html", data, logger)
}

// renderServiceForm отображает форму услуги.
func renderServiceForm(w http.ResponseWriter, db *sql.DB, id string, logger interface{ Printf(string, ...interface{}) }) {
	data := PageData{Services: services}
	data.AdminServices = []AdminService{}
	data.AdminService = AdminService{Direction: directions[0]}
	if id != "" {
		item, err := serviceByID(db, id)
		if err != nil {
			http.NotFound(w, nil)
			return
		}
		data.AdminService = item
	}
	renderNamedTemplate(w, "admin/service_form.html", data, logger)
}

// saveStaff создаёт или обновляет сотрудника.
func saveStaff(w http.ResponseWriter, r *http.Request, db *sql.DB, id string, logger interface{ Printf(string, ...interface{}) }) {
	_ = r.ParseMultipartForm(8 << 20)
	name := strings.TrimSpace(r.FormValue("full_name"))
	if name == "" {
		http.Error(w, "full_name is required", http.StatusBadRequest)
		return
	}
	status := normalizeStatus(r.FormValue("status"))
	photo := strings.TrimSpace(r.FormValue("photo_url"))
	var err error
	if id == "" {
		_, err = db.Exec(`INSERT INTO staff(full_name,position,short_info,practice,achievements,education,photo_url,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, name, r.FormValue("position"), r.FormValue("short_info"), r.FormValue("practice"), r.FormValue("achievements"), r.FormValue("education"), photo, status)
	} else {
		_, err = db.Exec(`UPDATE staff SET full_name=?,position=?,short_info=?,practice=?,achievements=?,education=?,photo_url=CASE WHEN ?='' THEN photo_url ELSE ? END,status=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, name, r.FormValue("position"), r.FormValue("short_info"), r.FormValue("practice"), r.FormValue("achievements"), r.FormValue("education"), photo, photo, status, id)
	}
	if err != nil {
		logger.Printf("save staff: %v", err)
		http.Error(w, "save failed", 500)
		return
	}
	redirectTo := "/admin/staff/new"
	if id != "" {
		redirectTo = "/admin/staff/" + id
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// saveAdminService создаёт или обновляет услугу.
func saveAdminService(w http.ResponseWriter, r *http.Request, db *sql.DB, id string, logger interface{ Printf(string, ...interface{}) }) {
	direction, title := strings.TrimSpace(r.FormValue("direction")), strings.TrimSpace(r.FormValue("title"))
	if direction == "" || title == "" {
		http.Error(w, "direction and title are required", 400)
		return
	}
	status := normalizeStatus(r.FormValue("status"))
	var err error
	if id == "" {
		_, err = db.Exec(`INSERT INTO services(direction,title,description,status) VALUES(?,?,?,?)`, direction, title, r.FormValue("description"), status)
	} else {
		_, err = db.Exec(`UPDATE services SET direction=?,title=?,description=?,status=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, direction, title, r.FormValue("description"), status, id)
	}
	if err != nil {
		logger.Printf("save service: %v", err)
		http.Error(w, "save failed", 500)
		return
	}
	redirectTo := "/admin/services/new"
	if id != "" {
		redirectTo = "/admin/services/" + id
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// expertsList отображает опубликованных экспертов.
func expertsList(w http.ResponseWriter, db *sql.DB, logger interface{ Printf(string, ...interface{}) }) {
	renderNamedTemplate(w, "experts/list.html", PageData{Staffs: listPublishedStaff(db)}, logger)
}

// expertView отображает профиль опубликованного эксперта.
func expertView(w http.ResponseWriter, db *sql.DB, id string, logger interface{ Printf(string, ...interface{}) }) {
	item, err := staffByID(db, id)
	if err != nil || item.Status != "published" {
		http.NotFound(w, nil)
		return
	}
	renderNamedTemplate(w, "experts/view.html", PageData{Staff: item}, logger)
}
