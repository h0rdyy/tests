package main

import (
	"log"
	"net/http"
	"os"
)

// main запускает локальный HTTP-сервер сайта.
func main() {
	logger := log.New(os.Stdout, "cybersec: ", log.LstdFlags)
	db, err := openDatabase()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	if err := initContentTables(db); err != nil {
		logger.Fatal(err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: newRouter(db, logger),
	}

	logger.Println("server started at http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}
