package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Service содержит данные услуги, доступные HTML-шаблонам.
type Service struct {
	ID        int
	Title     string
	Hint      string
	Cell      string
	Thumb     string
	Direction string
	Slug      string
	ParentID  int
	Status    string
}

// URL возвращает адрес публичной страницы легаси-услуги: id или slug.
func (s Service) URL() string {
	if s.Slug != "" {
		return "/services/" + s.Slug
	}
	return "/services/" + strconv.Itoa(s.ID)
}

var services = []Service{
	{ID: 1, Title: "Защита персональных данных", Hint: "Ваш юридический щит и репутационный актив", Cell: "col1", Thumb: "shield.png"},
	{ID: 2, Title: "Услуги по обеспечению информационной безопасности", Hint: "Технологическая крепость вашего бизнеса", Cell: "top", Thumb: "fire.png"},
	{ID: 3, Title: "Информационное право", Cell: "bl", Thumb: "lightning.png"},
	{ID: 4, Title: "Разработка образовательных программ", Cell: "br", Thumb: "light.png"},
	{ID: 5, Title: "Обучение и развитие персонала", Hint: "Инвестиции в вашу безопасность и репутацию", Cell: "col4", Thumb: "study.png"},
}

// Article содержит опубликованный материал для карточки и страницы статьи.
type Article struct {
	Slug        string
	Title       string
	PublishDate string
	Author      string
	Tags        []string
	ContentMD   string
	ContentHTML string
	CoverURL    string
}

// PageData содержит данные, передаваемые страницам сайта.
type PageData struct {
	Services                 []Service
	AdminServicesByDirection map[string][]AdminService
	Service                  Service
	Articles                 []Article
	Article                  Article
	Staff                    Staff
	Staffs                   []Staff
	AdminService             AdminService
	AdminServices            []AdminService
	DBServices               []AdminService
	ServiceSections          []ServiceSection
	ServiceNav               []AdminService
	ServiceChildren          []AdminService
}

// serviceNav возвращает пункты выпадающего меню «Услуги»:
// опубликованные корневые услуги из БД, затем легаси-разделы, отсутствующие в БД.
func serviceNav(db *sql.DB) []Service {
	nav := make([]Service, 0, len(services))
	titles := map[string]bool{}
	for _, s := range listPublishedServices(db) {
		if s.ParentID == 0 && !titles[s.Title] {
			titles[s.Title] = true
			nav = append(nav, Service{ID: s.ID, Title: s.Title, Slug: s.Slug})
		}
	}
	for _, s := range services {
		if !titles[s.Title] {
			nav = append(nav, s)
		}
	}
	return nav
}

// reserveNumericSlugs приводит числовые slug-и записей БД к зарезервированным
// id легаси-услуг (1..5). Такие адреса всегда открывают страницу легаси-услуги,
// поэтому у записи из БД не может быть «чисто цифрового» адреса.
// Функция идемпотентна и безопасна для повторных запусков.
func reserveNumericSlugs(db *sql.DB, logger *log.Logger) {
	legacyIDs := map[int]bool{}
	for _, s := range services {
		legacyIDs[s.ID] = true
	}
	rows, err := db.Query(`SELECT id,title,slug FROM services`)
	if err != nil {
		logger.Printf("reserve slugs query: %v", err)
		return
	}
	type fix struct {
		id   int
		slug string
	}
	var fixes []fix
	for rows.Next() {
		var id int
		var title, slug string
		if err := rows.Scan(&id, &title, &slug); err != nil {
			logger.Printf("reserve slugs scan: %v", err)
			continue
		}
		if !isNumeric(slug) || !legacyIDs[id] {
			continue
		}
		n, _ := strconv.Atoi(slug)
		if legacyIDs[n] {
			continue // адрес ведёт на легаси-услугу с таким id — это нормально
		}
		newSlug := slug + "-" + slugify(title)
		if newSlug == slug {
			newSlug = slug + "-service"
		}
		if len(newSlug) > 96 {
			newSlug = strings.TrimRight(newSlug[:96], "-")
		}
		fixes = append(fixes, fix{id, newSlug})
	}
	rows.Close()
	for _, f := range fixes {
		if _, err := db.Exec(`UPDATE services SET slug=? WHERE id=?`, f.slug, f.id); err != nil {
			logger.Printf("reserve slug %d: %v", f.id, err)
		} else {
			logger.Printf("reserve slug: id=%d -> %s", f.id, f.slug)
		}
	}
}

