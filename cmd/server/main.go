package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"cybersec/internal/config"
	"cybersec/internal/database"
	"cybersec/internal/handlers"
	"cybersec/internal/mailer"
	"cybersec/internal/middleware"
	templaterenderer "cybersec/internal/templates"
	"github.com/flosch/pongo2/v6"
	"gopkg.in/natefinch/lumberjack.v2"
	_ "modernc.org/sqlite"
)

func main() {
	root, _ := os.Getwd()
	cfg := config.Load()
	dbPath := cfg.DBPath
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(root, dbPath)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Initialize(db, root, dbPath); err != nil {
		log.Fatal(err)
	}

	set := pongo2.NewSet("site", pongo2.MustNewLocalFileSystemLoader(filepath.Join(root, "templates")))
	if _, err := set.FromFile("index.html"); err != nil {
		log.Fatal(err)
	}
	if !filepath.IsAbs(cfg.LogPath) {
		cfg.LogPath = filepath.Join(root, cfg.LogPath)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0755); err != nil {
		log.Fatal(err)
	}
	logger := log.New(&lumberjack.Logger{Filename: cfg.LogPath, MaxSize: 5, MaxBackups: 3, MaxAge: 30, Compress: true}, "", log.LstdFlags)
	app := &handlers.App{DB: db, Log: logger, Limiter: middleware.NewLimiter(), Render: templaterenderer.New(set, logger), Mailer: mailer.New(mailer.Config{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Password: cfg.SMTPPass, Receiver: cfg.Receiver}, logger)}
	server := &http.Server{Addr: "0.0.0.0:" + cfg.Port, Handler: app.Routes(filepath.Join(root, "static"))}
	logger.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}
