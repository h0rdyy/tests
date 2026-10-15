package handlers

import (
	"net/http"

	"cybersec/internal/middleware"
)

// Routes builds the public and API routes for the application.
func (a *App) Routes(staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/api/lead", a.handleLeadRoute)
	mux.HandleFunc("/", a.Pages)
	return middleware.LimitBody(mux)
}

func (a *App) handleLeadRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	a.Lead(w, r)
}