// publicPageData возвращает направления услуг и опубликованные услуги из БД.
func publicPageData(db *sql.DB) PageData {
	return PageData{
		Services:                 serviceNav(db),
		AdminServicesByDirection: publishedServicesByDirection(db),
		DBServices:               listPublishedServices(db),
		ServiceSections:          publishedServiceSections(db),
	}
}

// renderServicePage отображает страницу услуги из БД: описание + аккордеоны подуслуг.
// onlyPublished — публичный режим (только опубликованные); в админке видны и черновики.
func renderServicePage(w http.ResponseWriter, r *http.Request, db *sql.DB, item AdminService, status int, onlyPublished bool, logger *log.Logger) {
	data := publicPageData(db)
	data.AdminService = item
	data.ServiceChildren = childServicesHTML(db, item.ID, onlyPublished)
	data.Article.ContentHTML = markdownHTML(item.Description)
	renderTemplateStatus(w, "service.html", data, status, logger)
}

// childServicesHTML возвращает подуслуги раздела с Markdown-описаниями,
// преобразованными в HTML для аккордеонов. includeDrafts — режим админки.
func childServicesHTML(db *sql.DB, parentID int, onlyPublished bool) []AdminService {
	var children []AdminService
	if onlyPublished {
		children = publishedChildServices(db, parentID)
	} else {
		all, _ := listAllServices(db)
		for _, s := range all {
			if s.ParentID == parentID {
				children = append(children, s)
			}
		}
	}
	for i := range children {
		children[i].Description = markdownHTML(children[i].Description)
	}
	return children
}

// newRouter собирает маршруты сайта и раздачу статических файлов.
func newRouter(db *sql.DB, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(filepath.Join("data", "uploads")))))
	mux.HandleFunc("/", pageHandler(db, logger))
	return mux
}

// pageHandler обрабатывает главную страницу и страницы сайта.
func pageHandler(db *sql.DB, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderTemplate(w, "index.html", publicPageData(db), logger)
		case r.URL.Path == "/about":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderTemplate(w, "about.html", publicPageData(db), logger)
		case r.URL.Path == "/privacy":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderTemplate(w, "privacy.html", publicPageData(db), logger)
		case r.URL.Path == "/expertise":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			data := publicPageData(db)
			data.Articles = publishedArticles(db)
			renderTemplate(w, "expertise.html", data, logger)
		case r.URL.Path == "/articles/kibernsk":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderTemplate(w, "articles/kibernsk.html", publicPageData(db), logger)
		case r.URL.Path == "/admin/articles/new":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderAdminForm(w, logger)
		case r.URL.Path == "/admin":
			http.Redirect(w, r, "/admin/articles/new", http.StatusSeeOther)
		case r.URL.Path == "/services":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderTemplate(w, "services.html", publicPageData(db), logger)
		case strings.HasPrefix(r.URL.Path, "/services/preview/"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			servicePreview(w, r, db, strings.TrimPrefix(r.URL.Path, "/services/preview/"), logger)
		case strings.HasPrefix(r.URL.Path, "/services/"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			servicePage(w, r, db, logger)
		case strings.HasPrefix(r.URL.Path, "/admin/staff") || strings.HasPrefix(r.URL.Path, "/admin/services"):
			adminContentHandler(w, r, db, logger)
		case r.URL.Path == "/experts":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", 405)
				return
			}
			expertsList(w, db, logger)
		case strings.HasPrefix(r.URL.Path, "/experts/"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", 405)
				return
			}
			expertView(w, r, db, strings.TrimPrefix(r.URL.Path, "/experts/"), logger)
		case r.URL.Path == "/admin/articles/new/save":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			saveArticle(w, r, db, logger)
		case r.URL.Path == "/admin/upload":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			uploadMedia(w, r, logger)
		case r.URL.Path == "/admin/media/upload":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			uploadMedia(w, r, logger)
		case strings.HasPrefix(r.URL.Path, "/articles/") && r.URL.Path != "/articles/kibernsk":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderArticle(w, r, db, logger)
		default:
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			renderPublicPage(w, r, db, logger)
		}
	}
}

// renderPublicPage отображает существующие публичные страницы сайта.
func renderPublicPage(w http.ResponseWriter, r *http.Request, db *sql.DB, logger *log.Logger) {
	data := publicPageData(db)
	name := ""
	status := http.StatusOK
	switch {
	case r.URL.Path == "/":
		name = "index.html"
	case r.URL.Path == "/about":
		name = "about.html"
	case r.URL.Path == "/privacy":
		name = "privacy.html"
	case r.URL.Path == "/expertise":
		name = "expertise.html"
		data.Articles = publishedArticles(db)
	case r.URL.Path == "/articles/kibernsk":
		name = "articles/kibernsk.html"
	case strings.HasPrefix(r.URL.Path, "/services/"):
		value := strings.TrimPrefix(r.URL.Path, "/services/")
		if value == "" {
			name, status = "404.html", http.StatusNotFound
			break
		}
		item, dbErr := serviceByID(db, value)
		if dbErr != nil || item.Status != "published" {
			name, status = "404.html", http.StatusNotFound
			break
		}
		renderServicePage(w, r, db, item, http.StatusOK, true, logger)
		return
	default:
		name, status = "404.html", http.StatusNotFound
	}
	renderTemplateStatus(w, name, data, status, logger)
}

// publishedArticles возвращает опубликованные статьи для карточек экспертизы.
func publishedArticles(db *sql.DB) []Article {
	rows, err := db.Query(`SELECT slug,title,publish_date,author,tags,content_md,cover_path FROM articles WHERE status='published' ORDER BY updated_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		var tagsJSON, coverPath string
		if err := rows.Scan(&article.Slug, &article.Title, &article.PublishDate, &article.Author, &tagsJSON, &article.ContentMD, &coverPath); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(tagsJSON), &article.Tags)
		article.CoverURL = mediaURL(coverPath)
		articles = append(articles, article)
	}
	return articles
}

// renderArticle отображает опубликованную статью по slug.
func renderArticle(w http.ResponseWriter, r *http.Request, db *sql.DB, logger *log.Logger) {
	slug := strings.TrimPrefix(r.URL.Path, "/articles/")
	var article Article
	var tagsJSON, coverPath string
	err := db.QueryRow(`SELECT slug,title,publish_date,author,tags,content_md,cover_path FROM articles WHERE slug=? AND status='published'`, slug).
		Scan(&article.Slug, &article.Title, &article.PublishDate, &article.Author, &tagsJSON, &article.ContentMD, &coverPath)
	if err != nil {
		renderTemplateStatus(w, "404.html", publicPageData(db), http.StatusNotFound, logger)
		return
	}
	_ = json.Unmarshal([]byte(tagsJSON), &article.Tags)
	article.CoverURL = mediaURL(coverPath)
	article.ContentHTML = markdownHTML(article.ContentMD)
	data := publicPageData(db)
	data.Article = article
	renderTemplateStatus(w, "articles/dynamic.html", data, http.StatusOK, logger)
}

// saveArticle сохраняет статью и её обложку из формы редактора.
func saveArticle(w http.ResponseWriter, r *http.Request, db *sql.DB, logger *log.Logger) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		logger.Printf("parse multipart form: %v", err)
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	date := strings.TrimSpace(r.FormValue("date"))
	author := strings.TrimSpace(r.FormValue("author"))
	contentMD := strings.TrimSpace(r.FormValue("content_md"))
	tagsRaw := strings.TrimSpace(r.FormValue("tags"))

	logger.Printf("saveArticle: title=%q date=%q author=%q tags=%q content_len=%d", title, date, author, tagsRaw, len(contentMD))

	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	slug := slugify(title)
	if slug == "" {
		http.Error(w, "title cannot create slug", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	status := "draft"
	if r.FormValue("status") == "published" {
		status = "published"
	}

	var tagsSlice []string
	if tagsRaw != "" {
		for _, part := range strings.Split(tagsRaw, ",") {
			if t := strings.TrimSpace(part); t != "" {
				tagsSlice = append(tagsSlice, t)
			}
		}
	}
	tagsJSON, _ := json.Marshal(tagsSlice)

	coverPath := ""
	if file, _, err := r.FormFile("cover"); err == nil {
		defer file.Close()
		coverPath, err = storeCover(file)
		if err != nil {
			logger.Printf("store cover: %v", err)
			http.Error(w, "invalid cover image", http.StatusBadRequest)
			return
		}
	}

	_, err := db.Exec(`INSERT INTO articles(slug,title,publish_date,author,tags,content_md,cover_path,status,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(slug) DO UPDATE SET title=excluded.title,publish_date=excluded.publish_date,author=excluded.author,tags=excluded.tags,content_md=excluded.content_md,cover_path=CASE WHEN excluded.cover_path='' THEN articles.cover_path ELSE excluded.cover_path END,status=excluded.status,updated_at=excluded.updated_at`,
		slug, title, date, author, string(tagsJSON), contentMD, coverPath, status, now, now)
	if err != nil {
		logger.Printf("save article DB: %v", err)
		http.Error(w, "could not save article", http.StatusInternalServerError)
		return
	}

	logger.Printf("article saved: slug=%s status=%s", slug, status)
	http.Redirect(w, r, "/expertise", http.StatusSeeOther)
}

// storeCover проверяет изображение и сохраняет его в data/uploads.
func storeCover(source io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(source, 8<<20))
	if err != nil {
		return "", err
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 {
		return "", fmt.Errorf("unsupported image")
	}
	ext := map[string]string{"jpeg": ".jpg", "png": ".png", "gif": ".gif"}[format]
	if ext == "" {
		return "", fmt.Errorf("unsupported image")
	}
	name := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	path := filepath.Join("data", "uploads", name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}

// markdownHTML преобразует Markdown в HTML.
func markdownHTML(source string) string {
	var output strings.Builder
	if err := goldmark.New(goldmark.WithExtensions(extension.GFM)).Convert([]byte(source), &output); err != nil {
		return ""
	}
	return output.String()
}

// mediaURL превращает имя файла обложки в URL сайта.
func mediaURL(path string) string {
	if path == "" {
		return ""
	}
	return "/media/" + filepath.Base(path)
}

// slugify создаёт URL-идентификатор статьи.
func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out []rune
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'а' && r <= 'я' || r >= '0' && r <= '9' {
			out = append(out, r)
		} else if len(out) > 0 && out[len(out)-1] != '-' {
			out = append(out, '-')
		}
	}
	return strings.Trim(string(out), "-")
}

// uploadMedia сохраняет изображение из редактора и возвращает URL для Markdown.
func uploadMedia(w http.ResponseWriter, r *http.Request, logger *log.Logger) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, "invalid upload", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	name, err := storeCover(file)
	if err != nil {
		logger.Printf("upload media: %v", err)
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"url":"/media/%s"}`, name)
}

// renderAdminForm отображает визуальный макет редактора статьи.
func renderAdminForm(w http.ResponseWriter, logger *log.Logger) {
	tpl, err := pongo2.FromFile(filepath.Join("templates", "admin", "form.html"))
	if err != nil {
		logger.Printf("template parse error admin/form.html: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	if err := tpl.ExecuteWriter(pongo2.Context{}, w); err != nil {
		logger.Printf("template execute error admin/form.html: %v", err)
	}
}

// servicePage находит услугу в БД (по id или slug) и отображает динамическую страницу.
// Опубликованные услуги доступны всем; черновик — только по адресу /services/preview/<id>.
func servicePage(w http.ResponseWriter, r *http.Request, db *sql.DB, logger *log.Logger) {
	value := strings.TrimPrefix(r.URL.Path, "/services/")
	if value == "" || value == "new" || strings.Contains(value, "/") {
		renderTemplateStatus(w, "404.html", publicPageData(db), http.StatusNotFound, logger)
		return
	}

	item, err := serviceByID(db, value)
	if err != nil {
		renderTemplateStatus(w, "404.html", publicPageData(db), http.StatusNotFound, logger)
		return
	}
	if item.Status != "published" {
		renderTemplateStatus(w, "404.html", publicPageData(db), http.StatusNotFound, logger)
		return
	}
	renderServicePage(w, r, db, item, http.StatusOK, true, logger)
}

// servicePreview показывает предпросмотр услуги из админки, включая черновики.
func servicePreview(w http.ResponseWriter, r *http.Request, db *sql.DB, id string, logger *log.Logger) {
	item, err := serviceByID(db, id)
	if err != nil {
		renderTemplateStatus(w, "404.html", publicPageData(db), http.StatusNotFound, logger)
		return
	}
	renderServicePage(w, r, db, item, http.StatusOK, false, logger)
}

// renderTemplate выполняет шаблон с успешным HTTP-статусом.
func renderTemplate(w http.ResponseWriter, name string, data PageData, logger *log.Logger) {
	renderTemplateStatus(w, name, data, http.StatusOK, logger)
}

// renderNamedTemplate выполняет один из дополнительных шаблонов сайта.
func renderNamedTemplate(w http.ResponseWriter, name string, data PageData, logger interface{ Printf(string, ...interface{}) }) {
	loader := pongo2.MustNewLocalFileSystemLoader(filepath.Join("templates"))
	set := pongo2.NewSet("site", loader)
	tpl, err := set.FromFile(name)
	if err != nil {
		logger.Printf("template parse error %s: %v", name, err)
		http.Error(w, "template error", 500)
		return
	}
	if err := tpl.ExecuteWriter(pongo2.Context{
		"services":              data.Services,
		"service":               data.Service,
		"db_service":            data.AdminService,
		"db_services":           data.DBServices,
		"service_children":      data.ServiceChildren,
		"articles":              data.Articles,
		"article":               data.Article,
		"staff":                 data.Staff,
		"staffs":                data.Staffs,
		"admin_service":         data.AdminService,
		"admin_services":        data.AdminServices,
		"services_by_direction": data.AdminServicesByDirection,
		"service_sections":      data.ServiceSections,
		"directions":            directions,
	}, w); err != nil {
		logger.Printf("template execute error %s: %v", name, err)
	}
}

// renderTemplateStatus загружает Pongo2-шаблон и его partial-файлы.
func renderTemplateStatus(w http.ResponseWriter, name string, data PageData, status int, logger *log.Logger) {
	loader := pongo2.MustNewLocalFileSystemLoader(filepath.Join("templates"))
	set := pongo2.NewSet("site", loader)
	tpl, err := set.FromFile(name)
	if err != nil {
		logger.Printf("template parse error %s: %v", name, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if err := tpl.ExecuteWriter(pongo2.Context{
		"services":              data.Services,
		"service":               data.Service,
		"db_service":            data.AdminService,
		"db_services":           data.DBServices,
		"service_children":      data.ServiceChildren,
		"articles":              data.Articles,
		"article":               data.Article,
		"staff":                 data.Staff,
		"staffs":                data.Staffs,
		"admin_service":         data.AdminService,
		"admin_services":        data.AdminServices,
		"services_by_direction": data.AdminServicesByDirection,
		"service_sections":      data.ServiceSections,
		"directions":            directions,
	}, w); err != nil {
		logger.Printf("template execute error %s: %v", name, err)
	}
}
